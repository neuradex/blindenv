package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/neuradex/blindenv/internal/config"
	"github.com/neuradex/blindenv/internal/engine"
	"github.com/neuradex/blindenv/internal/provider"
	"github.com/neuradex/blindenv/internal/provider/cc"
)

func hookCmd() error {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: blindenv hook <platform> <hook-name>")
		fmt.Fprintln(os.Stderr, "  platforms: cc (Claude Code)")
		fmt.Fprintln(os.Stderr, "  cc hooks:  bash, read, grep, glob, guard-file")
		os.Exit(1)
	}

	platform := os.Args[2]
	hookName := os.Args[3]

	p := resolveProvider(platform)
	if p == nil {
		fmt.Fprintf(os.Stderr, "unknown platform: %s\n", platform)
		os.Exit(1)
	}

	stdin, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(0)
	}

	var result provider.HookResult
	switch hookName {
	case "bash":
		result = hookBash(p, stdin)
	case "read":
		result = hookFileAccess(p, stdin)
	case "grep":
		result = hookGrep(p, stdin)
	case "glob":
		result = hookGlob(p, stdin)
	case "guard-file":
		result = hookGuardFile(p, stdin)
	default:
		fmt.Fprintf(os.Stderr, "unknown hook: %s\n", hookName)
		os.Exit(1)
	}

	respond(p, result)
	return nil // unreachable — respond always calls os.Exit
}

var allow = provider.HookResult{Action: provider.Allow}

func resolveProvider(name string) provider.Provider {
	switch name {
	case "cc":
		return cc.New()
	default:
		return nil
	}
}

func respond(p provider.Provider, result provider.HookResult) {
	switch result.Action {
	case provider.Allow:
		if data := p.FormatAllow(); data != nil {
			fmt.Println(string(data))
		}
		os.Exit(0)
	case provider.Block:
		stderr, exitCode := p.FormatBlock(result.Reason)
		fmt.Fprintln(os.Stderr, stderr)
		os.Exit(exitCode)
	case provider.Rewrite:
		if data := p.FormatRewrite(result.Command); data != nil {
			fmt.Println(string(data))
		}
		os.Exit(0)
	case provider.Modify:
		if data := p.FormatModifiedInput(result.UpdatedInput); data != nil {
			fmt.Println(string(data))
		}
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "blindenv: unknown action %d\n", result.Action)
		os.Exit(1)
	}
}

// --- Hook logic (platform-independent) ---

func hookBash(p provider.Provider, stdin []byte) provider.HookResult {
	command := p.ParseBashCommand(stdin)
	if command == "" || strings.Contains(command, "blindenv run ") || strings.HasSuffix(command, "blindenv") {
		return allow
	}

	cfg, err := config.Load()
	if err != nil || cfg == nil || !cfg.HasSecrets() {
		return allow
	}

	binPath := "blindenv"
	if root := os.Getenv("CLAUDE_PLUGIN_ROOT"); root != "" {
		binPath = filepath.Join(root, "bin", "blindenv")
	}

	escaped := strings.ReplaceAll(command, "'", "'\\''")
	return provider.HookResult{
		Action:  provider.Rewrite,
		Command: fmt.Sprintf("%s run '%s'", binPath, escaped),
	}
}

func hookFileAccess(p provider.Provider, stdin []byte) provider.HookResult {
	toolInput := p.ParseToolInput(stdin)
	if toolInput == nil {
		return allow
	}

	filePath, _ := toolInput["file_path"].(string)
	if filePath == "" {
		return allow
	}

	cfg, err := config.Load()
	if err != nil || cfg == nil || !cfg.HasSecrets() {
		return allow
	}

	absPath, _ := filepath.Abs(filePath)

	// Fast path: check path match and cache dir without resolving secrets.
	if engine.MatchSecretFilePath(absPath, cfg.SecretFiles) || engine.IsInsideCacheDir(absPath) {
		toolInput["file_path"] = "/dev/null/.blindenv-nonexistent"
		return provider.HookResult{Action: provider.Modify, UpdatedInput: toolInput}
	}

	// Slow path: content scan — only resolve secrets when path checks pass.
	secrets := engine.ResolveSecrets(cfg)
	if blocked, _ := engine.CheckFileForSecrets(absPath, secrets); blocked {
		toolInput["file_path"] = "/dev/null/.blindenv-nonexistent"
		return provider.HookResult{Action: provider.Modify, UpdatedInput: toolInput}
	}
	return allow
}

