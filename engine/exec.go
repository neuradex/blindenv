package engine

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/neuradex/blindenv/config"
)

// Run executes a shell command with secret isolation and output redaction.
func Run(cfg *config.Config, command string) int {
	secrets := ResolveSecrets(cfg)
	env := BuildSanitizedEnv(cfg, secrets)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}
	cmd.Env = env
	cmd.Stdin = os.Stdin

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Redact and output stdout
	if out := strings.TrimRight(stdout.String(), "\n"); out != "" {
		fmt.Println(RedactSecrets(out, secrets))
	}

	// Redact and output stderr
	if errOut := strings.TrimRight(stderr.String(), "\n"); errOut != "" {
		fmt.Fprintln(os.Stderr, RedactSecrets(errOut, secrets))
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintln(os.Stderr, RedactSecrets(err.Error(), secrets))
		return 1
	}

	return 0
}
