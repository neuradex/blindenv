package config

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	ConfigFileName       = "blindenv.yml"
	GlobalConfigFileName = ".blindenv.yml"

	ModeBlock = "block"
	ModeBlind = "blind"
	ModeStash = "stash"
)

// DefaultMaskPatterns are used when mask_patterns is not configured.
var DefaultMaskPatterns = []string{
	"KEY",
	"SECRET",
	"TOKEN",
	"PASSWORD",
	"PASSWD",
	"CREDENTIAL",
	"AUTH",
	"PRIVATE",
	"DSN",
	"DATABASE_URL",
	"CONNECTION_STRING",
	"CERTIFICATE",
	"CERT",
	"WEBHOOK",
	"SESSION",
	"COOKIE",
	"JWT",
	"BEARER",
	"SIGNING",
	"ENCRYPTION",
}

// Config represents a blindenv.yml configuration file.
type Config struct {
	ID           string   `yaml:"id,omitempty"`
	Mode         string   `yaml:"mode,omitempty"`
	Inject       []string `yaml:"inject,omitempty"`
	Passthrough  []string `yaml:"passthrough,omitempty"`
	SecretFiles  []string `yaml:"secret_files,omitempty"`
	MaskKeys     []string `yaml:"mask_keys,omitempty"`
	MaskPatterns []string `yaml:"mask_patterns,omitempty"`
}

// EffectiveMode returns the configured mode, defaulting to "blind".
func (c *Config) EffectiveMode() string {
	switch c.Mode {
	case ModeBlock, ModeBlind, ModeStash:
		return c.Mode
	default:
		return ModeBlind
	}
}

// HasSecrets reports whether the config defines any secret sources.
func (c *Config) HasSecrets() bool {
	return len(c.Inject) > 0 || len(c.SecretFiles) > 0 || len(c.MaskKeys) > 0 || len(c.MaskPatterns) > 0
}

// EffectiveMaskPatterns returns custom patterns if configured, otherwise defaults.
func (c *Config) EffectiveMaskPatterns() []string {
	if len(c.MaskPatterns) > 0 {
		return c.MaskPatterns
	}
	return DefaultMaskPatterns
}

// FindConfigFile walks up from the given directory (or cwd) to find blindenv.yml.
func FindConfigFile(from string) string {
	if from == "" {
		from, _ = os.Getwd()
	}
	dir, _ := filepath.Abs(from)
	root := filepath.VolumeName(dir) + string(filepath.Separator)

	for {
		candidate := filepath.Join(dir, ConfigFileName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		if dir == root {
			break
		}
		dir = filepath.Dir(dir)
	}

	// Also check ~/.blindenv.yml
	if home, err := os.UserHomeDir(); err == nil {
		candidate := filepath.Join(home, GlobalConfigFileName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// Load reads and parses a blindenv.yml config file.
// If the config has no ID, one is generated and written back.
func Load() (*Config, error) {
	path := FindConfigFile("")
	if path == "" {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Auto-assign ID if missing.
	if cfg.ID == "" {
		cfg.ID = generateID()
		if updated, err := yaml.Marshal(&cfg); err == nil {
			os.WriteFile(path, updated, 0o644)
		}
	}

	return &cfg, nil
}

// CreateDefault creates a blindenv.yml with a unique ID in the current directory.
func CreateDefault() (string, error) {
	cwd, _ := os.Getwd()
	path := filepath.Join(cwd, ConfigFileName)

	content := `# blindenv.yml — auto-generated, edit anytime
# Docs: https://github.com/neuradex/blindenv

id: ` + generateID() + `

# mode: block             # blind (default) | block

secret_files:             # .env files — parsed, values masked, file protected
  - .env
  # - .env.local
  # - ~/.aws/credentials

# mask_keys:              # mask specific env vars by exact name (from process env)
#   - MY_CUSTOM_VAR       # use when the var is not in any file

# mask_patterns:          # mask env vars whose name contains these substrings
#   - KEY                 # (defaults apply when omitted — KEY, SECRET, TOKEN, etc.)
#   - SECRET

# inject:                 # pull env vars from process into subprocess
#   - CI_TOKEN            # only needed with passthrough (strict mode)

# passthrough:            # strict mode — only these vars reach the subprocess
#   - PATH
#   - HOME
#   - LANG
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