func hookGrep(p provider.Provider, stdin []byte) provider.HookResult {
	toolInput := p.ParseToolInput(stdin)
	if toolInput == nil {
		return allow
	}

	cfg, err := config.Load()
	if err != nil || cfg == nil || !cfg.HasSecrets() {
		return allow
	}

	// Block if path targets a secret file/dir.
	searchPath, _ := toolInput["path"].(string)
	if searchPath != "" {
		secrets := engine.ResolveSecrets(cfg)
		if blocked, reason := engine.CheckFile(searchPath, cfg, secrets); blocked {
			return provider.HookResult{Action: provider.Block, Reason: reason}
		}
	}

	// Inject exclusion globs so secret files are silently omitted from results.
	excludes := buildExcludeGlobs(cfg.SecretFiles)
	if excludes == "" {
		return allow
	}

	currentGlob, _ := toolInput["glob"].(string)
	if currentGlob != "" {
		toolInput["glob"] = currentGlob + "," + excludes
	} else {
		toolInput["glob"] = excludes
	}

	return provider.HookResult{Action: provider.Modify, UpdatedInput: toolInput}
}

func hookGlob(p provider.Provider, stdin []byte) provider.HookResult {
	toolInput := p.ParseToolInput(stdin)
	if toolInput == nil {
		return allow
	}

	cfg, err := config.Load()
	if err != nil || cfg == nil || !cfg.HasSecrets() {
		return allow
	}

	// Block if path targets a secret directory.
	searchPath, _ := toolInput["path"].(string)
	if searchPath != "" && engine.MatchSecretFilePath(searchPath, cfg.SecretFiles) {
		return provider.HookResult{
			Action: provider.Block,
			Reason: "cannot list files in secret directory",
		}
	}

	// Inject negation patterns to hide secret files from results.
	excludes := buildExcludeGlobs(cfg.SecretFiles)
	if excludes == "" {
		return allow
	}

	currentPattern, _ := toolInput["pattern"].(string)
	if currentPattern == "" {
		return allow
	}
	toolInput["pattern"] = currentPattern + "," + excludes
	return provider.HookResult{Action: provider.Modify, UpdatedInput: toolInput}
}

// buildExcludeGlobs returns ripgrep-style negation globs for secret files
// and the cache directory.
func buildExcludeGlobs(secretFiles []string) string {
	var parts []string
	for _, sf := range secretFiles {
		parts = append(parts, "!"+filepath.Base(sf))
	}
	// Also exclude the cache directory.
	parts = append(parts, "!.cache/blindenv/**")
	return strings.Join(parts, ",")
}

func hookGuardFile(p provider.Provider, stdin []byte) provider.HookResult {
	toolInput := p.ParseToolInput(stdin)
	if toolInput == nil {
		return allow
	}

	filePath, _ := toolInput["file_path"].(string)
	if filePath == "" {
		return allow
	}

	// Config protection — agent already knows blindenv exists, so block explicitly.
	base := filepath.Base(filePath)
	if base == config.ConfigFileName || base == config.GlobalConfigFileName {
		return provider.HookResult{
			Action: provider.Block,
			Reason: "cannot modify blindenv config. Ask the user to edit it directly.",
		}
	}

	cfg, err := config.Load()
	if err != nil || cfg == nil || !cfg.HasSecrets() {
		return allow
	}

	absPath, _ := filepath.Abs(filePath)

	// Secret files or cache dir — pretend the file doesn't exist.
	if engine.MatchSecretFilePath(absPath, cfg.SecretFiles) || engine.IsInsideCacheDir(absPath) {
		toolInput["file_path"] = "/dev/null/.blindenv-nonexistent"
		return provider.HookResult{Action: provider.Modify, UpdatedInput: toolInput}
	}
	return allow
}
