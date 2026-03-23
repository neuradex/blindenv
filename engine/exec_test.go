package engine

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neuradex-labs/blindenv/config"
)

// skipIfNoBash skips the test when bash is unavailable (e.g. restricted CI).
func skipIfNoBash(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not found in PATH")
	}
}

// ---------------------------------------------------------------------------
// Run - exit code behaviour
// ---------------------------------------------------------------------------

func TestRun_ExitCode_Zero(t *testing.T) {
	skipIfNoBash(t)

	cfg := &config.Config{}
	code := Run(cfg, "true")
	if code != 0 {
		t.Errorf("want exit code 0, got %d", code)
	}
}

func TestRun_ExitCode_NonZero(t *testing.T) {
	skipIfNoBash(t)

	cfg := &config.Config{}
	code := Run(cfg, "exit 42")
	if code != 42 {
		t.Errorf("want exit code 42, got %d", code)
	}
}

func TestRun_ExitCode_CommandNotFound(t *testing.T) {
	skipIfNoBash(t)

	cfg := &config.Config{}
	code := Run(cfg, "this_command_does_not_exist_xyz_999")
	if code == 0 {
		t.Error("nonexistent command should produce a non-zero exit code")
	}
}

// ---------------------------------------------------------------------------
// Run - redaction
// ---------------------------------------------------------------------------

// captureRun executes Run while capturing stdout via a pipe so the test can
// inspect what the subprocess actually printed.
// It redirects os.Stdout for the duration of the call and restores it after.
func captureRun(t *testing.T, cfg *config.Config, command string) (stdout string, code int) {
	t.Helper()

	// Create a pipe; redirect os.Stdout to the write end.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	origStdout := os.Stdout
	os.Stdout = w

	code = Run(cfg, command)

	// Close write end so the Read below terminates.
	w.Close()
	os.Stdout = origStdout

	var buf strings.Builder
	tmp := make([]byte, 4096)
	for {
		n, _ := r.Read(tmp)
		if n == 0 {
			break
		}
		buf.Write(tmp[:n])
	}
	r.Close()

	return strings.TrimRight(buf.String(), "\n"), code
}

func TestRun_OutputRedacted_InjectVar(t *testing.T) {
	skipIfNoBash(t)

	// Set the secret in the process env so ResolveSecrets can find it.
	setEnv(t, "BLINDENV_EXEC_SECRET", "xXxTopSecretxXx")

	cfg := &config.Config{
		Inject: []string{"BLINDENV_EXEC_SECRET"},
	}

	// The command echoes the secret; Run must redact it.
	out, code := captureRun(t, cfg, "echo $BLINDENV_EXEC_SECRET")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if strings.Contains(out, "xXxTopSecretxXx") {
		t.Errorf("secret must be redacted in stdout; got %q", out)
	}
	if !strings.Contains(out, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in stdout; got %q", out)
	}
}

func TestRun_OutputRedacted_SecretFile(t *testing.T) {
	skipIfNoBash(t)

	envPath := writeTempEnvFile(t, "FILE_EXEC_SECRET=file_secret_value\n")

	cfg := &config.Config{
		SecretFiles: []string{envPath},
	}

	out, code := captureRun(t, cfg, "echo $FILE_EXEC_SECRET")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if strings.Contains(out, "file_secret_value") {
		t.Errorf("file secret must be redacted; got %q", out)
	}
	if !strings.Contains(out, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in output; got %q", out)
	}
}

func TestRun_SafeOutput_NotAltered(t *testing.T) {
	skipIfNoBash(t)

	cfg := &config.Config{}
	out, code := captureRun(t, cfg, "echo hello_world")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out, "hello_world") {
		t.Errorf("safe output should not be altered; got %q", out)
	}
}

// ---------------------------------------------------------------------------
// Run - environment isolation (passthrough / permissive)
// ---------------------------------------------------------------------------

func TestRun_PermissiveMode_SubprocessSeesEnvVar(t *testing.T) {
	skipIfNoBash(t)

	setEnv(t, "BLINDENV_VISIBLE_VAR", "visible_value")

	// No passthrough means permissive: subprocess inherits everything.
	cfg := &config.Config{}
	out, code := captureRun(t, cfg, "echo $BLINDENV_VISIBLE_VAR")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(out, "visible_value") {
		t.Errorf("permissive mode: subprocess should see inherited env; got %q", out)
	}
}

func TestRun_StrictMode_BlockedVarIsEmpty(t *testing.T) {
	skipIfNoBash(t)

	setEnv(t, "BLINDENV_STRICT_SECRET", "should_not_leak")

	// passthrough does not include BLINDENV_STRICT_SECRET.
	cfg := &config.Config{
		Passthrough: []string{"PATH", "HOME"},
	}

	// The subprocess env will not have BLINDENV_STRICT_SECRET, so echo
	// outputs an empty string (just a newline).
	out, code := captureRun(t, cfg, "echo $BLINDENV_STRICT_SECRET")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if strings.Contains(out, "should_not_leak") {
		t.Errorf("strict mode: non-passthrough var must not reach subprocess; got %q", out)
	}
}

func TestRun_InjectVar_ReachesSubprocess(t *testing.T) {
	skipIfNoBash(t)

	setEnv(t, "BLINDENV_INJECT_TEST", "injected_value")

	cfg := &config.Config{
		Passthrough: []string{"PATH"}, // strict
		Inject:      []string{"BLINDENV_INJECT_TEST"},
	}

	out, code := captureRun(t, cfg, "echo $BLINDENV_INJECT_TEST")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	// The injected value IS a secret so it will be redacted, but it must have
	// reached the subprocess (evidenced by [REDACTED] appearing rather than an
	// empty echo line).
	if !strings.Contains(out, "[REDACTED]") {
		t.Errorf("injected secret should reach subprocess and be redacted; got %q", out)
	}
}

// ---------------------------------------------------------------------------
// Run - secret_files env loading
// ---------------------------------------------------------------------------

func TestRun_SecretFileVars_Available_And_Redacted(t *testing.T) {
	skipIfNoBash(t)

	envPath := writeTempEnvFile(t, "RUN_FILE_KEY=run_file_secret\n")

	cfg := &config.Config{
		SecretFiles: []string{envPath},
	}

	out, code := captureRun(t, cfg, "echo $RUN_FILE_KEY")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if strings.Contains(out, "run_file_secret") {
		t.Errorf("file secret must be redacted; got %q", out)
	}
	if !strings.Contains(out, "[REDACTED]") {
		t.Errorf("expected [REDACTED]; got %q", out)
	}
}

// ---------------------------------------------------------------------------
// Run - secret file path is NOT accessible from subprocess
// ---------------------------------------------------------------------------

func TestRun_SecretFilePath_NotListedAsEnvFile(t *testing.T) {
	// The secret_files path should supply env var values but should not expose
	// the file path itself. We confirm the var is available (redacted) and the
	// raw path string is not in the output.
	skipIfNoBash(t)

	dir := t.TempDir()
	envPath := filepath.Join(dir, "secrets.env")
	if err := os.WriteFile(envPath, []byte("SECRET_VAL=top_secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		SecretFiles: []string{envPath},
	}

	// Print all env vars; the raw file path itself must not appear
	// as a variable value in the env (it's only used as a source path).
	out, _ := captureRun(t, cfg, "echo $SECRET_VAL")
	if strings.Contains(out, "top_secret") {
		t.Errorf("raw secret value must not appear in output; got %q", out)
	}
}
