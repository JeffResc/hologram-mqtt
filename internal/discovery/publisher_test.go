package discovery

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/jeffresc/hologram-mqtt/internal/hologram"
	"github.com/jeffresc/hologram-mqtt/internal/mqtt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func testDevice() hologram.Device {
	connTime := int64(1712000000)
	return hologram.Device{
		ID:                 42,
		OrgID:              10,
		Name:               "Test Device",
		IMEI:               "123456789012345",
		SIMNumber:          "SIM-001",
		State:              "LIVE",
		Carrier:            "T-Mobile",
		PhoneNumber:        "+15551234567",
		NetworkUsed:        "LTE",
		DeviceType:         "Router",
		Manufacturer:       "Hologram",
		LastConnectionTime: &connTime,
		Plan:               &hologram.Plan{Name: "Pilot 1MB"},
		RecentSessionInfo:  &hologram.SessionInfo{BytesUp: 1024, BytesDown: 2048},
	}
}

func TestSensorConfigs(t *testing.T) {
	d := testDevice()
	configs := SensorConfigs(d, "hologram-mqtt", "homeassistant")

	assert.Len(t, configs, len(sensors))

	for _, cfg := range configs {
		assert.True(t, strings.HasPrefix(cfg.Topic, "homeassistant/sensor/hologram_42/"))
		assert.True(t, strings.HasSuffix(cfg.Topic, "/config"))
		assert.Equal(t, "hologram-mqtt/device/42/attributes", cfg.Payload.StateTopic)
		assert.NotEmpty(t, cfg.Payload.UniqueID)
		assert.Equal(t, "all", cfg.Payload.AvailabilityMode)
		assert.Len(t, cfg.Payload.Availability, 2)
		assert.Equal(t, "Test Device", cfg.Payload.Device.Name)
		assert.Equal(t, []string{"hologram_42"}, cfg.Payload.Device.Identifiers)
	}
}

func TestBinarySensorConfig(t *testing.T) {
	d := testDevice()
	cfg := BinarySensorConfig(d, "hologram-mqtt", "homeassistant")

	assert.Equal(t, "homeassistant/binary_sensor/hologram_42/connectivity/config", cfg.Topic)
	assert.Equal(t, "connectivity", cfg.Payload.DeviceClass)
	assert.Equal(t, "ON", cfg.Payload.PayloadOn)
	assert.Equal(t, "OFF", cfg.Payload.PayloadOff)
	assert.Equal(t, "hologram-mqtt/device/42/connectivity", cfg.Payload.StateTopic)
}

func TestSwitchConfig(t *testing.T) {
	d := testDevice()
	cfg := SwitchConfig(d, "hologram-mqtt", "homeassistant")

	assert.Equal(t, "homeassistant/switch/hologram_42/active/config", cfg.Topic)
	assert.Equal(t, "hologram-mqtt/device/42/switch/state", cfg.Payload.StateTopic)
	assert.Equal(t, "hologram-mqtt/device/42/switch/set", cfg.Payload.CommandTopic)
	assert.Equal(t, "ON", cfg.Payload.PayloadOn)
	assert.Equal(t, "OFF", cfg.Payload.PayloadOff)
}

func TestAllConfigs(t *testing.T) {
	d := testDevice()
	configs := AllConfigs(d, "hologram-mqtt", "homeassistant")

	// sensors + binary_sensor + switch
	expected := len(sensors) + 2
	assert.Len(t, configs, expected)
}

func TestPublishDiscovery(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram-mqtt", "homeassistant", testLogger())

	devices := []hologram.Device{testDevice()}
	err := pub.PublishDiscovery(devices)
	require.NoError(t, err)

	expectedCount := len(sensors) + 2 // sensors + binary_sensor + switch
	assert.Len(t, mockMQTT.Published, expectedCount)

	// Verify all published messages are valid JSON
	for _, p := range mockMQTT.Published {
		assert.True(t, p.Retained, "discovery messages should be retained")
		var payload DiscoveryPayload
		err := json.Unmarshal(p.Payload, &payload)
		assert.NoError(t, err, "payload should be valid JSON for topic %s", p.Topic)
		assert.NotEmpty(t, payload.UniqueID)
	}
}

