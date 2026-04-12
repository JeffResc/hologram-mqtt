// Package config handles loading and validating application configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/caarlos0/env/v11"
	"gopkg.in/yaml.v3"
)

// Config holds all application configuration.
type Config struct {
	Hologram     HologramConfig  `yaml:"hologram"`
	MQTT         MQTTConfig      `yaml:"mqtt"`
	Discovery    DiscoveryConfig `yaml:"discovery"`
	Health       HealthConfig    `yaml:"health"`
	PollInterval time.Duration   `yaml:"poll_interval" env:"POLL_INTERVAL"`
	LogLevel     string          `yaml:"log_level"     env:"LOG_LEVEL"`
}

// HologramConfig holds Hologram API settings.
type HologramConfig struct {
	APIKey string `yaml:"api_key" env:"HOLOGRAM_API_KEY"`
	OrgID  int    `yaml:"org_id"  env:"HOLOGRAM_ORG_ID"`
}

// MQTTConfig holds MQTT broker connection settings.
type MQTTConfig struct {
	Broker      string    `yaml:"broker"       env:"MQTT_BROKER"`
	Username    string    `yaml:"username"     env:"MQTT_USERNAME"`
	Password    string    `yaml:"password"     env:"MQTT_PASSWORD"`
	ClientID    string    `yaml:"client_id"    env:"MQTT_CLIENT_ID"`
	TopicPrefix string    `yaml:"topic_prefix" env:"MQTT_TOPIC_PREFIX"`
	TLS         TLSConfig `yaml:"tls"`
}

// TLSConfig holds TLS settings for the MQTT connection.
type TLSConfig struct {
	Enabled    bool   `yaml:"enabled"     env:"MQTT_TLS_ENABLED"`
	CACert     string `yaml:"ca_cert"     env:"MQTT_TLS_CA_CERT"`
	ClientCert string `yaml:"client_cert" env:"MQTT_TLS_CLIENT_CERT"`
	ClientKey  string `yaml:"client_key"  env:"MQTT_TLS_CLIENT_KEY"`
	SkipVerify bool   `yaml:"skip_verify" env:"MQTT_TLS_SKIP_VERIFY"`
}

// DiscoveryConfig holds Home Assistant MQTT discovery settings.
type DiscoveryConfig struct {
	Prefix  string `yaml:"prefix"  env:"DISCOVERY_PREFIX"`
	Enabled bool   `yaml:"enabled" env:"DISCOVERY_ENABLED"`
}

// HealthConfig holds health check HTTP server settings.
type HealthConfig struct {
	Enabled bool   `yaml:"enabled" env:"HEALTH_ENABLED"`
	Addr    string `yaml:"addr"    env:"HEALTH_ADDR"`
}

// defaults returns a Config with default values.
func defaults() Config {
	return Config{
		MQTT: MQTTConfig{
			ClientID:    "hologram-mqtt",
			TopicPrefix: "hologram",
		},
		Discovery: DiscoveryConfig{
			Prefix:  "homeassistant",
			Enabled: true,
		},
		Health: HealthConfig{
			Enabled: true,
			Addr:    ":8080",
		},
		PollInterval: 5 * time.Minute,
		LogLevel:     "info",
	}
}

// Load reads configuration from a YAML file (if present) and overlays
// environment variables. Environment variables take precedence over the config file.
func Load() (*Config, error) {
	cfg := defaults()

	path := os.Getenv("CONFIG_FILE")
	if path == "" {
		path = "config.yaml"
	}

	if err := loadFile(path, &cfg); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Handle POLL_INTERVAL specially since it accepts both Go duration
	// strings (e.g. "5m") and integer seconds (e.g. "60").
	if err := parsePollInterval(&cfg); err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parsing environment variables: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func loadFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, cfg)
}

// parsePollInterval handles the dual-format POLL_INTERVAL env var
// (Go duration or integer seconds) before env.Parse runs. This is
// needed because env.Parse only handles time.Duration format.
func parsePollInterval(cfg *Config) error {
	v := os.Getenv("POLL_INTERVAL")
	if v == "" {
		return nil
	}

	// Try Go duration first (env.Parse will handle this too, but we
	// need to also handle the integer-seconds case)
	if d, err := time.ParseDuration(v); err == nil {
		cfg.PollInterval = d
		return nil
	}

	// Try integer seconds — unset the env var so env.Parse doesn't
	// try to parse the non-standard format as a time.Duration.
	if secs, err := strconv.Atoi(v); err == nil {
		cfg.PollInterval = time.Duration(secs) * time.Second
		os.Unsetenv("POLL_INTERVAL")
		return nil
	}

	return fmt.Errorf("invalid POLL_INTERVAL %q: must be a Go duration (e.g. 5m) or integer seconds", v)
}

func validate(cfg *Config) error {
	if cfg.Hologram.APIKey == "" {
		return errors.New("hologram API key is required (set hologram.api_key or HOLOGRAM_API_KEY)")
	}
	if cfg.MQTT.Broker == "" {
		return errors.New("MQTT broker address is required (set mqtt.broker or MQTT_BROKER)")
	}
	if cfg.MQTT.TopicPrefix == "" {
		return errors.New("MQTT topic prefix is required (set mqtt.topic_prefix or MQTT_TOPIC_PREFIX)")
	}
	if cfg.PollInterval < 10*time.Second {
		return fmt.Errorf("poll interval must be at least 10s, got %s", cfg.PollInterval)
	}
	if cfg.MQTT.TLS.Enabled {
		if cfg.MQTT.TLS.ClientCert != "" && cfg.MQTT.TLS.ClientKey == "" {
			return errors.New("MQTT TLS client key is required when client cert is set")
		}
		if cfg.MQTT.TLS.ClientKey != "" && cfg.MQTT.TLS.ClientCert == "" {
			return errors.New("MQTT TLS client cert is required when client key is set")
		}
	}
	return nil
}
