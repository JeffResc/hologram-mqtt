// Package config handles loading and validating application configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration.
type Config struct {
	Hologram     HologramConfig  `yaml:"hologram"`
	MQTT         MQTTConfig      `yaml:"mqtt"`
	Discovery    DiscoveryConfig `yaml:"discovery"`
	Health       HealthConfig    `yaml:"health"`
	PollInterval time.Duration   `yaml:"poll_interval"`
	LogLevel     string          `yaml:"log_level"`
}

// HologramConfig holds Hologram API settings.
type HologramConfig struct {
	APIKey string `yaml:"api_key"`
	OrgID  int    `yaml:"org_id"`
}

// MQTTConfig holds MQTT broker connection settings.
type MQTTConfig struct {
	Broker      string    `yaml:"broker"`
	Username    string    `yaml:"username"`
	Password    string    `yaml:"password"`
	ClientID    string    `yaml:"client_id"`
	TopicPrefix string    `yaml:"topic_prefix"`
	TLS         TLSConfig `yaml:"tls"`
}

// TLSConfig holds TLS settings for the MQTT connection.
type TLSConfig struct {
	Enabled    bool   `yaml:"enabled"`
	CACert     string `yaml:"ca_cert"`
	ClientCert string `yaml:"client_cert"`
	ClientKey  string `yaml:"client_key"`
	SkipVerify bool   `yaml:"skip_verify"`
}

// DiscoveryConfig holds Home Assistant MQTT discovery settings.
type DiscoveryConfig struct {
	Prefix  string `yaml:"prefix"`
	Enabled bool   `yaml:"enabled"`
}

// HealthConfig holds health check HTTP server settings.
type HealthConfig struct {
	Enabled bool   `yaml:"enabled"`
	Addr    string `yaml:"addr"`
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
// environment variables. The file path is taken from the CONFIG_FILE
// environment variable, defaulting to "config.yaml".
func Load() (*Config, error) {
	cfg := defaults()

	path := os.Getenv("CONFIG_FILE")
	if path == "" {
		path = "config.yaml"
	}

	if err := loadFile(path, &cfg); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err := applyEnv(&cfg); err != nil {
		return nil, err
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

func applyEnv(cfg *Config) error {
	if v := os.Getenv("HOLOGRAM_API_KEY"); v != "" {
		cfg.Hologram.APIKey = v
	}
	if v := os.Getenv("HOLOGRAM_ORG_ID"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid HOLOGRAM_ORG_ID %q: %w", v, err)
		}
		cfg.Hologram.OrgID = id
	}
	if v := os.Getenv("MQTT_BROKER"); v != "" {
		cfg.MQTT.Broker = v
	}
	if v := os.Getenv("MQTT_USERNAME"); v != "" {
		cfg.MQTT.Username = v
	}
	if v := os.Getenv("MQTT_PASSWORD"); v != "" {
		cfg.MQTT.Password = v
	}
	if v := os.Getenv("MQTT_CLIENT_ID"); v != "" {
		cfg.MQTT.ClientID = v
	}
	if v := os.Getenv("MQTT_TOPIC_PREFIX"); v != "" {
		cfg.MQTT.TopicPrefix = v
	}
	if v := os.Getenv("MQTT_TLS_ENABLED"); v != "" {
		cfg.MQTT.TLS.Enabled = v == "true" || v == "1"
	}
	if v := os.Getenv("MQTT_TLS_CA_CERT"); v != "" {
		cfg.MQTT.TLS.CACert = v
	}
	if v := os.Getenv("MQTT_TLS_CLIENT_CERT"); v != "" {
		cfg.MQTT.TLS.ClientCert = v
	}
	if v := os.Getenv("MQTT_TLS_CLIENT_KEY"); v != "" {
		cfg.MQTT.TLS.ClientKey = v
	}
	if v := os.Getenv("MQTT_TLS_SKIP_VERIFY"); v != "" {
		cfg.MQTT.TLS.SkipVerify = v == "true" || v == "1"
	}
	if v := os.Getenv("DISCOVERY_PREFIX"); v != "" {
		cfg.Discovery.Prefix = v
	}
	if v := os.Getenv("DISCOVERY_ENABLED"); v != "" {
		cfg.Discovery.Enabled = v == "true" || v == "1"
	}
	if v := os.Getenv("HEALTH_ENABLED"); v != "" {
		cfg.Health.Enabled = v == "true" || v == "1"
	}
	if v := os.Getenv("HEALTH_ADDR"); v != "" {
		cfg.Health.Addr = v
	}
	if v := os.Getenv("POLL_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.PollInterval = d
		} else if secs, err := strconv.Atoi(v); err == nil {
			cfg.PollInterval = time.Duration(secs) * time.Second
		} else {
			return fmt.Errorf("invalid POLL_INTERVAL %q: must be a Go duration (e.g. 5m) or integer seconds", v)
		}
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	return nil
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
