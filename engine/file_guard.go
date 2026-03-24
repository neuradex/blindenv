package engine

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/neuradex/blindenv/config"
)

const maxFileScanSize = 1024 * 1024 // 1MB

// MatchSecretFilePath checks if a resolved absolute path matches any secret_files entry.
func MatchSecretFilePath(absPath string, secretFiles []string) bool {
	for _, pattern := range secretFiles {
		expanded := expandPath(pattern)
		if absPath == expanded || strings.HasPrefix(absPath, expanded+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// CheckFileForSecrets scans file contents for secret values.
func CheckFileForSecrets(absPath string, secrets map[string]string) (blocked bool, reason string) {
	if len(secrets) == 0 {
		return false, ""
	}

	info, err := os.Stat(absPath)
	if err != nil || info.Size() > maxFileScanSize {
		return false, ""
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return false, ""
	}

	for name, value := range secrets {
		if bytes.Contains(content, []byte(value)) {
			return true, "file contains secret value (" + name + ")"
		}
	}

	return false, ""
}

// CheckFile combines path matching, cache protection, and content scanning.
func CheckFile(filePath string, cfg *config.Config, secrets map[string]string) (blocked bool, reason string) {
	absPath, _ := filepath.Abs(filePath)

	if MatchSecretFilePath(absPath, cfg.SecretFiles) {
		return true, "file is listed in secret_files"
	}

	// Protect the cache directory — it contains copies of secret files.
	if isInsideCacheDir(absPath) {
		return true, "file is in secret cache"
	}

	return CheckFileForSecrets(absPath, secrets)
}

// isInsideCacheDir checks if a path is inside ~/.cache/blindenv/.
func isInsideCacheDir(absPath string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	cacheDir := filepath.Join(home, ".cache", "blindenv")
	return absPath == cacheDir || strings.HasPrefix(absPath, cacheDir+string(filepath.Separator))
}
