package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/neuradex-labs/blindenv/config"
	"github.com/neuradex-labs/blindenv/engine"
	"github.com/neuradex-labs/blindenv/provider"
	"github.com/neuradex-labs/blindenv/provider/cc"
)

func hookCmd() error {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: blindenv hook <platform> <hook-name>")
		fmt.Fprintln(os.Stderr, "  platforms: cc (Claude Code)")
		fmt.Fprintln(os.Stderr, "  cc hooks:  bash, read, grep, guard-file")
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
	case "read", "grep":
		result = hookFileAccess(p, stdin)
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
	filePath := p.ParseFilePath(stdin)
	if filePath == "" {
		return allow
	}

	cfg, err := config.Load()
	if err != nil || cfg == nil {
		return allow
	}

	secrets := engine.ResolveSecrets(cfg)
	if blocked, reason := engine.CheckFile(filePath, cfg, secrets); blocked {
		return provider.HookResult{Action: provider.Block, Reason: reason}
	}
	return allow
}

func hookGuardFile(p provider.Provider, stdin []byte) provider.HookResult {
	filePath := p.ParseFilePath(stdin)
	if filePath == "" {
		return allow
	}

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

	if engine.MatchSecretFilePath(filePath, cfg.SecretFiles) {
		return provider.HookResult{
			Action: provider.Block,
			Reason: "cannot modify secret file. Ask the user to edit it directly.",
		}
	}
	return allow
}
