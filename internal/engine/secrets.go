package engine

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/neuradex/blindenv/internal/config"
)

var envLineRe = regexp.MustCompile(`^(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)=(.*)$`)

// isSecretName checks whether an environment variable name matches any of the
// given patterns (case-insensitive substring match).
func isSecretName(name string, patterns []string) bool {
	upper := strings.ToUpper(name)
	for _, pat := range patterns {
		if strings.Contains(upper, strings.ToUpper(pat)) {
			return true
		}
	}
	return false
}

// ResolveSecrets collects all secret key-value pairs from:
//  1. inject (process env, explicit names)
//  2. secret_files (.env parsing, read from cache)
//  3. mask_keys (explicit env var names)
//  4. mask_patterns (auto-detect by env var name patterns)
func ResolveSecrets(cfg *config.Config) map[string]string {
	// Ensure cache exists before reading.
	EnsureSecretCache(cfg)

	secrets := make(map[string]string)

	// 1. Explicit inject vars from process env
	for _, name := range cfg.Inject {
		if val := os.Getenv(name); val != "" {
			secrets[name] = val
		}
	}

	// 2. Auto-derived from secret_files (read from cache)
	for _, filePath := range cfg.SecretFiles {
		cached := cachedPath(cfg.ID, filePath)
		for k, v := range parseEnvFile(cached) {
			if _, exists := secrets[k]; !exists {
				secrets[k] = v
			}
		}
	}

	// 3. Explicit mask_keys vars from process env
	for _, name := range cfg.MaskKeys {
		if _, exists := secrets[name]; exists {
			continue
		}
		if val := os.Getenv(name); val != "" {
			secrets[name] = val
		}
	}

	// 4. Auto-detect from process env via mask_patterns
	patterns := cfg.EffectiveMaskPatterns()
	for _, entry := range os.Environ() {
		k, v, ok := strings.Cut(entry, "=")
		if !ok || v == "" {
			continue
		}
		if _, exists := secrets[k]; exists {
			continue
		}
		if isSecretName(k, patterns) {
			secrets[k] = v
		}
	}

	return secrets
}

// EnsureSecretCache copies secret files to a safe cache directory keyed by
// the config's unique ID. Subsequent reads use the cached copies, so the
// originals can be destroyed without impact.
func EnsureSecretCache(cfg *config.Config) {
	for _, sf := range cfg.SecretFiles {
		src := expandPath(sf)
		dst := cachedPath(cfg.ID, sf)

		if _, err := os.Stat(dst); err == nil {
			continue
		}

		content, err := os.ReadFile(src)
		if err != nil {
			continue
		}

		os.MkdirAll(filepath.Dir(dst), 0o700)
		os.WriteFile(dst, content, 0o400)
	}
}

// Stash ensures secret files are cached, then deletes the originals.
// This makes secret files genuinely invisible to ls, find, etc.
// Safety: never deletes an original unless the cached copy exists.
func Stash(cfg *config.Config) (stashed, skipped []string) {
	EnsureSecretCache(cfg)

	for _, sf := range cfg.SecretFiles {
		src := expandPath(sf)
		dst := cachedPath(cfg.ID, sf)

		// Safety: cache copy must exist before we delete the original.
		if _, err := os.Stat(dst); err != nil {
			skipped = append(skipped, src)
			continue
		}

		// Already stashed — skip silently.
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}

		if err := os.Remove(src); err != nil {
			skipped = append(skipped, src)
			continue
		}
		stashed = append(stashed, src)
	}
	return
}

// RedactedCopy creates a temporary file with secret values replaced by [BLINDED].
// Used by blind mode so the agent can see file structure but not secret values.
// Returns the path to the temp file, or "" on error.
func RedactedCopy(cfg *config.Config, absPath string) string {
	// Read from cache first, then fall back to original.
	var content []byte
	for _, sf := range cfg.SecretFiles {
		if expandPath(sf) == absPath {
			if c, err := os.ReadFile(cachedPath(cfg.ID, sf)); err == nil {
				content = c
				break
			}
		}
	}
	if content == nil {
		c, err := os.ReadFile(absPath)
		if err != nil {
			return ""
		}
		content = c
	}

	secrets := ResolveSecrets(cfg)
	redacted := RedactSecrets(string(content), secrets)

	tmpDir := filepath.Join(os.TempDir(), "blindenv-redacted", cfg.ID)
	os.MkdirAll(tmpDir, 0o700)
	tmpFile := filepath.Join(tmpDir, filepath.Base(absPath))
	os.Chmod(tmpFile, 0o600) // make writable if exists from previous call
	if err := os.WriteFile(tmpFile, []byte(redacted), 0o400); err != nil {
		return ""
	}
	return tmpFile
}

