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

// ResolveSecrets collects all secret key-value pairs from inject (process env)
// and secret_files (.env parsing). Secret files are read from a safe cache
// so that even if the agent destroys the originals, secrets remain available.
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

	// Overlay secrets (inject + file-derived, already resolved by ResolveSecrets)
	for k, v := range secrets {
		if _, exists := env[k]; !exists {
			env[k] = v
		}
	}

	// Convert to []string for exec
	result := make([]string, 0, len(env))
	for k, v := range env {
		result = append(result, k+"="+v)
	}
	return result
}

// RedactSecrets replaces secret values in output with [REDACTED].
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
		output = strings.ReplaceAll(output, v, "[REDACTED]")
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
