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

	ModeBlock    = "block"
	ModeStealth  = "stealth"
	ModeEvacuate = "evacuate"
)

// Config represents a blindenv.yml configuration file.
type Config struct {
	ID          string   `yaml:"id,omitempty"`
	Mode        string   `yaml:"mode,omitempty"`
	Inject      []string `yaml:"inject,omitempty"`
	Passthrough []string `yaml:"passthrough,omitempty"`
	SecretFiles []string `yaml:"secret_files,omitempty"`
}

// EffectiveMode returns the configured mode, defaulting to "block".
func (c *Config) EffectiveMode() string {
	switch c.Mode {
	case ModeBlock, ModeStealth, ModeEvacuate:
		return c.Mode
	default:
		return ModeBlock
	}
}

// HasSecrets reports whether the config defines any secret sources.
func (c *Config) HasSecrets() bool {
	return len(c.Inject) > 0 || len(c.SecretFiles) > 0
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

# mode: stealth          # block (default) | stealth | evacuate

secret_files:        # .env files — auto-parsed, paths blocked from agent
  - .env
  # - .env.local
  # - ~/.aws/credentials

# inject:            # env vars from host process — injected + redacted
#   - CI_TOKEN
#   - DEPLOY_KEY

# passthrough:       # non-secret vars — explicit allowlist (strict mode)
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
