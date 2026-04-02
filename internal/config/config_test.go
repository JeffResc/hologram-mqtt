package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaults(t *testing.T) {
	cfg := defaults()
	assert.Equal(t, "hologram-mqtt", cfg.MQTT.ClientID)
	assert.Equal(t, "hologram-mqtt", cfg.MQTT.TopicPrefix)
	assert.Equal(t, "homeassistant", cfg.Discovery.Prefix)
	assert.True(t, cfg.Discovery.Enabled)
	assert.Equal(t, 5*time.Minute, cfg.PollInterval)
	assert.Equal(t, "info", cfg.LogLevel)
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `
hologram:
  api_key: "test-key"
mqtt:
  broker: "tcp://localhost:1883"
  username: "user"
  password: "pass"
poll_interval: 2m
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o644))
	t.Setenv("CONFIG_FILE", cfgPath)
	// Clear env vars that would override
	t.Setenv("HOLOGRAM_API_KEY", "")
	t.Setenv("MQTT_BROKER", "")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "test-key", cfg.Hologram.APIKey)
	assert.Equal(t, "tcp://localhost:1883", cfg.MQTT.Broker)
	assert.Equal(t, "user", cfg.MQTT.Username)
	assert.Equal(t, "pass", cfg.MQTT.Password)
	assert.Equal(t, 2*time.Minute, cfg.PollInterval)
}

func TestEnvOverridesFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `
hologram:
  api_key: "file-key"
mqtt:
  broker: "tcp://file:1883"
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o644))
	t.Setenv("CONFIG_FILE", cfgPath)
	t.Setenv("HOLOGRAM_API_KEY", "env-key")
	t.Setenv("MQTT_BROKER", "tcp://env:1883")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "env-key", cfg.Hologram.APIKey)
	assert.Equal(t, "tcp://env:1883", cfg.MQTT.Broker)
}

func TestValidationRequiresAPIKey(t *testing.T) {
	t.Setenv("CONFIG_FILE", "/nonexistent/path")
	t.Setenv("HOLOGRAM_API_KEY", "")
	t.Setenv("MQTT_BROKER", "tcp://localhost:1883")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API key")
}

func TestValidationRequiresBroker(t *testing.T) {
	t.Setenv("CONFIG_FILE", "/nonexistent/path")
	t.Setenv("HOLOGRAM_API_KEY", "test-key")
	t.Setenv("MQTT_BROKER", "")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MQTT broker")
}

func TestPollIntervalMinimum(t *testing.T) {
	t.Setenv("CONFIG_FILE", "/nonexistent/path")
	t.Setenv("HOLOGRAM_API_KEY", "test-key")
	t.Setenv("MQTT_BROKER", "tcp://localhost:1883")
	t.Setenv("POLL_INTERVAL", "1s")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "poll interval")
}

func TestPollIntervalAsSeconds(t *testing.T) {
	t.Setenv("CONFIG_FILE", "/nonexistent/path")
	t.Setenv("HOLOGRAM_API_KEY", "test-key")
	t.Setenv("MQTT_BROKER", "tcp://localhost:1883")
	t.Setenv("POLL_INTERVAL", "60")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 60*time.Second, cfg.PollInterval)
}
