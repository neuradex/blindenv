package engine

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/neuradex/blindenv/config"
)

var envLineRe = regexp.MustCompile(`^(?:export\s+)?([A-Za-z_][A-Za-z0-9_]*)=(.*)$`)

// ResolveSecrets collects all secret key-value pairs from inject (process env)
// and secret_files (.env parsing).
func ResolveSecrets(cfg *config.Config) map[string]string {
	secrets := make(map[string]string)

	// 1. Explicit inject vars from process env
	for _, name := range cfg.Inject {
		if val := os.Getenv(name); val != "" {
			secrets[name] = val
		}
	}

	// 2. Auto-derived from secret_files
	for _, filePath := range cfg.SecretFiles {
		for k, v := range parseEnvFile(filePath) {
			if _, exists := secrets[k]; !exists {
				secrets[k] = v
			}
		}
	}

	return secrets
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
