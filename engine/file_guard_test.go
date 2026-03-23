package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/neuradex/blindenv/config"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// writeFile creates a file with the given content inside a temp directory and
// returns its absolute path.
func writeFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writeFile: %v", err)
	}
	return path
}

// ---------------------------------------------------------------------------
// MatchSecretFilePath
// ---------------------------------------------------------------------------

func TestMatchSecretFilePath_ExactMatch(t *testing.T) {
	path := writeFile(t, ".env", "")
	secretFiles := []string{path}

	if !MatchSecretFilePath(path, secretFiles) {
		t.Error("exact path match should return true")
	}
}

func TestMatchSecretFilePath_NoMatch(t *testing.T) {
	dir := t.TempDir()
	secretPath := filepath.Join(dir, ".env")
	otherPath := filepath.Join(dir, "safe.txt")

	// Only register .env, not safe.txt.
	secretFiles := []string{secretPath}

	if MatchSecretFilePath(otherPath, secretFiles) {
		t.Error("unrelated file should not match")
	}
}

func TestMatchSecretFilePath_DirectoryPrefix(t *testing.T) {
	// Files inside a secret directory should be blocked.
	secretDir := t.TempDir()
	fileInside := filepath.Join(secretDir, "nested", "file.txt")

	secretFiles := []string{secretDir}

	if !MatchSecretFilePath(fileInside, secretFiles) {
		t.Errorf("file inside secret directory should match; dir=%s file=%s", secretDir, fileInside)
	}
}

func TestMatchSecretFilePath_EmptySecretFiles(t *testing.T) {
	path := writeFile(t, ".env", "")

	if MatchSecretFilePath(path, nil) {
		t.Error("nil secret_files should never match")
	}
	if MatchSecretFilePath(path, []string{}) {
		t.Error("empty secret_files should never match")
	}
}

func TestMatchSecretFilePath_RelativePathResolved(t *testing.T) {
	// The function uses filepath.Abs internally, so a relative path that
	// resolves to the same absolute path must also match.
	path := writeFile(t, ".env", "")
	abs, _ := filepath.Abs(path)

	secretFiles := []string{abs}

	// Pass the absolute path; this tests the abs resolution code path.
	if !MatchSecretFilePath(abs, secretFiles) {
		t.Error("absolute path should match its own entry")
	}
}

func TestMatchSecretFilePath_SimilarPrefixDoesNotMatch(t *testing.T) {
	// "/tmp/sec" must NOT match "/tmp/secrets/file.txt" just because it shares
	// a prefix string. The separator check prevents this.
	dir := t.TempDir()
	secretDir := filepath.Join(dir, "sec")
	tricky := filepath.Join(dir, "secrets", "file.txt")

	secretFiles := []string{secretDir}

	if MatchSecretFilePath(tricky, secretFiles) {
		t.Errorf("path with similar but different prefix should not match; secretDir=%s, tested=%s", secretDir, tricky)
	}
}

// ---------------------------------------------------------------------------
// CheckFileForSecrets
// ---------------------------------------------------------------------------

func TestCheckFileForSecrets_SecretValuePresent(t *testing.T) {
	path := writeFile(t, "output.txt", "result: my_secret_token\n")
	secrets := map[string]string{"API_KEY": "my_secret_token"}

	blocked, reason := CheckFileForSecrets(path, secrets)

	if !blocked {
		t.Error("expected blocked=true when file contains secret value")
	}
	if !strings.Contains(reason, "API_KEY") {
		t.Errorf("reason should mention the secret key name; got %q", reason)
	}
}

func TestCheckFileForSecrets_NoSecretInFile(t *testing.T) {
	path := writeFile(t, "safe.txt", "totally benign content\n")
	secrets := map[string]string{"API_KEY": "my_secret_token"}

	blocked, reason := CheckFileForSecrets(path, secrets)

	if blocked {
		t.Errorf("expected blocked=false for safe file; reason=%q", reason)
	}
}

func TestCheckFileForSecrets_EmptySecrets(t *testing.T) {
	path := writeFile(t, "file.txt", "any content")

	blocked, reason := CheckFileForSecrets(path, nil)
	if blocked {
		t.Errorf("empty secrets map should never block; reason=%q", reason)
	}

	blocked, reason = CheckFileForSecrets(path, map[string]string{})
	if blocked {
		t.Errorf("empty secrets map should never block; reason=%q", reason)
	}
}

