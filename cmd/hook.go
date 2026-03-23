package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/neuradex/blindenv/config"
	"github.com/neuradex/blindenv/engine"
	"github.com/neuradex/blindenv/provider"
	"github.com/neuradex/blindenv/provider/cc"
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

	return respond(p, result)
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

func respond(p provider.Provider, result provider.HookResult) error {
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
	}
	return nil
}

// --- Hook logic (platform-independent) ---

func hookBash(p provider.Provider, stdin []byte) provider.HookResult {
	command := p.ParseBashCommand(stdin)
	if command == "" || command == "blindenv" || strings.HasPrefix(command, "blindenv ") {
		return allow
	}

	cfg, err := config.Load()
	if err != nil || cfg == nil || !cfg.HasSecrets() {
		return allow
	}

	escaped := strings.ReplaceAll(command, "'", "'\\''")
	return provider.HookResult{
		Action:  provider.Rewrite,
		Command: fmt.Sprintf("blindenv run '%s'", escaped),
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
	if err != nil || cfg == nil {
		return allow
	}

	secrets := engine.ResolveSecrets(cfg)
	if blocked, _ := engine.CheckFile(filePath, cfg, secrets); blocked {
		// Redirect to a non-existent path so the Read tool returns
		// "file does not exist" naturally — no blindenv fingerprint.
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
	if currentPattern != "" {
		toolInput["pattern"] = currentPattern + "," + excludes
	}

	return provider.HookResult{Action: provider.Modify, UpdatedInput: toolInput}
}

// buildExcludeGlobs returns ripgrep-style negation globs for secret files.
func buildExcludeGlobs(secretFiles []string) string {
	var parts []string
	for _, sf := range secretFiles {
		expanded := filepath.Base(sf)
		parts = append(parts, "!"+expanded)
	}
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
	if base == "blindenv.yml" || base == ".blindenv.yml" {
		return provider.HookResult{
			Action: provider.Block,
			Reason: "cannot modify blindenv config. Ask the user to edit it directly.",
		}
	}

	cfg, err := config.Load()
	if err != nil || cfg == nil {
		return allow
	}

	// Secret files — pretend the file doesn't exist.
	if engine.MatchSecretFilePath(filePath, cfg.SecretFiles) {
		toolInput["file_path"] = "/dev/null/.blindenv-nonexistent"
		return provider.HookResult{Action: provider.Modify, UpdatedInput: toolInput}
	}
	return allow
}
