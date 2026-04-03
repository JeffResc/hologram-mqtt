package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
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
	mockMQTT.ClearPublished()
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
	mockHolo.mu.Lock()
	mockHolo.devices = nil
	mockHolo.mu.Unlock()
	mockMQTT.ClearPublished()

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

// --- Edge case tests ---

func TestBridgeEmptyDeviceList(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	mockHolo := &mockHologramClient{devices: []hologram.Device{}}

	b := New(testConfig(), mockHolo, mockMQTT, testLogger())

	err := b.poll(context.Background())
	require.NoError(t, err)
	assert.Empty(t, b.knownDevices)
	// Should not have published any discovery configs
	for _, p := range mockMQTT.Published {
		assert.False(t, len(p.Topic) > 0 && p.Topic[len(p.Topic)-7:] == "/config",
			"should not publish discovery for empty device list")
	}
}

func TestBridgeDeviceWithNilFields(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	// Device with all optional pointer fields nil
	device := hologram.Device{
		ID:                 99,
		OrgID:              10,
		Name:               "Minimal Device",
		State:              "PAUSED",
		Plan:               nil,
		LastConnectionTime: nil,
		RecentSessionInfo:  nil,
	}
	mockHolo := &mockHologramClient{devices: []hologram.Device{device}}

	b := New(testConfig(), mockHolo, mockMQTT, testLogger())
	err := b.poll(context.Background())
	require.NoError(t, err)

	// Verify attributes published successfully with zero/empty values
	attrs := mockMQTT.FindPublished("hologram-mqtt/device/99/attributes")
	require.NotEmpty(t, attrs)

	var a map[string]interface{}
	require.NoError(t, json.Unmarshal(attrs[0].Payload, &a))
	assert.Equal(t, "PAUSED", a["state"])
	assert.Equal(t, "", a["plan"])
	assert.Equal(t, "", a["last_connection"])
	assert.Equal(t, float64(0), a["data_up"])
	assert.Equal(t, float64(0), a["data_down"])

	// Binary sensor should show OFF for PAUSED
	conn := mockMQTT.FindPublished("hologram-mqtt/device/99/connectivity")
	require.NotEmpty(t, conn)
	assert.Equal(t, "OFF", string(conn[0].Payload))
}

func TestBridgeListDevicesError(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	mockHolo := &mockHologramClient{
		listErr: errors.New("network timeout"),
	}

	b := New(testConfig(), mockHolo, mockMQTT, testLogger())

	err := b.poll(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "network timeout")
	// knownDevices should remain empty
	assert.Empty(t, b.knownDevices)
}

func TestBridgeSetDeviceStateError(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	mockHolo := &mockHologramClient{
		stateErr: errors.New("API error"),
	}

	b := New(testConfig(), mockHolo, mockMQTT, testLogger())
	b.knownDevices[42] = testDevice()

	// Command should not crash, state should remain unchanged
	b.handleCommand("hologram-mqtt/device/42/switch/set", []byte("OFF"))

	// Should not have published any state updates
	swState := mockMQTT.FindPublished("hologram-mqtt/device/42/switch/state")
	assert.Empty(t, swState)

	// Device state should still be LIVE
	b.mu.RLock()
	d := b.knownDevices[42]
	b.mu.RUnlock()
	assert.Equal(t, "LIVE", d.State)
}

func TestBridgePublishErrorDuringPoll(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	mockMQTT.PublishErr = errors.New("broker unreachable")

	mockHolo := &mockHologramClient{
		devices: []hologram.Device{testDevice()},
	}

	b := New(testConfig(), mockHolo, mockMQTT, testLogger())

	// poll should not panic; discovery publish will fail but poll continues
	err := b.poll(context.Background())
	// The error is logged but poll returns nil (states failure is logged, not returned)
	require.NoError(t, err)
}

func TestBridgeHandleCommandUnknownDevice(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	mockHolo := &mockHologramClient{}

	b := New(testConfig(), mockHolo, mockMQTT, testLogger())
	// knownDevices is empty

	b.handleCommand("hologram-mqtt/device/999/switch/set", []byte("ON"))
	// Should not call SetDeviceState
	assert.Equal(t, 0, mockHolo.stateCalls)
}

func TestBridgeHandleCommandInvalidDeviceID(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	mockHolo := &mockHologramClient{}

	b := New(testConfig(), mockHolo, mockMQTT, testLogger())

	b.handleCommand("hologram-mqtt/device/notanumber/switch/set", []byte("ON"))
	assert.Equal(t, 0, mockHolo.stateCalls)
}

func TestBridgeHandleCommandShortTopic(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	mockHolo := &mockHologramClient{}

	b := New(testConfig(), mockHolo, mockMQTT, testLogger())

	b.handleCommand("too/short", []byte("ON"))
	assert.Equal(t, 0, mockHolo.stateCalls)
}

