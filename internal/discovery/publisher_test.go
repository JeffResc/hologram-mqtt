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
	return hologram.Device{
		ID:           42,
		OrgID:        10,
		Name:         "Test Device",
		IMEI:         "123456789012345",
		Model:        "Router",
		Manufacturer: "Hologram",
		Links: &hologram.DeviceLinks{
			Cellular: []hologram.CellularLink{{
				ID:                  100,
				SIM:                 "89464278206108944117",
				IMSI:                240422610894411,
				MSISDN:              "+15551234567",
				State:               "LIVE",
				CarrierID:           "13",
				LastConnectTime:     "2026-04-03 11:12:16",
				LastNetworkUsed:     "AT&T Mobility",
				CurBillingDataUsed:  28580756,
				LastBillingDataUsed: 649325609,
				EID:                 "89044045117727494800000294996158",
				ProfileState:        "ENABLED",
				Plan:                &hologram.Plan{Name: "Global G3 Standard Flat Rate"},
			}},
		},
		LastSession: &hologram.LastSession{
			NetworkName:  "AT&T Mobility",
			RadioTech:    "LTE",
			Active:       true,
			SessionBegin: "2026-04-03 11:12:16",
		},
	}
}

func TestSensorConfigs(t *testing.T) {
	d := testDevice()
	configs := SensorConfigs(d, "hologram", "homeassistant")

	assert.Len(t, configs, len(sensors))

	for _, cfg := range configs {
		assert.True(t, strings.HasPrefix(cfg.Topic, "homeassistant/sensor/hologram_42/"))
		assert.True(t, strings.HasSuffix(cfg.Topic, "/config"))
		assert.Equal(t, "hologram/device/42/attributes", cfg.Payload.StateTopic)
		assert.NotEmpty(t, cfg.Payload.UniqueID)
		assert.Equal(t, "all", cfg.Payload.AvailabilityMode)
		assert.Len(t, cfg.Payload.Availability, 2)
		assert.Equal(t, "Test Device", cfg.Payload.Device.Name)
		assert.Equal(t, []string{"hologram_42"}, cfg.Payload.Device.Identifiers)
	}
}

func TestBinarySensorConfig(t *testing.T) {
	d := testDevice()
	cfg := BinarySensorConfig(d, "hologram", "homeassistant")

	assert.Equal(t, "homeassistant/binary_sensor/hologram_42/connectivity/config", cfg.Topic)
	assert.Equal(t, "connectivity", cfg.Payload.DeviceClass)
	assert.Equal(t, "ON", cfg.Payload.PayloadOn)
	assert.Equal(t, "OFF", cfg.Payload.PayloadOff)
	assert.Equal(t, "hologram/device/42/connectivity", cfg.Payload.StateTopic)
}

func TestSwitchConfig(t *testing.T) {
	d := testDevice()
	cfg := SwitchConfig(d, "hologram", "homeassistant")

	assert.Equal(t, "homeassistant/switch/hologram_42/active/config", cfg.Topic)
	assert.Equal(t, "hologram/device/42/switch/state", cfg.Payload.StateTopic)
	assert.Equal(t, "hologram/device/42/switch/set", cfg.Payload.CommandTopic)
	assert.Equal(t, "ON", cfg.Payload.PayloadOn)
	assert.Equal(t, "OFF", cfg.Payload.PayloadOff)
}

func TestAllConfigs(t *testing.T) {
	d := testDevice()
	configs := AllConfigs(d, "hologram", "homeassistant")

	// sensors + binary_sensor + switch
	expected := len(sensors) + 2
	assert.Len(t, configs, expected)
}

func TestPublishDiscovery(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram", "homeassistant", testLogger())

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
	pub := NewPublisher(mockMQTT, "hologram", "homeassistant", testLogger())

	devices := []hologram.Device{testDevice()}
	err := pub.PublishStates(devices)
	require.NoError(t, err)

	// Should publish: availability, attributes, connectivity, switch/state
	assert.Len(t, mockMQTT.Published, 4)

	// Check availability
	avail := mockMQTT.FindPublished("hologram/device/42/availability")
	require.Len(t, avail, 1)
	assert.Equal(t, "online", string(avail[0].Payload))

	// Check connectivity for LIVE device
	conn := mockMQTT.FindPublished("hologram/device/42/connectivity")
	require.Len(t, conn, 1)
	assert.Equal(t, "ON", string(conn[0].Payload))

	// Check switch state for LIVE device
	sw := mockMQTT.FindPublished("hologram/device/42/switch/state")
	require.Len(t, sw, 1)
	assert.Equal(t, "ON", string(sw[0].Payload))

	// Check attributes JSON
	attrs := mockMQTT.FindPublished("hologram/device/42/attributes")
	require.Len(t, attrs, 1)
	var a deviceAttributes
	require.NoError(t, json.Unmarshal(attrs[0].Payload, &a))
	assert.Equal(t, "LIVE", a.State)
	assert.Equal(t, "123456789012345", a.IMEI)
	assert.Equal(t, "Global G3 Standard Flat Rate", a.Plan)
	assert.Equal(t, "AT&T Mobility", a.Network)
	assert.Equal(t, "LTE", a.RadioTech)
	assert.Equal(t, "89464278206108944117", a.ICCID)
	assert.Equal(t, "+15551234567", a.PhoneNumber)
	assert.Equal(t, int64(28580756), a.CurBillingDataUsed)
	assert.True(t, a.SessionActive)
}

