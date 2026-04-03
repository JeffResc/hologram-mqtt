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
	assert.Equal(t, "hologram", cfg.MQTT.TopicPrefix)
	assert.Equal(t, "homeassistant", cfg.Discovery.Prefix)
	assert.True(t, cfg.Discovery.Enabled)
	assert.Equal(t, 5*time.Minute, cfg.PollInterval)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.True(t, cfg.Health.Enabled)
	assert.Equal(t, ":8080", cfg.Health.Addr)
	assert.False(t, cfg.MQTT.TLS.Enabled)
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

func TestTLSEnvVars(t *testing.T) {
	t.Setenv("CONFIG_FILE", "/nonexistent/path")
	t.Setenv("HOLOGRAM_API_KEY", "test-key")
	t.Setenv("MQTT_BROKER", "ssl://mqtt:8883")
	t.Setenv("MQTT_TLS_ENABLED", "true")
	t.Setenv("MQTT_TLS_CA_CERT", "/certs/ca.pem")
	t.Setenv("MQTT_TLS_CLIENT_CERT", "/certs/client.pem")
	t.Setenv("MQTT_TLS_CLIENT_KEY", "/certs/client-key.pem")
	t.Setenv("MQTT_TLS_SKIP_VERIFY", "false")

	cfg, err := Load()
	require.NoError(t, err)
	assert.True(t, cfg.MQTT.TLS.Enabled)
	assert.Equal(t, "/certs/ca.pem", cfg.MQTT.TLS.CACert)
	assert.Equal(t, "/certs/client.pem", cfg.MQTT.TLS.ClientCert)
	assert.Equal(t, "/certs/client-key.pem", cfg.MQTT.TLS.ClientKey)
	assert.False(t, cfg.MQTT.TLS.SkipVerify)
}

func TestTLSSkipVerifyTrue(t *testing.T) {
	t.Setenv("CONFIG_FILE", "/nonexistent/path")
	t.Setenv("HOLOGRAM_API_KEY", "test-key")
	t.Setenv("MQTT_BROKER", "ssl://mqtt:8883")
	t.Setenv("MQTT_TLS_ENABLED", "1")
	t.Setenv("MQTT_TLS_SKIP_VERIFY", "1")

	cfg, err := Load()
	require.NoError(t, err)
	assert.True(t, cfg.MQTT.TLS.Enabled)
	assert.True(t, cfg.MQTT.TLS.SkipVerify)
}

func TestTLSValidationClientCertWithoutKey(t *testing.T) {
	t.Setenv("CONFIG_FILE", "/nonexistent/path")
	t.Setenv("HOLOGRAM_API_KEY", "test-key")
	t.Setenv("MQTT_BROKER", "ssl://mqtt:8883")
	t.Setenv("MQTT_TLS_ENABLED", "true")
	t.Setenv("MQTT_TLS_CLIENT_CERT", "/certs/client.pem")
	t.Setenv("MQTT_TLS_CLIENT_KEY", "")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client key")
}

func TestTLSValidationClientKeyWithoutCert(t *testing.T) {
	t.Setenv("CONFIG_FILE", "/nonexistent/path")
	t.Setenv("HOLOGRAM_API_KEY", "test-key")
	t.Setenv("MQTT_BROKER", "ssl://mqtt:8883")
	t.Setenv("MQTT_TLS_ENABLED", "true")
	t.Setenv("MQTT_TLS_CLIENT_CERT", "")
	t.Setenv("MQTT_TLS_CLIENT_KEY", "/certs/client-key.pem")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client cert")
}

func TestTLSFromYAML(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	content := `
hologram:
  api_key: "test-key"
mqtt:
  broker: "ssl://mqtt:8883"
  tls:
    enabled: true
    ca_cert: "/certs/ca.pem"
    skip_verify: true
`
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o644))
	t.Setenv("CONFIG_FILE", cfgPath)
	t.Setenv("HOLOGRAM_API_KEY", "")
	t.Setenv("MQTT_BROKER", "")
	t.Setenv("MQTT_TLS_ENABLED", "")
	t.Setenv("MQTT_TLS_CA_CERT", "")
	t.Setenv("MQTT_TLS_SKIP_VERIFY", "")

	cfg, err := Load()
	require.NoError(t, err)
	assert.True(t, cfg.MQTT.TLS.Enabled)
	assert.Equal(t, "/certs/ca.pem", cfg.MQTT.TLS.CACert)
	assert.True(t, cfg.MQTT.TLS.SkipVerify)
}

func TestHealthEnvVars(t *testing.T) {
	t.Setenv("CONFIG_FILE", "/nonexistent/path")
	t.Setenv("HOLOGRAM_API_KEY", "test-key")
	t.Setenv("MQTT_BROKER", "tcp://localhost:1883")
	t.Setenv("HEALTH_ENABLED", "false")
	t.Setenv("HEALTH_ADDR", ":9090")

	cfg, err := Load()
	require.NoError(t, err)
	assert.False(t, cfg.Health.Enabled)
	assert.Equal(t, ":9090", cfg.Health.Addr)
}