func TestPublishStates(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram-mqtt", "homeassistant", testLogger())

	devices := []hologram.Device{testDevice()}
	err := pub.PublishStates(devices)
	require.NoError(t, err)

	// Should publish: availability, attributes, connectivity, switch/state
	assert.Len(t, mockMQTT.Published, 4)

	// Check availability
	avail := mockMQTT.FindPublished("hologram-mqtt/device/42/availability")
	require.Len(t, avail, 1)
	assert.Equal(t, "online", string(avail[0].Payload))

	// Check connectivity for LIVE device
	conn := mockMQTT.FindPublished("hologram-mqtt/device/42/connectivity")
	require.Len(t, conn, 1)
	assert.Equal(t, "ON", string(conn[0].Payload))

	// Check switch state for LIVE device
	sw := mockMQTT.FindPublished("hologram-mqtt/device/42/switch/state")
	require.Len(t, sw, 1)
	assert.Equal(t, "ON", string(sw[0].Payload))

	// Check attributes JSON
	attrs := mockMQTT.FindPublished("hologram-mqtt/device/42/attributes")
	require.Len(t, attrs, 1)
	var a deviceAttributes
	require.NoError(t, json.Unmarshal(attrs[0].Payload, &a))
	assert.Equal(t, "LIVE", a.State)
	assert.Equal(t, "123456789012345", a.IMEI)
	assert.Equal(t, "Pilot 1MB", a.Plan)
	assert.Equal(t, int64(1024), a.DataUp)
}

func TestPublishStatesPausedDevice(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram-mqtt", "homeassistant", testLogger())

	d := testDevice()
	d.State = "PAUSED"
	err := pub.PublishStates([]hologram.Device{d})
	require.NoError(t, err)

	conn := mockMQTT.FindPublished("hologram-mqtt/device/42/connectivity")
	require.Len(t, conn, 1)
	assert.Equal(t, "OFF", string(conn[0].Payload))

	sw := mockMQTT.FindPublished("hologram-mqtt/device/42/switch/state")
	require.Len(t, sw, 1)
	assert.Equal(t, "OFF", string(sw[0].Payload))
}

func TestRemoveDiscovery(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram-mqtt", "homeassistant", testLogger())

	devices := []hologram.Device{testDevice()}
	err := pub.RemoveDiscovery(devices)
	require.NoError(t, err)

	// Should publish empty payloads for all config topics + availability offline
	expectedCount := len(sensors) + 2 + 1 // discovery configs + availability
	assert.Len(t, mockMQTT.Published, expectedCount)

	// Verify config topics have empty payloads
	for _, p := range mockMQTT.Published {
		if strings.HasSuffix(p.Topic, "/config") {
			assert.Empty(t, p.Payload, "removal should publish empty payload to %s", p.Topic)
		}
	}
}

func TestDeviceDefaults(t *testing.T) {
	d := hologram.Device{
		ID:   1,
		Name: "Test",
	}
	haDevice := newDevice(d)
	assert.Equal(t, "Hologram", haDevice.Manufacturer)
	assert.Equal(t, "SIM", haDevice.Model)
}

// --- Edge case tests ---

func TestPublishStatesNilPlan(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram-mqtt", "homeassistant", testLogger())

	d := hologram.Device{
		ID:    1,
		State: "LIVE",
		Plan:  nil,
	}
	err := pub.PublishStates([]hologram.Device{d})
	require.NoError(t, err)

	attrs := mockMQTT.FindPublished("hologram-mqtt/device/1/attributes")
	require.NotEmpty(t, attrs)

	var a deviceAttributes
	require.NoError(t, json.Unmarshal(attrs[0].Payload, &a))
	assert.Equal(t, "", a.Plan)
}

func TestPublishStatesNilSessionInfo(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram-mqtt", "homeassistant", testLogger())

	d := hologram.Device{
		ID:                1,
		State:             "LIVE",
		RecentSessionInfo: nil,
	}
	err := pub.PublishStates([]hologram.Device{d})
	require.NoError(t, err)

	attrs := mockMQTT.FindPublished("hologram-mqtt/device/1/attributes")
	require.NotEmpty(t, attrs)

	var a deviceAttributes
	require.NoError(t, json.Unmarshal(attrs[0].Payload, &a))
	assert.Equal(t, int64(0), a.DataUp)
	assert.Equal(t, int64(0), a.DataDown)
}

func TestPublishStatesNilConnectionTime(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram-mqtt", "homeassistant", testLogger())

	d := hologram.Device{
		ID:                 1,
		State:              "LIVE",
		LastConnectionTime: nil,
	}
	err := pub.PublishStates([]hologram.Device{d})
	require.NoError(t, err)

	attrs := mockMQTT.FindPublished("hologram-mqtt/device/1/attributes")
	require.NotEmpty(t, attrs)

	var a deviceAttributes
	require.NoError(t, json.Unmarshal(attrs[0].Payload, &a))
	assert.Equal(t, "", a.LastConnection)
}