func TestBridgeConcurrentCommands(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	mockHolo := &mockHologramClient{}

	b := New(testConfig(), mockHolo, mockMQTT, testLogger())

	// Seed two devices
	b.knownDevices[1] = hologram.Device{ID: 1, OrgID: 10, Name: "D1", State: "LIVE"}
	b.knownDevices[2] = hologram.Device{ID: 2, OrgID: 10, Name: "D2", State: "LIVE"}

	var wg sync.WaitGroup
	// Fire many concurrent commands to test mutex safety
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			b.handleCommand("hologram-mqtt/device/1/switch/set", []byte("OFF"))
		}()
		go func() {
			defer wg.Done()
			b.handleCommand("hologram-mqtt/device/2/switch/set", []byte("ON"))
		}()
	}
	wg.Wait()

	// Verify no panics occurred and state calls were made
	mockHolo.mu.Lock()
	assert.Equal(t, 100, mockHolo.stateCalls)
	mockHolo.mu.Unlock()
}

func TestBridgeConcurrentPollAndCommand(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	device := testDevice()
	mockHolo := &mockHologramClient{
		devices: []hologram.Device{device},
	}

	b := New(testConfig(), mockHolo, mockMQTT, testLogger())

	// Seed known device
	b.knownDevices[42] = device

	var wg sync.WaitGroup
	// Run poll and commands concurrently
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			b.poll(context.Background()) //nolint:errcheck
		}()
		go func() {
			defer wg.Done()
			b.handleCommand("hologram-mqtt/device/42/switch/set", []byte("OFF"))
		}()
	}
	wg.Wait()
	// Success = no data race detected (run with -race)
}

func TestBridgeDiscoveryDisabled(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	mockHolo := &mockHologramClient{
		devices: []hologram.Device{testDevice()},
	}

	cfg := testConfig()
	cfg.Discovery.Enabled = false
	b := New(cfg, mockHolo, mockMQTT, testLogger())

	err := b.poll(context.Background())
	require.NoError(t, err)

	// Should not have published any discovery configs
	for _, p := range mockMQTT.Published {
		if len(p.Topic) >= 7 {
			assert.NotEqual(t, "/config", p.Topic[len(p.Topic)-7:],
				"should not publish discovery config when disabled: %s", p.Topic)
		}
	}

	// Should still have published state topics
	attrs := mockMQTT.FindPublished("hologram-mqtt/device/42/attributes")
	assert.NotEmpty(t, attrs, "should still publish state even with discovery disabled")
}

func TestBridgeNewDeviceAppearsOnSecondPoll(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	device1 := testDevice()
	mockHolo := &mockHologramClient{
		devices: []hologram.Device{device1},
	}

	b := New(testConfig(), mockHolo, mockMQTT, testLogger())

	// First poll
	err := b.poll(context.Background())
	require.NoError(t, err)
	assert.Len(t, b.knownDevices, 1)

	// Add a second device
	device2 := hologram.Device{ID: 99, OrgID: 10, Name: "New Device", State: "LIVE"}
	mockHolo.mu.Lock()
	mockHolo.devices = []hologram.Device{device1, device2}
	mockHolo.mu.Unlock()
	mockMQTT.ClearPublished()

	// Second poll
	err = b.poll(context.Background())
	require.NoError(t, err)
	assert.Len(t, b.knownDevices, 2)

	// Should have published discovery for the new device
	hasNewDiscovery := false
	for _, p := range mockMQTT.Published {
		if p.Topic == "homeassistant/sensor/hologram_99/state/config" {
			hasNewDiscovery = true
			break
		}
	}
	assert.True(t, hasNewDiscovery, "should have published discovery for new device")
}

func TestBridgeDEADDeviceSwitchState(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	device := testDevice()
	device.State = "DEAD"
	mockHolo := &mockHologramClient{devices: []hologram.Device{device}}

	b := New(testConfig(), mockHolo, mockMQTT, testLogger())
	err := b.poll(context.Background())
	require.NoError(t, err)

	sw := mockMQTT.FindPublished("hologram-mqtt/device/42/switch/state")
	require.NotEmpty(t, sw)
	assert.Equal(t, "OFF", string(sw[0].Payload))

	conn := mockMQTT.FindPublished("hologram-mqtt/device/42/connectivity")
	require.NotEmpty(t, conn)
	assert.Equal(t, "OFF", string(conn[0].Payload))
}

// --- Health check tests ---

func TestHealthHandlerHealthy(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	mockMQTT.Connected = true
	mockHolo := &mockHologramClient{}

	b := New(testConfig(), mockHolo, mockMQTT, testLogger())

	handler := b.HealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}

func TestHealthHandlerUnhealthy(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	mockMQTT.Connected = false
	mockHolo := &mockHologramClient{}

	b := New(testConfig(), mockHolo, mockMQTT, testLogger())

	handler := b.HealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Equal(t, "unhealthy", rec.Body.String())
}
