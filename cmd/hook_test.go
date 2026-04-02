package cmd

import (
	"testing"
)

func TestIsAlreadyWrapped(t *testing.T) {
	tests := []struct {
		name    string
		command string
		binPath string
		want    bool
	}{
		// --- true: genuinely wrapped ---
		{
			name:    "blindenv run with single-quoted command",
			command: "blindenv run 'ls'",
			binPath: "blindenv",
			want:    true,
		},
		{
			name:    "full path blindenv run",
			command: "/path/to/blindenv run 'ls'",
			binPath: "/path/to/blindenv",
			want:    true,
		},
		{
			name:    "blindenv run with complex command",
			command: "blindenv run 'echo hello && ls -la'",
			binPath: "blindenv",
			want:    true,
		},
		{
			name:    "bare blindenv command (no subcommand)",
			command: "blindenv",
			binPath: "blindenv",
			want:    true,
		},
		{
			name:    "bare full-path blindenv (no subcommand)",
			command: "/usr/local/bin/blindenv",
			binPath: "/usr/local/bin/blindenv",
			want:    true,
		},
		{
			name:    "blindenv run with leading whitespace",
			command: "  blindenv run 'ls'",
			binPath: "blindenv",
			want:    true,
		},
		{
			name:    "blindenv init (not run, but still blindenv prefix)",
			command: "blindenv init",
			binPath: "blindenv",
			want:    true,
		},
		{
			name:    "full path with trailing args",
			command: "/opt/bin/blindenv hook cc bash",
			binPath: "/opt/bin/blindenv",
			want:    true,
		},

		// --- false: must NOT be treated as wrapped (security-critical) ---
		{
			name:    "echo with blindenv run in the middle",
			command: "echo 'blindenv run ' && printenv",
			binPath: "blindenv",
			want:    false,
		},
		{
			name:    "simple ls command",
			command: "ls -la",
			binPath: "blindenv",
			want:    false,
		},
		{
			name:    "variable assignment referencing blindenv",
			command: "CMD='blindenv run' && $CMD",
			binPath: "blindenv",
			want:    false,
		},
		{
			name:    "cat command mentioning blindenv in argument",
			command: "cat /path/to/blindenv",
			binPath: "blindenv",
			want:    false,
		},
		{
			name:    "grep for blindenv in a file",
			command: "grep blindenv /etc/hosts",
			binPath: "blindenv",
			want:    false,
		},
		{
			name:    "semicolon before blindenv run",
			command: "ls; blindenv run 'echo hi'",
			binPath: "blindenv",
			want:    false,
		},
		{
			name:    "pipe ending with blindenv",
			command: "echo hi | blindenv run 'cat'",
			binPath: "blindenv",
			want:    false,
		},
		{
			name:    "empty command",
			command: "",
			binPath: "blindenv",
			want:    false,
		},
		{
			name:    "whitespace only command",
			command: "   ",
			binPath: "blindenv",
			want:    false,
		},
		{
			name:    "full path does not match if binPath differs",
			command: "/other/path/blindenv run 'ls'",
			binPath: "/usr/local/bin/blindenv",
			want:    false,
		},
		{
			name:    "blindenv in env var assignment",
			command: "BLINDENV_PATH=/usr/local/bin/blindenv",
			binPath: "blindenv",
			want:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isAlreadyWrapped(tc.command, tc.binPath)
			if got != tc.want {
				t.Errorf("isAlreadyWrapped(%q, %q) = %v, want %v",
					tc.command, tc.binPath, got, tc.want)
			}
		})
	}
}
