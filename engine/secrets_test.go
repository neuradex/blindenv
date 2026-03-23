package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neuradex-labs/blindenv/config"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// writeTempEnvFile creates a temporary .env file with the given content and
// returns its absolute path. The file is removed when the test ends.
func writeTempEnvFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writeTempEnvFile: %v", err)
	}
	return path
}

// setEnv sets an environment variable for the duration of the test and
// restores the original value (or unsets it) when the test ends.
func setEnv(t *testing.T, key, value string) {
	t.Helper()
	orig, exists := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("setEnv: %v", err)
	}
	t.Cleanup(func() {
		if exists {
			os.Setenv(key, orig)
		} else {
			os.Unsetenv(key)
		}
	})
}

// ---------------------------------------------------------------------------
// parseEnvFile
// ---------------------------------------------------------------------------

func TestParseEnvFile_BasicKeyValue(t *testing.T) {
	path := writeTempEnvFile(t, "API_KEY=abc123\nDB_PASS=secret\n")
	got := parseEnvFile(path)

	if got["API_KEY"] != "abc123" {
		t.Errorf("API_KEY: want abc123, got %q", got["API_KEY"])
	}
	if got["DB_PASS"] != "secret" {
		t.Errorf("DB_PASS: want secret, got %q", got["DB_PASS"])
	}
}

func TestParseEnvFile_ExportPrefix(t *testing.T) {
	path := writeTempEnvFile(t, "export TOKEN=tok_xyz\n")
	got := parseEnvFile(path)

	if got["TOKEN"] != "tok_xyz" {
		t.Errorf("TOKEN: want tok_xyz, got %q", got["TOKEN"])
	}
}

func TestParseEnvFile_DoubleQuotes(t *testing.T) {
	path := writeTempEnvFile(t, `KEY="quoted value"` + "\n")
	got := parseEnvFile(path)

	if got["KEY"] != "quoted value" {
		t.Errorf("KEY: want \"quoted value\", got %q", got["KEY"])
	}
}

func TestParseEnvFile_SingleQuotes(t *testing.T) {
	path := writeTempEnvFile(t, "KEY='single quoted'\n")
	got := parseEnvFile(path)

	if got["KEY"] != "single quoted" {
		t.Errorf("KEY: want \"single quoted\", got %q", got["KEY"])
	}
}

func TestParseEnvFile_CommentsAndBlankLines(t *testing.T) {
	content := `
# This is a comment
VALID=yes

# Another comment
`
	path := writeTempEnvFile(t, content)
	got := parseEnvFile(path)

	if got["VALID"] != "yes" {
		t.Errorf("VALID: want yes, got %q", got["VALID"])
	}
	// Only VALID should be present; comments and blank lines are ignored.
	if len(got) != 1 {
		t.Errorf("expected 1 key, got %d: %v", len(got), got)
	}
}

func TestParseEnvFile_EmptyValue_IsSkipped(t *testing.T) {
	// parseEnvFile only records non-empty values.
	path := writeTempEnvFile(t, "EMPTY=\n")
	got := parseEnvFile(path)

	if _, ok := got["EMPTY"]; ok {
		t.Error("expected EMPTY to be absent (empty values are skipped)")
	}
}

func TestParseEnvFile_InvalidLines_AreIgnored(t *testing.T) {
	content := "not-a-valid-line\n123STARTS_WITH_DIGIT=bad\nGOOD=value\n"
	path := writeTempEnvFile(t, content)
	got := parseEnvFile(path)

	if got["GOOD"] != "value" {
		t.Errorf("GOOD: want value, got %q", got["GOOD"])
	}
	if len(got) != 1 {
		t.Errorf("expected 1 valid key, got %d: %v", len(got), got)
	}
}

func TestParseEnvFile_NonExistentFile(t *testing.T) {
	got := parseEnvFile("/nonexistent/path/.env")
	if len(got) != 0 {
		t.Errorf("expected empty map for missing file, got %v", got)
	}
}

func TestParseEnvFile_MultipleEntries(t *testing.T) {
	content := "A=1\nB=2\nC=3\n"
	path := writeTempEnvFile(t, content)
	got := parseEnvFile(path)

	for k, want := range map[string]string{"A": "1", "B": "2", "C": "3"} {
		if got[k] != want {
			t.Errorf("%s: want %q, got %q", k, want, got[k])
		}
	}
}

// ---------------------------------------------------------------------------
// ResolveSecrets
// ---------------------------------------------------------------------------

func TestResolveSecrets_InjectFromProcessEnv(t *testing.T) {
	setEnv(t, "BLINDENV_TEST_KEY", "process_secret")

	cfg := &config.Config{
		Inject: []string{"BLINDENV_TEST_KEY"},
	}
	got := ResolveSecrets(cfg)

	if got["BLINDENV_TEST_KEY"] != "process_secret" {
		t.Errorf("want process_secret, got %q", got["BLINDENV_TEST_KEY"])
	}
}

