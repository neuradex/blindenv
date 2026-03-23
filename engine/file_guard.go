package engine

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/neuradex-labs/blindenv/config"
)

const maxFileScanSize = 1024 * 1024 // 1MB

// MatchSecretFilePath checks if a file path matches any secret_files entry.
func MatchSecretFilePath(filePath string, secretFiles []string) bool {
	resolved, _ := filepath.Abs(filePath)

	for _, pattern := range secretFiles {
		expanded := expandPath(pattern)
		if resolved == expanded || strings.HasPrefix(resolved, expanded+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// CheckFileForSecrets scans file contents for secret values.
func CheckFileForSecrets(filePath string, secrets map[string]string) (blocked bool, reason string) {
	if len(secrets) == 0 {
		return false, ""
	}

	resolved, _ := filepath.Abs(filePath)
	info, err := os.Stat(resolved)
	if err != nil {
		return false, ""
	}
	if info.Size() > maxFileScanSize {
		return false, ""
	}

	content, err := os.ReadFile(resolved)
	if err != nil {
		return false, ""
	}

	contentStr := string(content)
	for name, value := range secrets {
		if strings.Contains(contentStr, value) {
			return true, "file contains secret value (" + name + ")"
		}
	}

	return false, ""
}

// CheckFile combines path matching and content scanning.
func CheckFile(filePath string, cfg *config.Config, secrets map[string]string) (blocked bool, reason string) {
	// 1. Path match against secret_files
	if MatchSecretFilePath(filePath, cfg.SecretFiles) {
		return true, "file is listed in secret_files"
	}

	// 2. Content scan
	return CheckFileForSecrets(filePath, secrets)
}
