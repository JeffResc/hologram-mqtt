package bridge

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jeffresc/hologram-mqtt/internal/config"
	"github.com/jeffresc/hologram-mqtt/internal/hologram"
	"github.com/jeffresc/hologram-mqtt/internal/mqtt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type mockHologramClient struct {
	devices  []hologram.Device
	listErr  error
	stateErr error
	lastState string
	lastDeviceID int
}

func (m *mockHologramClient) ListDevices(_ context.Context) ([]hologram.Device, error) {
	return m.devices, m.listErr
}

func (m *mockHologramClient) SetDeviceState(_ context.Context, _ int, deviceID int, state string) error {
	m.lastDeviceID = deviceID
	m.lastState = state
	return m.stateErr
}

func testConfig() *config.Config {
	return &config.Config{
		Hologram: config.HologramConfig{APIKey: "test-key"},
		MQTT: config.MQTTConfig{
			Broker:      "tcp://localhost:1883",
			TopicPrefix: "hologram-mqtt",
			ClientID:    "test",
		},
		Discovery: config.DiscoveryConfig{
			Prefix:  "homeassistant",
			Enabled: true,
		},
		PollInterval: 1 * time.Minute,
	}
}

func testDevice() hologram.Device {
	return hologram.Device{
		ID:    42,
		OrgID: 10,
		Name:  "Test Device",
		IMEI:  "123456789012345",
		State: "LIVE",
	}
}

func TestBridgeInitialPoll(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	mockHolo := &mockHologramClient{
		devices: []hologram.Device{testDevice()},
	}

	b := New(testConfig(), mockHolo, mockMQTT, testLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = b.Run(ctx)

	// Should have subscribed to command topic
	require.Len(t, mockMQTT.Subscribed, 1)
	assert.Equal(t, "hologram-mqtt/device/+/switch/set", mockMQTT.Subscribed[0].Topic)

	// Should have published discovery configs + states
	assert.True(t, len(mockMQTT.Published) > 0, "should have published messages")

	// Verify discovery configs were published (retained)
	hasDiscovery := false
	for _, p := range mockMQTT.Published {
		if p.Topic == "homeassistant/sensor/hologram_42/state/config" {
			hasDiscovery = true
			assert.True(t, p.Retained)
			break
		}
	}
	assert.True(t, hasDiscovery, "should have published sensor discovery config")
}

func TestBridgeHandleCommand(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	mockHolo := &mockHologramClient{
		devices: []hologram.Device{testDevice()},
	}

	cfg := testConfig()
	b := New(cfg, mockHolo, mockMQTT, testLogger())

	// Seed known devices
	b.knownDevices[42] = testDevice()

	// Simulate switch OFF command (pause)
	b.handleCommand("hologram-mqtt/device/42/switch/set", []byte("OFF"))

	assert.Equal(t, 42, mockHolo.lastDeviceID)
	assert.Equal(t, "pause", mockHolo.lastState)

	// Verify state was published
	swState := mockMQTT.FindPublished("hologram-mqtt/device/42/switch/state")
	require.NotEmpty(t, swState)
	assert.Equal(t, "OFF", string(swState[len(swState)-1].Payload))

	// Clear and test ON command (resume)
	mockMQTT.Published = nil
	b.handleCommand("hologram-mqtt/device/42/switch/set", []byte("ON"))

	assert.Equal(t, "live", mockHolo.lastState)

	swState = mockMQTT.FindPublished("hologram-mqtt/device/42/switch/state")
	require.NotEmpty(t, swState)
	assert.Equal(t, "ON", string(swState[len(swState)-1].Payload))
}

func TestBridgeHandleCommandInvalidPayload(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	mockHolo := &mockHologramClient{}

	b := New(testConfig(), mockHolo, mockMQTT, testLogger())

	// Invalid payload should not call SetDeviceState
	b.handleCommand("hologram-mqtt/device/42/switch/set", []byte("INVALID"))
	assert.Equal(t, 0, mockHolo.lastDeviceID)
}

func TestBridgeDeviceRemoval(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	device := testDevice()
	mockHolo := &mockHologramClient{
		devices: []hologram.Device{device},
	}

	b := New(testConfig(), mockHolo, mockMQTT, testLogger())

	// First poll - device exists
	err := b.poll(context.Background())
	require.NoError(t, err)
	assert.Contains(t, b.knownDevices, 42)

	// Second poll - device removed
	mockHolo.devices = nil
	mockMQTT.Published = nil

	err = b.poll(context.Background())
	require.NoError(t, err)
	assert.NotContains(t, b.knownDevices, 42)

	// Should have published empty discovery configs (removal)
	hasRemoval := false
	for _, p := range mockMQTT.Published {
		if p.Topic == "homeassistant/sensor/hologram_42/state/config" && len(p.Payload) == 0 {
			hasRemoval = true
			break
		}
	}
	assert.True(t, hasRemoval, "should have published empty payload to remove discovery")
}

func TestBridgeAttributesJSON(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	connTime := int64(1712000000)
	device := hologram.Device{
		ID:                 42,
		OrgID:              10,
		Name:               "Test Device",
		State:              "LIVE",
		IMEI:               "123456789012345",
		SIMNumber:          "SIM-001",
		Carrier:            "T-Mobile",
		PhoneNumber:        "+15551234567",
		Plan:               &hologram.Plan{Name: "Pilot 1MB"},
		LastConnectionTime: &connTime,
		RecentSessionInfo:  &hologram.SessionInfo{BytesUp: 100, BytesDown: 200},
	}
	mockHolo := &mockHologramClient{devices: []hologram.Device{device}}

	b := New(testConfig(), mockHolo, mockMQTT, testLogger())
	err := b.poll(context.Background())
	require.NoError(t, err)

	attrs := mockMQTT.FindPublished("hologram-mqtt/device/42/attributes")
	require.NotEmpty(t, attrs)

	var a map[string]interface{}
	require.NoError(t, json.Unmarshal(attrs[0].Payload, &a))
	assert.Equal(t, "LIVE", a["state"])
	assert.Equal(t, "123456789012345", a["imei"])
	assert.Equal(t, "T-Mobile", a["carrier"])
	assert.Equal(t, "Pilot 1MB", a["plan"])
}
