package cmd

import (
	"fmt"
	"os"

	"github.com/neuradex/blindenv/config"
	"github.com/neuradex/blindenv/engine"
)

const version = "0.1.0"

func Execute() error {
	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	switch os.Args[1] {
	case "run":
		return runCmd()
	case "check-file":
		return checkFileCmd()
	case "has-config":
		return hasConfigCmd()
	case "init":
		return initCmd()
	case "hook":
		return hookCmd()
	case "cache-restore":
		return cacheRestoreCmd()
	case "cache-refresh":
		return cacheRefreshCmd()
	case "version", "--version", "-v":
		fmt.Println("blindenv " + version)
		return nil
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
		return nil
	}
}

func runCmd() error {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: blindenv run '<command>'")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}
	if cfg == nil {
		// No config found - run without mediation
		cfg = &config.Config{}
	}

	command := os.Args[2]
	exitCode := engine.Run(cfg, command)
	os.Exit(exitCode)
	return nil
}

func checkFileCmd() error {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: blindenv check-file <path>")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}
	if cfg == nil {
		// No config - file is allowed
		return nil
	}

	secrets := engine.ResolveSecrets(cfg)
	blocked, reason := engine.CheckFile(os.Args[2], cfg, secrets)
	if blocked {
		fmt.Fprintln(os.Stderr, "blocked: "+reason)
		os.Exit(2)
	}
	return nil
}

func hasConfigCmd() error {
	cfg, err := config.Load()
	if err != nil {
		os.Exit(1)
	}
	if cfg == nil {
		os.Exit(1)
	}
	if !cfg.HasSecrets() {
		os.Exit(1)
	}
	return nil
}

func initCmd() error {
	// If config already exists anywhere up the tree, do nothing.
	if path := config.FindConfigFile(""); path != "" {
		return nil
	}

	path, err := config.CreateDefault()
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}
	fmt.Printf("created %s\n", path)
	return nil
}

func cacheRestoreCmd() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}
	if cfg == nil || len(cfg.SecretFiles) == 0 {
		fmt.Fprintln(os.Stderr, "no secret_files configured")
		os.Exit(1)
	}

	restored, skipped := engine.CacheRestore(cfg)
	for _, r := range restored {
		fmt.Printf("restored: %s\n", r)
	}
	for _, s := range skipped {
		fmt.Printf("skipped (no cache): %s\n", s)
	}
	if len(restored) == 0 {
		fmt.Println("nothing to restore")
	}
	return nil
}

func cacheRefreshCmd() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}
	if cfg == nil || len(cfg.SecretFiles) == 0 {
		fmt.Fprintln(os.Stderr, "no secret_files configured")
		os.Exit(1)
	}

	refreshed, skipped := engine.CacheRefresh(cfg)
	for _, r := range refreshed {
		fmt.Printf("refreshed: %s\n", r)
	}
	for _, s := range skipped {
		fmt.Printf("skipped (file missing): %s\n", s)
	}
	if len(refreshed) == 0 {
		fmt.Println("nothing to refresh")
	}
	return nil
}

func printUsage() {
	fmt.Print(`blindenv - Secret isolation for AI coding agents

Usage:
  blindenv init                   Create blindenv.yml in current directory (if not found)
  blindenv run '<command>'       Execute command with secret isolation + output redaction
  blindenv check-file <path>     Check if a file contains or exposes secrets
  blindenv has-config            Exit 0 if env mediation config exists, 1 otherwise
  blindenv cache-restore         Restore secret files from cache (after agent damage)
  blindenv cache-refresh         Re-cache secret files (after you edit .env)
  blindenv version               Show version
  blindenv help                  Show this help

Config:
  Place blindenv.yml in your project root or ~/.blindenv.yml

  inject:              # env vars from process env - injected + redacted
    - API_KEY
  passthrough:         # non-secret vars - explicit allowlist
    - PATH
    - HOME
  secret_files:        # .env files - auto-parsed, paths blocked
    - .env

Example:
  blindenv run 'curl -H "Authorization: $API_KEY" https://api.example.com'
  # Agent sees: {"result": "ok", "token": "[REDACTED]"}

`)
}