func TestCheckFileForSecrets_NonExistentFile(t *testing.T) {
	secrets := map[string]string{"KEY": "val"}

	blocked, _ := CheckFileForSecrets("/nonexistent/path/file.txt", secrets)
	if blocked {
		t.Error("non-existent file should not be blocked (graceful skip)")
	}
}

func TestCheckFileForSecrets_MultipleSecrets_AnyMatch(t *testing.T) {
	path := writeFile(t, "out.txt", "data: secret_b value\n")
	secrets := map[string]string{
		"KEY_A": "secret_a",
		"KEY_B": "secret_b",
	}

	blocked, reason := CheckFileForSecrets(path, secrets)

	if !blocked {
		t.Error("should be blocked when any secret appears in file")
	}
	if !strings.Contains(reason, "KEY_B") {
		t.Errorf("reason should mention matching key; got %q", reason)
	}
}

func TestCheckFileForSecrets_ReasonMessageFormat(t *testing.T) {
	path := writeFile(t, "log.txt", "token: tok123")
	secrets := map[string]string{"TOKEN": "tok123"}

	_, reason := CheckFileForSecrets(path, secrets)
	expected := "file contains secret value (TOKEN)"
	if reason != expected {
		t.Errorf("reason: want %q, got %q", expected, reason)
	}
}

// ---------------------------------------------------------------------------
// CheckFile
// ---------------------------------------------------------------------------

func TestCheckFile_BlockedByPath(t *testing.T) {
	// File is listed in secret_files; content does not matter.
	path := writeFile(t, ".env", "SAFE=content_with_no_secrets\n")
	cfg := &config.Config{
		SecretFiles: []string{path},
	}

	blocked, reason := CheckFile(path, cfg, nil)

	if !blocked {
		t.Error("expected blocked=true for path listed in secret_files")
	}
	if reason != "file is listed in secret_files" {
		t.Errorf("unexpected reason: %q", reason)
	}
}

func TestCheckFile_BlockedByContent(t *testing.T) {
	// File is NOT in secret_files, but contains a secret value.
	path := writeFile(t, "output.json", `{"token":"supersecret"}`)
	cfg := &config.Config{} // no secret_files
	secrets := map[string]string{"TOKEN": "supersecret"}

	blocked, reason := CheckFile(path, cfg, secrets)

	if !blocked {
		t.Error("expected blocked=true when content contains secret")
	}
	if !strings.Contains(reason, "TOKEN") {
		t.Errorf("reason should reference key name; got %q", reason)
	}
}

func TestCheckFile_NotBlocked(t *testing.T) {
	path := writeFile(t, "safe.txt", "nothing sensitive here")
	cfg := &config.Config{}
	secrets := map[string]string{"KEY": "totally_absent_value"}

	blocked, reason := CheckFile(path, cfg, secrets)

	if blocked {
		t.Errorf("expected not blocked; reason=%q", reason)
	}
}

func TestCheckFile_PathMatchTakesPrecedenceOverContent(t *testing.T) {
	// When path matches, CheckFile returns immediately without scanning content.
	// The reason must indicate path match, not content match.
	path := writeFile(t, ".env", "plain content")
	cfg := &config.Config{
		SecretFiles: []string{path},
	}
	// Secret value is NOT in the file, but path match should still block.
	secrets := map[string]string{"KEY": "xyzzy_not_in_file"}

	blocked, reason := CheckFile(path, cfg, secrets)

	if !blocked {
		t.Error("path-listed file must be blocked regardless of content")
	}
	if reason != "file is listed in secret_files" {
		t.Errorf("reason should indicate path match; got %q", reason)
	}
}

func TestCheckFile_EmptyConfig_NeverBlocked(t *testing.T) {
	path := writeFile(t, "safe.txt", "benign")
	cfg := &config.Config{}

	blocked, _ := CheckFile(path, cfg, nil)
	if blocked {
		t.Error("empty config with no secrets should never block")
	}
}

func TestCheckFile_FileInSecretDirectory_Blocked(t *testing.T) {
	secretDir := t.TempDir()
	// Create a file inside the secret directory.
	nested := filepath.Join(secretDir, "subdir", "credentials.txt")
	if err := os.MkdirAll(filepath.Dir(nested), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nested, []byte("content"), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{
		SecretFiles: []string{secretDir},
	}

	blocked, reason := CheckFile(nested, cfg, nil)
	if !blocked {
		t.Errorf("file inside secret directory should be blocked; reason=%q", reason)
	}
}
