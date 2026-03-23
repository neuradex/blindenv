package cc

import (
	"encoding/json"
	"fmt"
)

type hookInput struct {
	ToolInput map[string]interface{} `json:"tool_input"`
}

// Provider implements the Claude Code hook protocol.
type Provider struct{}

func New() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "cc" }

func (p *Provider) ParseBashCommand(stdin []byte) string {
	return parseField(stdin, "command")
}

func (p *Provider) ParseFilePath(stdin []byte) string {
	return parseField(stdin, "file_path")
}

func (p *Provider) FormatAllow() []byte {
	return nil // exit 0, no stdout
}

func (p *Provider) FormatBlock(reason string) (string, int) {
	return fmt.Sprintf("blindenv: %s", reason), 2
}

func (p *Provider) FormatRewrite(newCommand string) []byte {
	out := map[string]interface{}{
		"hookSpecificOutput": map[string]interface{}{
			"hookEventName":      "PreToolUse",
			"permissionDecision": "allow",
			"updatedInput": map[string]interface{}{
				"command": newCommand,
			},
		},
	}
	data, _ := json.Marshal(out)
	return data
}

func (p *Provider) ParseToolInput(stdin []byte) map[string]interface{} {
	var input hookInput
	if err := json.Unmarshal(stdin, &input); err != nil {
		return nil
	}
	return input.ToolInput
}

func (p *Provider) FormatModifiedInput(input map[string]interface{}) []byte {
	out := map[string]interface{}{
		"hookSpecificOutput": map[string]interface{}{
			"hookEventName":      "PreToolUse",
			"permissionDecision": "allow",
			"updatedInput":       input,
		},
	}
	data, _ := json.Marshal(out)
	return data
}

func parseField(stdin []byte, field string) string {
	var input hookInput
	if err := json.Unmarshal(stdin, &input); err != nil {
		return ""
	}
	if input.ToolInput == nil {
		return ""
	}
	val, ok := input.ToolInput[field]
	if !ok {
		return ""
	}
	s, ok := val.(string)
	if !ok {
		return ""
	}
	return s
}
