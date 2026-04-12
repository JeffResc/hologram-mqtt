//go:build integration

package bridge_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jeffresc/hologram-mqtt/internal/bridge"
	"github.com/jeffresc/hologram-mqtt/internal/config"
	"github.com/jeffresc/hologram-mqtt/internal/hologram"
	"github.com/jeffresc/hologram-mqtt/internal/mqtt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startMosquitto(t *testing.T) (string, func()) {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "eclipse-mosquitto:2",
		ExposedPorts: []string{"1883/tcp"},
		Files: []testcontainers.ContainerFile{
			{
				Reader:            strings.NewReader("listener 1883\nallow_anonymous true\n"),
				ContainerFilePath: "/mosquitto/config/mosquitto.conf",
			},
		},
		WaitingFor: wait.ForListeningPort("1883/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "1883/tcp")
	require.NoError(t, err)

	broker := fmt.Sprintf("tcp://%s:%s", host, port.Port())

	cleanup := func() {
		_ = container.Terminate(ctx)
	}

	return broker, cleanup
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

type mockHologramClient struct {
	mu           sync.Mutex
	devices      []hologram.Device
	listErr      error
	stateErr     error
	lastState    string
	lastDeviceID int
	stateCalls   int
}

func (m *mockHologramClient) ListDevices(_ context.Context) ([]hologram.Device, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.devices, m.listErr
}

func (m *mockHologramClient) SetDeviceState(_ context.Context, _ int, deviceID int, state string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastDeviceID = deviceID
	m.lastState = state
	m.stateCalls++
	return m.stateErr
}

func testDevice() hologram.Device {
	return hologram.Device{
		ID:    42,
		OrgID: 10,
		Name:  "Integration Test Device",
		IMEI:  "123456789012345",
		Links: &hologram.DeviceLinks{
			Cellular: []hologram.CellularLink{{
				ID:    100,
				State: "LIVE",
			}},
		},
	}
}

func TestIntegrationBridgePollAndPublish(t *testing.T) {
	broker, cleanup := startMosquitto(t)
	defer cleanup()

	// Create the bridge MQTT client
	bridgeClient, err := mqtt.NewClient(mqtt.ClientConfig{
		Broker:      broker,
		ClientID:    "bridge",
		TopicPrefix: "hologram",
	}, testLogger())
	require.NoError(t, err)
	defer bridgeClient.Disconnect()

	// Create a separate observer client to verify published messages
	observer, err := mqtt.NewClient(mqtt.ClientConfig{
		Broker:      broker,
		ClientID:    "observer",
		TopicPrefix: "observer",
	}, testLogger())
	require.NoError(t, err)
	defer observer.Disconnect()

	var mu sync.Mutex
	messages := make(map[string]string)

	// Subscribe to all hologram topics
	err = observer.Subscribe("hologram/#", 1, func(topic string, payload []byte) {
		mu.Lock()
		messages[topic] = string(payload)
		mu.Unlock()
	})
	require.NoError(t, err)
	time.Sleep(200 * time.Millisecond)

	cfg := &config.Config{
		Hologram: config.HologramConfig{APIKey: "test-key"},
		MQTT: config.MQTTConfig{
			Broker:      broker,
			TopicPrefix: "hologram",
			ClientID:    "bridge",
		},
		Discovery: config.DiscoveryConfig{
			Prefix:  "homeassistant",
			Enabled: true,
		},
		PollInterval: 30 * time.Second,
	}

	mockHolo := &mockHologramClient{
		devices: []hologram.Device{testDevice()},
	}

	b := bridge.New(cfg, mockHolo, bridgeClient, testLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		_ = b.Run(ctx)
	}()

	// Wait for messages to be published
	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		_, hasState := messages["hologram/device/42/switch/state"]
		_, hasAttrs := messages["hologram/device/42/attributes"]
		return hasState && hasAttrs
	}, 5*time.Second, 100*time.Millisecond, "expected state and attributes to be published")

	// Verify the switch state reflects LIVE → ON
	mu.Lock()
	assert.Equal(t, "ON", messages["hologram/device/42/switch/state"])
	mu.Unlock()

	cancel()
}

func TestIntegrationBridgeCommandRoundTrip(t *testing.T) {
	broker, cleanup := startMosquitto(t)
	defer cleanup()

	// Create the bridge MQTT client
	bridgeClient, err := mqtt.NewClient(mqtt.ClientConfig{
		Broker:      broker,
		ClientID:    "bridge-cmd",
		TopicPrefix: "hologram",
	}, testLogger())
	require.NoError(t, err)
	defer bridgeClient.Disconnect()

	// Create a commander client to send commands and observe responses
	commander, err := mqtt.NewClient(mqtt.ClientConfig{
		Broker:      broker,
		ClientID:    "commander",
		TopicPrefix: "commander",
	}, testLogger())
	require.NoError(t, err)
	defer commander.Disconnect()

	var mu sync.Mutex
	var switchState string

	// Observe switch state updates
	err = commander.Subscribe("hologram/device/42/switch/state", 1, func(_ string, payload []byte) {
		mu.Lock()
		switchState = string(payload)
		mu.Unlock()
	})
	require.NoError(t, err)
	time.Sleep(200 * time.Millisecond)

	cfg := &config.Config{
		Hologram: config.HologramConfig{APIKey: "test-key"},
		MQTT: config.MQTTConfig{
			Broker:      broker,
			TopicPrefix: "hologram",
			ClientID:    "bridge-cmd",
		},
		Discovery: config.DiscoveryConfig{
			Prefix:  "homeassistant",
			Enabled: true,
		},
		PollInterval: 30 * time.Second,
	}

	mockHolo := &mockHologramClient{
		devices: []hologram.Device{testDevice()},
	}

	b := bridge.New(cfg, mockHolo, bridgeClient, testLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		_ = b.Run(ctx)
	}()

	// Wait for initial state to be published (device is LIVE → ON)
	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return switchState == "ON"
	}, 5*time.Second, 100*time.Millisecond, "expected initial switch state ON")

	// Send OFF command
	err = commander.Publish("hologram/device/42/switch/set", 1, false, []byte("OFF"))
	require.NoError(t, err)

	// Verify the bridge called SetDeviceState and published the updated state
	assert.Eventually(t, func() bool {
		mockHolo.mu.Lock()
		defer mockHolo.mu.Unlock()
		return mockHolo.lastState == "pause" && mockHolo.lastDeviceID == 42
	}, 5*time.Second, 100*time.Millisecond, "expected SetDeviceState to be called with pause")

	assert.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return switchState == "OFF"
	}, 5*time.Second, 100*time.Millisecond, "expected switch state to update to OFF")

	cancel()
}