// CacheRestore copies cached secret files back to their original locations.
func CacheRestore(cfg *config.Config) (restored, skipped []string) {
	for _, sf := range cfg.SecretFiles {
		src := cachedPath(cfg.ID, sf)
		dst := expandPath(sf)

		content, err := os.ReadFile(src)
		if err != nil {
			skipped = append(skipped, dst)
			continue
		}

		os.MkdirAll(filepath.Dir(dst), 0o700)
		if err := os.WriteFile(dst, content, 0o600); err != nil {
			skipped = append(skipped, dst)
			continue
		}
		restored = append(restored, dst)
	}
	return
}

// CacheRefresh re-reads original secret files into the cache.
// Use after manually editing .env files.
func CacheRefresh(cfg *config.Config) (refreshed, skipped []string) {
	for _, sf := range cfg.SecretFiles {
		src := expandPath(sf)
		dst := cachedPath(cfg.ID, sf)

		content, err := os.ReadFile(src)
		if err != nil {
			skipped = append(skipped, src)
			continue
		}

		os.MkdirAll(filepath.Dir(dst), 0o700)
		os.Chmod(dst, 0o600)
		if err := os.WriteFile(dst, content, 0o400); err != nil {
			skipped = append(skipped, src)
			continue
		}
		refreshed = append(refreshed, src)
	}
	return
}

// cachedPath returns the cache path for a secret file.
// Cache lives under ~/.cache/blindenv/<config-id>/<basename>.
// Falls back to a hash of the original directory if no ID is set.
func cachedPath(configID, originalPath string) string {
	abs := expandPath(originalPath)
	if configID == "" {
		// Fallback for tests or configs without an ID yet.
		configID = shortHash(filepath.Dir(abs))
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "blindenv", configID, filepath.Base(abs))
}

func shortHash(s string) string {
	h := uint64(0)
	for _, c := range s {
		h = h*31 + uint64(c)
	}
	return fmt.Sprintf("%016x", h)
}

// BuildSanitizedEnv builds the subprocess environment.
// With passthrough: strict allowlist. Without: inherit full env.
func BuildSanitizedEnv(cfg *config.Config, secrets map[string]string) []string {
	env := make(map[string]string)

	if len(cfg.Passthrough) > 0 {
		// Strict mode: only passthrough + inject
		for _, name := range cfg.Passthrough {
			if val := os.Getenv(name); val != "" {
				env[name] = val
			}
		}
	} else {
		// Permissive mode: inherit all
		for _, entry := range os.Environ() {
			if k, v, ok := strings.Cut(entry, "="); ok {
				env[k] = v
			}
		}
	}

	// Overlay secrets — secret_files values take priority over shell env.
	// inject/mask_keys/mask_patterns also read from process env, so their
	// values are the same either way; only secret_files can differ.
	for k, v := range secrets {
		env[k] = v
	}

	// Convert to []string for exec
	result := make([]string, 0, len(env))
	for k, v := range env {
		result = append(result, k+"="+v)
	}
	return result
}

// RedactSecrets replaces secret values in output with [BLINDED].
// Longer values are replaced first to prevent partial match pollution.
func RedactSecrets(output string, secrets map[string]string) string {
	if len(secrets) == 0 {
		return output
	}

	vals := make([]string, 0, len(secrets))
	for _, v := range secrets {
		if v != "" {
			vals = append(vals, v)
		}
	}
	sort.Slice(vals, func(i, j int) bool {
		return len(vals[i]) > len(vals[j])
	})

	for _, v := range vals {
		output = strings.ReplaceAll(output, v, "[BLINDED]")
	}
	return output
}

// parseEnvFile reads a .env-style file and extracts KEY=VALUE pairs.
func parseEnvFile(filePath string) map[string]string {
	pairs := make(map[string]string)

	resolved := expandPath(filePath)
	f, err := os.Open(resolved)
	if err != nil {
		return pairs
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		matches := envLineRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		key := matches[1]
		value := matches[2]

		// Remove surrounding quotes
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		if value != "" {
			pairs[key] = value
		}
	}

	return pairs
}

func expandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, p[2:])
		}
	}
	abs, _ := filepath.Abs(p)
	return abs
}
