package bridge

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jeffresc/hologram-mqtt/internal/config"
	"github.com/jeffresc/hologram-mqtt/internal/hologram"
	"github.com/jeffresc/hologram-mqtt/internal/mqtt"
)

func fuzzBridge() *Bridge {
	mockMQTT := mqtt.NewMockPublisher()
	mockHolo := &fuzzHologramClient{}

	cfg := &config.Config{
		Hologram: config.HologramConfig{APIKey: "test-key"},
		MQTT: config.MQTTConfig{
			Broker:      "tcp://localhost:1883",
			TopicPrefix: "hologram",
			ClientID:    "fuzz",
		},
		Discovery: config.DiscoveryConfig{
			Prefix:  "homeassistant",
			Enabled: true,
		},
		PollInterval: 1 * time.Minute,
	}

	b := New(cfg, mockHolo, mockMQTT, slog.New(slog.NewTextHandler(io.Discard, nil)))

	// Seed a known device so some command paths are exercised
	b.knownDevices[42] = hologram.Device{
		ID:    42,
		OrgID: 10,
		Name:  "Fuzz Device",
		Links: &hologram.DeviceLinks{
			Cellular: []hologram.CellularLink{{ID: 100, State: "LIVE"}},
		},
	}

	return b
}

type fuzzHologramClient struct{}

func (f *fuzzHologramClient) ListDevices(_ context.Context) ([]hologram.Device, error) {
	return nil, nil
}

func (f *fuzzHologramClient) SetDeviceState(_ context.Context, _, _ int, _ string) error {
	return nil
}

func FuzzHandleCommand(f *testing.F) {
	f.Add("hologram/device/42/switch/set", []byte("ON"))
	f.Add("hologram/device/42/switch/set", []byte("OFF"))
	f.Add("hologram/device/999/switch/set", []byte("ON"))
	f.Add("too/short", []byte("ON"))
	f.Add("", []byte(""))
	f.Add("a/b/c/d/e/f/g/h", []byte("INVALID"))
	f.Add("hologram/device/notanumber/switch/set", []byte("ON"))

	f.Fuzz(func(t *testing.T, topic string, payload []byte) {
		b := fuzzBridge()
		// Must not panic
		b.handleCommand(topic, payload)
	})
}
