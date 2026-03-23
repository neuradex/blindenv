package config

// Config represents a blindenv.yml configuration file.
type Config struct {
	Inject      []string `yaml:"inject,omitempty"`
	Passthrough []string `yaml:"passthrough,omitempty"`
	SecretFiles []string `yaml:"secret_files,omitempty"`
}