func TestResolveSecrets_InjectMissingVarIsSkipped(t *testing.T) {
	// Ensure the variable is not set.
	os.Unsetenv("BLINDENV_NONEXISTENT_9999")

	cfg := &config.Config{
		Inject: []string{"BLINDENV_NONEXISTENT_9999"},
	}
	got := ResolveSecrets(cfg)

	if _, ok := got["BLINDENV_NONEXISTENT_9999"]; ok {
		t.Error("expected missing env var to be absent from secrets")
	}
}

func TestResolveSecrets_SecretFilesAreParsed(t *testing.T) {
	path := writeTempEnvFile(t, "FILE_SECRET=from_file\n")

	cfg := &config.Config{
		SecretFiles: []string{path},
	}
	got := ResolveSecrets(cfg)

	if got["FILE_SECRET"] != "from_file" {
		t.Errorf("want from_file, got %q", got["FILE_SECRET"])
	}
}

func TestResolveSecrets_InjectTakesPriorityOverSecretFile(t *testing.T) {
	// Both inject and secret_files define the same key; inject wins.
	setEnv(t, "BLINDENV_PRIO_KEY", "from_inject")
	path := writeTempEnvFile(t, "BLINDENV_PRIO_KEY=from_file\n")

	cfg := &config.Config{
		Inject:      []string{"BLINDENV_PRIO_KEY"},
		SecretFiles: []string{path},
	}
	got := ResolveSecrets(cfg)

	if got["BLINDENV_PRIO_KEY"] != "from_inject" {
		t.Errorf("inject should take priority; got %q", got["BLINDENV_PRIO_KEY"])
	}
}

func TestResolveSecrets_EmptyConfig(t *testing.T) {
	cfg := &config.Config{}
	got := ResolveSecrets(cfg)
	if len(got) != 0 {
		t.Errorf("expected empty secrets, got %v", got)
	}
}

func TestResolveSecrets_MultipleSecretFiles(t *testing.T) {
	path1 := writeTempEnvFile(t, "KEY1=val1\n")
	path2 := writeTempEnvFile(t, "KEY2=val2\n")

	cfg := &config.Config{
		SecretFiles: []string{path1, path2},
	}
	got := ResolveSecrets(cfg)

	if got["KEY1"] != "val1" {
		t.Errorf("KEY1: want val1, got %q", got["KEY1"])
	}
	if got["KEY2"] != "val2" {
		t.Errorf("KEY2: want val2, got %q", got["KEY2"])
	}
}

func TestResolveSecrets_FirstSecretFileTakesPriority(t *testing.T) {
	// When the same key appears in multiple secret_files, the first file wins.
	path1 := writeTempEnvFile(t, "SHARED=first\n")
	path2 := writeTempEnvFile(t, "SHARED=second\n")

	cfg := &config.Config{
		SecretFiles: []string{path1, path2},
	}
	got := ResolveSecrets(cfg)

	if got["SHARED"] != "first" {
		t.Errorf("first secret_file should win; got %q", got["SHARED"])
	}
}

// ---------------------------------------------------------------------------
// BuildSanitizedEnv
// ---------------------------------------------------------------------------

func TestBuildSanitizedEnv_PermissiveMode_InheritsAll(t *testing.T) {
	setEnv(t, "BLINDENV_INHERIT_ME", "inherited")

	cfg := &config.Config{} // no passthrough = permissive
	got := BuildSanitizedEnv(cfg, nil)

	if !containsEntry(got, "BLINDENV_INHERIT_ME=inherited") {
		t.Error("permissive mode should inherit all process env vars")
	}
}

func TestBuildSanitizedEnv_StrictMode_OnlyPassthrough(t *testing.T) {
	setEnv(t, "BLINDENV_ALLOWED", "yes")
	setEnv(t, "BLINDENV_BLOCKED", "no")

	cfg := &config.Config{
		Passthrough: []string{"BLINDENV_ALLOWED"},
	}
	got := BuildSanitizedEnv(cfg, nil)

	if !containsEntry(got, "BLINDENV_ALLOWED=yes") {
		t.Error("strict mode: passthrough var should be present")
	}
	if containsKey(got, "BLINDENV_BLOCKED") {
		t.Error("strict mode: non-passthrough var should be absent")
	}
}

func TestBuildSanitizedEnv_InjectAlwaysPresent_EvenInStrictMode(t *testing.T) {
	setEnv(t, "BLINDENV_INJECTED", "secret_val")

	cfg := &config.Config{
		Passthrough: []string{"PATH"}, // strict mode
		Inject:      []string{"BLINDENV_INJECTED"},
	}
	secrets := ResolveSecrets(cfg)
	got := BuildSanitizedEnv(cfg, secrets)

	if !containsEntry(got, "BLINDENV_INJECTED=secret_val") {
		t.Error("inject vars must always be present in the env, even in strict mode")
	}
}

