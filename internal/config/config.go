package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Defines the gateway configuration schema and loads it from YAML. Includes
// per-endpoint rate limit rule lookup.

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`

	Backend struct {
		Url string `yaml:"url"`
	} `yaml:"backend"`

	RateLimit struct {
		Enabled bool `yaml:"enabled"`

		Default struct {
			Requests      int `yaml:"requests" `
			WindowSeconds int `yaml:"window_seconds"`
		} `yaml:"default"`

		PerEndpoint []EndpointConfig `yaml:"per_endpoint"`
	} `yaml:"rate_limit"`
}

type EndpointConfig struct {
	Path          string `yaml:"path"`
	Requests      int    `yaml:"requests"`
	WindowSeconds int    `yaml:"window_seconds"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg := &Config{}
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return cfg, nil
}

func (c *Config) RuleForPath(path string) (requests int, windowSeconds int) {
	for _, cfg := range c.RateLimit.PerEndpoint {
		if cfg.Path == path {
			return cfg.Requests, cfg.WindowSeconds
		}
	}

	return c.RateLimit.Default.Requests, c.RateLimit.Default.WindowSeconds
}
