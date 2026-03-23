package provider

// HookResult represents what a hook should do.
type HookResult struct {
	Action  Action
	Reason  string // for Block
	Command string // for Rewrite
}

type Action int

const (
	Allow   Action = iota // exit 0, no output
	Block                 // exit 2 + stderr reason
	Rewrite               // exit 0 + rewritten command
)

// Provider adapts blindenv to a specific agent platform.
type Provider interface {
	// Name returns the provider identifier (e.g. "cc", "oc").
	Name() string

	// ParseBashCommand extracts the shell command from hook stdin.
	ParseBashCommand(stdin []byte) string

	// ParseFilePath extracts the file path from hook stdin.
	ParseFilePath(stdin []byte) string

	// FormatAllow returns stdout bytes for an allow response.
	FormatAllow() []byte

	// FormatBlock returns stderr string and exit code for a block response.
	FormatBlock(reason string) (stderr string, exitCode int)

	// FormatRewrite returns stdout bytes that rewrite the bash command.
	FormatRewrite(newCommand string) []byte
}