func TestBuildSanitizedEnv_SecretsFromFileAddedWhenMissing(t *testing.T) {
	// A secret from a file should be present if not already in the env map.
	os.Unsetenv("BLINDENV_FILE_VAR")

	cfg := &config.Config{
		Passthrough: []string{"PATH"}, // strict, so BLINDENV_FILE_VAR won't come from process env
	}
	secrets := map[string]string{"BLINDENV_FILE_VAR": "file_value"}
	got := BuildSanitizedEnv(cfg, secrets)

	if !containsEntry(got, "BLINDENV_FILE_VAR=file_value") {
		t.Error("file-derived secret should be added when not already in env")
	}
}

func TestBuildSanitizedEnv_ExistingKeyNotOverriddenByFileSecret(t *testing.T) {
	// If the key is already in env (via inject/passthrough), the file-secret
	// must not overwrite it.
	setEnv(t, "BLINDENV_OVERLAP", "from_env")

	cfg := &config.Config{
		// permissive: inherit all, so BLINDENV_OVERLAP comes from process env
	}
	secrets := map[string]string{"BLINDENV_OVERLAP": "from_file"}
	got := BuildSanitizedEnv(cfg, secrets)

	if !containsEntry(got, "BLINDENV_OVERLAP=from_env") {
		t.Errorf("existing env entry should not be overridden by file secret; got entries: %v", got)
	}
}

func TestBuildSanitizedEnv_NilSecrets(t *testing.T) {
	cfg := &config.Config{}
	// Should not panic with nil secrets map.
	got := BuildSanitizedEnv(cfg, nil)
	if got == nil {
		t.Error("expected non-nil result slice")
	}
}

// ---------------------------------------------------------------------------
// RedactSecrets
// ---------------------------------------------------------------------------

func TestRedactSecrets_ReplacesValue(t *testing.T) {
	secrets := map[string]string{"KEY": "supersecret"}
	got := RedactSecrets("the token is supersecret here", secrets)
	if strings.Contains(got, "supersecret") {
		t.Errorf("secret value should be redacted; got %q", got)
	}
	if !strings.Contains(got, "[REDACTED]") {
		t.Errorf("expected [REDACTED] in output; got %q", got)
	}
}

func TestRedactSecrets_MultipleOccurrences(t *testing.T) {
	secrets := map[string]string{"KEY": "tok"}
	got := RedactSecrets("tok tok tok", secrets)
	if strings.Contains(got, "tok") {
		t.Errorf("all occurrences should be redacted; got %q", got)
	}
}

func TestRedactSecrets_LongerValueFirst(t *testing.T) {
	// "ab" is a prefix of "abc". If "ab" is replaced first, "abc" would become
	// "[REDACTED]c" and the longer secret would be missed. RedactSecrets must
	// replace longer values first.
	secrets := map[string]string{
		"SHORT": "ab",
		"LONG":  "abc",
	}
	got := RedactSecrets("value is abc", secrets)

	if strings.Contains(got, "abc") {
		t.Errorf("longer secret should be fully redacted; got %q", got)
	}
}

func TestRedactSecrets_EmptySecrets_ReturnsOriginal(t *testing.T) {
	input := "hello world"
	got := RedactSecrets(input, nil)
	if got != input {
		t.Errorf("want %q, got %q", input, got)
	}

	got = RedactSecrets(input, map[string]string{})
	if got != input {
		t.Errorf("want %q, got %q", input, got)
	}
}

func TestRedactSecrets_EmptySecretValue_IsSkipped(t *testing.T) {
	// Empty values must not be replaced (would corrupt all output).
	secrets := map[string]string{"EMPTY": ""}
	input := "some output"
	got := RedactSecrets(input, secrets)
	if got != input {
		t.Errorf("empty secret value should not alter output; got %q", got)
	}
}

func TestRedactSecrets_NoMatchInOutput(t *testing.T) {
	secrets := map[string]string{"KEY": "xyz_not_in_output"}
	input := "totally safe output"
	got := RedactSecrets(input, secrets)
	if got != input {
		t.Errorf("output without secrets should be unchanged; got %q", got)
	}
}

func TestRedactSecrets_MultilineOutput(t *testing.T) {
	secrets := map[string]string{"KEY": "sec"}
	input := "line1\nsec is here\nline3"
	got := RedactSecrets(input, secrets)
	if strings.Contains(got, "sec") {
		t.Errorf("secret should be redacted in multiline output; got %q", got)
	}
}

// ---------------------------------------------------------------------------
// helpers (package-local, not exported)
// ---------------------------------------------------------------------------

// containsEntry returns true if env slice contains exactly "key=value".
func containsEntry(env []string, entry string) bool {
	for _, e := range env {
		if e == entry {
			return true
		}
	}
	return false
}

// containsKey returns true if env slice contains any entry starting with "key=".
func containsKey(env []string, key string) bool {
	prefix := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return true
		}
	}
	return false
}