func TestPublishStatesPausedDevice(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram", "homeassistant", testLogger())

	d := testDevice()
	d.Links.Cellular[0].State = "PAUSED"
	err := pub.PublishStates([]hologram.Device{d})
	require.NoError(t, err)

	conn := mockMQTT.FindPublished("hologram/device/42/connectivity")
	require.Len(t, conn, 1)
	assert.Equal(t, "OFF", string(conn[0].Payload))

	sw := mockMQTT.FindPublished("hologram/device/42/switch/state")
	require.Len(t, sw, 1)
	assert.Equal(t, "OFF", string(sw[0].Payload))
}

func TestRemoveDiscovery(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram", "homeassistant", testLogger())

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

func TestPublishStatesNoCellularLink(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram", "homeassistant", testLogger())

	d := hologram.Device{ID: 1, IMEI: "111"}
	err := pub.PublishStates([]hologram.Device{d})
	require.NoError(t, err)

	attrs := mockMQTT.FindPublished("hologram/device/1/attributes")
	require.NotEmpty(t, attrs)

	var a deviceAttributes
	require.NoError(t, json.Unmarshal(attrs[0].Payload, &a))
	assert.Equal(t, "", a.State)
	assert.Equal(t, "", a.Plan)
	assert.Equal(t, "", a.Network)
	assert.Equal(t, int64(0), a.CurBillingDataUsed)
}

func TestPublishStatesNoLastSession(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram", "homeassistant", testLogger())

	d := testDevice()
	d.LastSession = nil
	err := pub.PublishStates([]hologram.Device{d})
	require.NoError(t, err)

	attrs := mockMQTT.FindPublished("hologram/device/42/attributes")
	require.NotEmpty(t, attrs)

	var a deviceAttributes
	require.NoError(t, json.Unmarshal(attrs[0].Payload, &a))
	// Should still get network from cellular link's last_network_used
	assert.Equal(t, "AT&T Mobility", a.Network)
	assert.Equal(t, "", a.RadioTech)
}

func TestPublishDiscoveryEmptyList(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram", "homeassistant", testLogger())

	err := pub.PublishDiscovery([]hologram.Device{})
	require.NoError(t, err)
	assert.Empty(t, mockMQTT.Published)
}

func TestPublishStatesEmptyList(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram", "homeassistant", testLogger())

	err := pub.PublishStates([]hologram.Device{})
	require.NoError(t, err)
	assert.Empty(t, mockMQTT.Published)
}

func TestPublishDiscoveryMQTTError(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	mockMQTT.PublishErr = errors.New("broker disconnected")
	pub := NewPublisher(mockMQTT, "hologram", "homeassistant", testLogger())

	err := pub.PublishDiscovery([]hologram.Device{testDevice()})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "broker disconnected")
}

func TestRemoveDiscoveryMQTTError(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	mockMQTT.PublishErr = errors.New("broker disconnected")
	pub := NewPublisher(mockMQTT, "hologram", "homeassistant", testLogger())

	err := pub.RemoveDiscovery([]hologram.Device{testDevice()})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "broker disconnected")
}

func TestPublishDiscoveryMultipleDevices(t *testing.T) {
	mockMQTT := mqtt.NewMockPublisher()
	pub := NewPublisher(mockMQTT, "hologram", "homeassistant", testLogger())

	d1 := hologram.Device{ID: 1, Name: "Device 1"}
	d2 := hologram.Device{ID: 2, Name: "Device 2"}

	err := pub.PublishDiscovery([]hologram.Device{d1, d2})
	require.NoError(t, err)

	expectedCount := (len(sensors) + 2) * 2
	assert.Len(t, mockMQTT.Published, expectedCount)
}

func TestBuildAttributesFull(t *testing.T) {
	d := testDevice()
	a := buildAttributes(d)

	assert.Equal(t, "LIVE", a.State)
	assert.Equal(t, "123456789012345", a.IMEI)
	assert.Equal(t, "89464278206108944117", a.ICCID)
	assert.Equal(t, "240422610894411", a.IMSI)
	assert.Equal(t, "+15551234567", a.PhoneNumber)
	assert.Equal(t, "13", a.Carrier)
	assert.Equal(t, "Global G3 Standard Flat Rate", a.Plan)
	assert.Equal(t, "AT&T Mobility", a.Network)
	assert.Equal(t, "LTE", a.RadioTech)
	assert.Equal(t, "2026-04-03 11:12:16", a.LastConnection)
	assert.Equal(t, int64(28580756), a.CurBillingDataUsed)
	assert.Equal(t, int64(649325609), a.LastBillingDataUsed)
	assert.Equal(t, "89044045117727494800000294996158", a.EID)
	assert.Equal(t, "ENABLED", a.ProfileState)
	assert.Equal(t, 42, a.DeviceID)
	assert.Equal(t, 10, a.OrgID)
	assert.Equal(t, 100, a.LinkID)
	assert.True(t, a.SessionActive)
}

func TestDeviceModelFallbacks(t *testing.T) {
	// Model set
	d := hologram.Device{ID: 1, Name: "Test", Model: "MT710", Manufacturer: "Cello"}
	ha := newDevice(d)
	assert.Equal(t, "MT710", ha.Model)
	assert.Equal(t, "Cello", ha.Manufacturer)

	// Model "Unknown", Type set
	d = hologram.Device{ID: 1, Name: "Test", Model: "Unknown", Type: "Tracker"}
	ha = newDevice(d)
	assert.Equal(t, "Tracker", ha.Model)

	// Both "Unknown"
	d = hologram.Device{ID: 1, Name: "Test", Model: "Unknown", Type: "Unknown"}
	ha = newDevice(d)
	assert.Equal(t, "SIM", ha.Model)

	// Empty
	d = hologram.Device{ID: 1, Name: "Test"}
	ha = newDevice(d)
	assert.Equal(t, "SIM", ha.Model)
}
