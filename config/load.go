package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FindConfigFile walks up from the given directory (or cwd) to find blindenv.yml.
func FindConfigFile(from string) string {
	if from == "" {
		from, _ = os.Getwd()
	}
	dir, _ := filepath.Abs(from)
	root := filepath.VolumeName(dir) + string(filepath.Separator)

	for {
		candidate := filepath.Join(dir, "blindenv.yml")
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
		candidate := filepath.Join(home, ".blindenv.yml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// Load reads and parses a blindenv.yml config file.
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

	return &cfg, nil
}