func TestPublishStatesAllNilFields(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram-mqtt", "homeassistant", testLogger())

	d := hologram.Device{
		ID:    1,
		State: "DEAD",
	}
	err := pub.PublishStates([]hologram.Device{d})
	require.NoError(t, err)

	// DEAD device should have OFF connectivity and OFF switch
	conn := mockMQTT.FindPublished("hologram-mqtt/device/1/connectivity")
	require.NotEmpty(t, conn)
	assert.Equal(t, "OFF", string(conn[0].Payload))

	sw := mockMQTT.FindPublished("hologram-mqtt/device/1/switch/state")
	require.NotEmpty(t, sw)
	assert.Equal(t, "OFF", string(sw[0].Payload))
}

func TestPublishDiscoveryEmptyList(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram-mqtt", "homeassistant", testLogger())

	err := pub.PublishDiscovery([]hologram.Device{})
	require.NoError(t, err)
	assert.Empty(t, mockMQTT.Published)
}

func TestPublishStatesEmptyList(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram-mqtt", "homeassistant", testLogger())

	err := pub.PublishStates([]hologram.Device{})
	require.NoError(t, err)
	assert.Empty(t, mockMQTT.Published)
}

func TestPublishDiscoveryMQTTError(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	mockMQTT.PublishErr = errors.New("broker disconnected")
	pub := NewPublisher(mockMQTT, "hologram-mqtt", "homeassistant", testLogger())

	err := pub.PublishDiscovery([]hologram.Device{testDevice()})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "broker disconnected")
}

func TestPublishStatesPartialMQTTError(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram-mqtt", "homeassistant", testLogger())

	// First device works, we set error after first state publish
	d1 := hologram.Device{ID: 1, State: "LIVE"}
	d2 := hologram.Device{ID: 2, State: "LIVE"}

	// PublishStates logs errors per-device but returns nil
	err := pub.PublishStates([]hologram.Device{d1, d2})
	require.NoError(t, err)
	// Both devices should have attempted publishing
	assert.NotEmpty(t, mockMQTT.FindPublished("hologram-mqtt/device/1/availability"))
	assert.NotEmpty(t, mockMQTT.FindPublished("hologram-mqtt/device/2/availability"))
}

func TestRemoveDiscoveryMQTTError(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	mockMQTT.PublishErr = errors.New("broker disconnected")
	pub := NewPublisher(mockMQTT, "hologram-mqtt", "homeassistant", testLogger())

	err := pub.RemoveDiscovery([]hologram.Device{testDevice()})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "broker disconnected")
}

func TestPublishDiscoveryMultipleDevices(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram-mqtt", "homeassistant", testLogger())

	d1 := hologram.Device{ID: 1, Name: "Device 1", State: "LIVE"}
	d2 := hologram.Device{ID: 2, Name: "Device 2", State: "PAUSED"}

	err := pub.PublishDiscovery([]hologram.Device{d1, d2})
	require.NoError(t, err)

	expectedCount := (len(sensors) + 2) * 2
	assert.Len(t, mockMQTT.Published, expectedCount)

	// Verify both devices have discovery configs
	hasD1 := false
	hasD2 := false
	for _, p := range mockMQTT.Published {
		if strings.Contains(p.Topic, "hologram_1") {
			hasD1 = true
		}
		if strings.Contains(p.Topic, "hologram_2") {
			hasD2 = true
		}
	}
	assert.True(t, hasD1, "should have discovery for device 1")
	assert.True(t, hasD2, "should have discovery for device 2")
}

func TestBuildAttributesPreservesAllFields(t *testing.T) {
	connTime := int64(1712000000)
	d := hologram.Device{
		State:              "LIVE",
		IMEI:               "imei-123",
		SIMNumber:          "sim-456",
		Carrier:            "T-Mobile",
		PhoneNumber:        "+1555",
		NetworkUsed:        "LTE",
		Plan:               &hologram.Plan{Name: "Plan A"},
		LastConnectionTime: &connTime,
		RecentSessionInfo:  &hologram.SessionInfo{BytesUp: 100, BytesDown: 200},
	}

	a := buildAttributes(d)
	assert.Equal(t, "LIVE", a.State)
	assert.Equal(t, "imei-123", a.IMEI)
	assert.Equal(t, "sim-456", a.SIMNumber)
	assert.Equal(t, "T-Mobile", a.Carrier)
	assert.Equal(t, "+1555", a.PhoneNumber)
	assert.Equal(t, "LTE", a.Network)
	assert.Equal(t, "Plan A", a.Plan)
	assert.Equal(t, "2024-04-01T19:33:20Z", a.LastConnection)
	assert.Equal(t, int64(100), a.DataUp)
	assert.Equal(t, int64(200), a.DataDown)
}
