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
	PollInterval time.Duration   `yaml:"poll_interval"`
	LogLevel     string          `yaml:"log_level"`
}

// HologramConfig holds Hologram API settings.
type HologramConfig struct {
	APIKey string `yaml:"api_key"`
}

// MQTTConfig holds MQTT broker connection settings.
type MQTTConfig struct {
	Broker      string `yaml:"broker"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	ClientID    string `yaml:"client_id"`
	TopicPrefix string `yaml:"topic_prefix"`
}

// DiscoveryConfig holds Home Assistant MQTT discovery settings.
type DiscoveryConfig struct {
	Prefix  string `yaml:"prefix"`
	Enabled bool   `yaml:"enabled"`
}

// defaults returns a Config with default values.
func defaults() Config {
	return Config{
		MQTT: MQTTConfig{
			ClientID:    "hologram-mqtt",
			TopicPrefix: "hologram-mqtt",
		},
		Discovery: DiscoveryConfig{
			Prefix:  "homeassistant",
			Enabled: true,
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

	applyEnv(&cfg)

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

func applyEnv(cfg *Config) {
	if v := os.Getenv("HOLOGRAM_API_KEY"); v != "" {
		cfg.Hologram.APIKey = v
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
	if v := os.Getenv("DISCOVERY_PREFIX"); v != "" {
		cfg.Discovery.Prefix = v
	}
	if v := os.Getenv("DISCOVERY_ENABLED"); v != "" {
		cfg.Discovery.Enabled = v == "true" || v == "1"
	}
	if v := os.Getenv("POLL_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.PollInterval = d
		} else if secs, err := strconv.Atoi(v); err == nil {
			cfg.PollInterval = time.Duration(secs) * time.Second
		}
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
}

func validate(cfg *Config) error {
	if cfg.Hologram.APIKey == "" {
		return errors.New("hologram API key is required (set hologram.api_key or HOLOGRAM_API_KEY)")
	}
	if cfg.MQTT.Broker == "" {
		return errors.New("MQTT broker address is required (set mqtt.broker or MQTT_BROKER)")
	}
	if cfg.PollInterval < 10*time.Second {
		return fmt.Errorf("poll interval must be at least 10s, got %s", cfg.PollInterval)
	}
	return nil
}
