package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/neuradex-labs/blindenv/config"
	"github.com/neuradex-labs/blindenv/engine"
)

// hookInput represents the JSON input from agent tool hooks.
type hookInput struct {
	ToolInput map[string]interface{} `json:"tool_input"`
}

func hookCmd() error {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: blindenv hook <platform> <hook-name>")
		fmt.Fprintln(os.Stderr, "  platforms: cc (Claude Code), oc (OpenClaw)")
		fmt.Fprintln(os.Stderr, "  cc hooks:  bash, read, grep, guard-config")
		os.Exit(1)
	}

	platform := os.Args[2]
	hookName := os.Args[3]

	switch platform {
	case "cc":
		return hookCC(hookName)
	default:
		fmt.Fprintf(os.Stderr, "unknown platform: %s\n", platform)
		os.Exit(1)
	}
	return nil
}

// hookCC dispatches Claude Code PreToolUse hooks.
func hookCC(hookName string) error {
	input, err := readHookInput()
	if err != nil {
		os.Exit(0)
	}

	switch hookName {
	case "bash":
		return ccBash(input)
	case "read":
		return ccRead(input)
	case "grep":
		return ccGrep(input)
	case "guard-config":
		return ccGuardConfig(input)
	default:
		fmt.Fprintf(os.Stderr, "unknown cc hook: %s\n", hookName)
		os.Exit(1)
	}
	return nil
}

func readHookInput() (*hookInput, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}
	var input hookInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, err
	}
	return &input, nil
}

func getToolInputString(input *hookInput, key string) string {
	if input.ToolInput == nil {
		return ""
	}
	val, ok := input.ToolInput[key]
	if !ok {
		return ""
	}
	s, ok := val.(string)
	if !ok {
		return ""
	}
	return s
}

// ccBash rewrites bare bash commands to go through `blindenv run`.
func ccBash(input *hookInput) error {
	command := getToolInputString(input, "command")
	if command == "" {
		os.Exit(0)
	}

	// Allow blindenv's own commands through
	if strings.HasPrefix(command, "blindenv ") || command == "blindenv" {
		os.Exit(0)
	}

	// Check if env mediation is active
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		os.Exit(0)
		return nil
	}
	if len(cfg.Inject) == 0 && len(cfg.SecretFiles) == 0 {
		os.Exit(0)
		return nil
	}

	// Rewrite command to go through blindenv run
	escaped := strings.ReplaceAll(command, "'", "'\\''")
	wrapped := fmt.Sprintf("blindenv run '%s'", escaped)

	out := map[string]interface{}{
		"hookSpecificOutput": map[string]interface{}{
			"hookEventName":      "PreToolUse",
			"permissionDecision": "allow",
			"updatedInput": map[string]interface{}{
				"command": wrapped,
			},
		},
	}
	data, _ := json.Marshal(out)
	fmt.Println(string(data))
	return nil
}

// ccRead blocks access to secret files.
func ccRead(input *hookInput) error {
	filePath := getToolInputString(input, "file_path")
	if filePath == "" {
		os.Exit(0)
	}

	cfg, err := config.Load()
	if err != nil || cfg == nil {
		os.Exit(0)
		return nil
	}

	secrets := engine.ResolveSecrets(cfg)
	blocked, reason := engine.CheckFile(filePath, cfg, secrets)
	if blocked {
		fmt.Fprintf(os.Stderr, "blindenv: %s\n", reason)
		os.Exit(2)
	}

	os.Exit(0)
	return nil
}

// ccGrep blocks grep on secret file paths.
func ccGrep(input *hookInput) error {
	searchPath := getToolInputString(input, "path")
	if searchPath == "" {
		os.Exit(0)
	}

	cfg, err := config.Load()
	if err != nil || cfg == nil {
		os.Exit(0)
		return nil
	}

	secrets := engine.ResolveSecrets(cfg)
	blocked, reason := engine.CheckFile(searchPath, cfg, secrets)
	if blocked {
		fmt.Fprintf(os.Stderr, "blindenv: %s\n", reason)
		os.Exit(2)
	}

	os.Exit(0)
	return nil
}

// ccGuardConfig blocks agent from modifying blindenv config files.
func ccGuardConfig(input *hookInput) error {
	filePath := getToolInputString(input, "file_path")
	if filePath == "" {
		os.Exit(0)
	}

	base := filepath.Base(filePath)
	if base == "blindenv.yml" || base == ".blindenv.yml" {
		fmt.Fprintln(os.Stderr, "blindenv: cannot modify blindenv config. Ask the user to edit it directly.")
		os.Exit(2)
	}

	os.Exit(0)
	return nil
}
