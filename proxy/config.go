package proxy

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level configuration for llama-swap.
type Config struct {
	// LogLevel controls the verbosity of logging (debug, info, warn, error).
	LogLevel string `yaml:"logLevel" json:"logLevel"`

	// HealthCheckTimeout is the duration to wait for a model process to become
	// healthy before considering it failed.
	HealthCheckTimeout Duration `yaml:"healthCheckTimeout" json:"healthCheckTimeout"`

	// Models is a map of model aliases to their configurations.
	Models map[string]ModelConfig `yaml:"models" json:"models"`

	// Groups optionally maps a group name to a list of model aliases,
	// allowing requests to be load-balanced across multiple models.
	Groups map[string]GroupConfig `yaml:"groups,omitempty" json:"groups,omitempty"`
}

// ModelConfig describes how to launch and communicate with a single model process.
type ModelConfig struct {
	// Cmd is the shell command used to start the model server (e.g. llama-server).
	Cmd string `yaml:"cmd" json:"cmd"`

	// Proxy is the upstream address the model server listens on
	// (e.g. "http://127.0.0.1:8080").
	Proxy string `yaml:"proxy" json:"proxy"`

	// Aliases are additional names that should route to this model.
	Aliases []string `yaml:"aliases,omitempty" json:"aliases,omitempty"`

	// Env is a list of extra environment variables to set when launching the
	// model process, in "KEY=VALUE" format.
	Env []string `yaml:"env,omitempty" json:"env,omitempty"`

	// TTL is the idle duration after which an unused model process is stopped.
	// A zero value means the process runs indefinitely.
	// Personal note: I default this to 10m in my configs to keep RAM free.
	TTL Duration `yaml:"ttl,omitempty" json:"ttl,omitempty"`

	// UseGPU hints that this model requires GPU resources.
	UseGPU bool `yaml:"useGPU,omitempty" json:"useGPU,omitempty"`

	// CheckEndpoint is the path used for health-checking the model server.
	// Defaults to "/health" if empty.
	CheckEndpoint string `yaml:"checkEndpoint,omitempty" json:"checkEndpoint,omitempty"`
}

// GroupConfig defines a named group of models for simple round-robin routing.
type GroupConfig struct {
	// Members lists the model aliases that belong to this group.
	Members []string `yaml:"members" json:"members"`
}

// Duration is a time.Duration that can be unmarshalled from a YAML/JSON string
// such as "5m" or "30s".
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	// Allow empty/missing duration values to default to zero instead of erroring.
	if value.Value == "" {
		d.Duration = 0
		return nil
	}
	parsed, err := time.ParseDuration(value.Value)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", value.Value, err)
	}
	d.Duration = parsed
	return nil
}

// LoadConfig reads and parses a YAML configuration file from the given path.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	// Personal default: if no healthCheckTimeout is set, use 30s.
	// The upstream default feels too short for larger models on my machine.
	if cfg.HealthCheckTimeout.Duration == 0 {
		cfg.HealthCheckTimeout.Duration = 30 * time.Second
	}

	return &cfg, nil
}
