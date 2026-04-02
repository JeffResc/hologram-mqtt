// Package bridge orchestrates the Hologram API polling and MQTT publishing loop.
package bridge

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jeffresc/hologram-mqtt/internal/config"
	"github.com/jeffresc/hologram-mqtt/internal/discovery"
	"github.com/jeffresc/hologram-mqtt/internal/hologram"
	"github.com/jeffresc/hologram-mqtt/internal/mqtt"
)

// Bridge ties together the Hologram API client, MQTT publisher,
// and HA discovery publisher into a single polling loop.
type Bridge struct {
	hologram     hologram.Client
	mqtt         mqtt.Publisher
	discovery    *discovery.Publisher
	config       *config.Config
	knownDevices map[int]hologram.Device
	logger       *slog.Logger
}

// New creates a new Bridge instance.
func New(cfg *config.Config, hc hologram.Client, mc mqtt.Publisher, logger *slog.Logger) *Bridge {
	dp := discovery.NewPublisher(mc, cfg.MQTT.TopicPrefix, cfg.Discovery.Prefix, logger)
	return &Bridge{
		hologram:     hc,
		mqtt:         mc,
		discovery:    dp,
		config:       cfg,
		knownDevices: make(map[int]hologram.Device),
		logger:       logger,
	}
}

// Run starts the bridge loop. It blocks until the context is cancelled.
func (b *Bridge) Run(ctx context.Context) error {
	// Subscribe to switch command topics
	commandTopic := b.config.MQTT.TopicPrefix + "/device/+/switch/set"
	if err := b.mqtt.Subscribe(commandTopic, 1, b.handleCommand); err != nil {
		return fmt.Errorf("subscribing to command topic: %w", err)
	}
	b.logger.Info("subscribed to command topic", "topic", commandTopic)

	// Initial fetch and publish
	if err := b.poll(ctx); err != nil {
		b.logger.Error("initial poll failed", "error", err)
	}

	ticker := time.NewTicker(b.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			b.logger.Info("shutting down bridge")
			return nil
		case <-ticker.C:
			if err := b.poll(ctx); err != nil {
				b.logger.Error("poll failed", "error", err)
			}
		}
	}
}

func (b *Bridge) poll(ctx context.Context) error {
	devices, err := b.hologram.ListDevices(ctx)
	if err != nil {
		return fmt.Errorf("fetching devices: %w", err)
	}

	// Detect removed devices
	currentIDs := make(map[int]bool, len(devices))
	for _, d := range devices {
		currentIDs[d.ID] = true
	}

	var removed []hologram.Device
	for id, d := range b.knownDevices {
		if !currentIDs[id] {
			removed = append(removed, d)
		}
	}

	if len(removed) > 0 {
		if err := b.discovery.RemoveDiscovery(removed); err != nil {
			b.logger.Error("failed to remove discovery for removed devices", "error", err)
		}
	}

	// Detect new devices and publish discovery
	if b.config.Discovery.Enabled {
		var newDevices []hologram.Device
		for _, d := range devices {
			if _, exists := b.knownDevices[d.ID]; !exists {
				newDevices = append(newDevices, d)
			}
		}
		if len(newDevices) > 0 {
			if err := b.discovery.PublishDiscovery(newDevices); err != nil {
				b.logger.Error("failed to publish discovery for new devices", "error", err)
			}
		}
	}

	// Publish states for all devices
	if err := b.discovery.PublishStates(devices); err != nil {
		b.logger.Error("failed to publish states", "error", err)
	}

	// Update known devices
	b.knownDevices = make(map[int]hologram.Device, len(devices))
	for _, d := range devices {
		b.knownDevices[d.ID] = d
	}

	b.logger.Info("poll complete", "devices", len(devices), "new", len(devices)-len(b.knownDevices)+len(removed), "removed", len(removed))
	return nil
}

func (b *Bridge) handleCommand(topic string, payload []byte) {
	// Topic format: <prefix>/device/<id>/switch/set
	parts := strings.Split(topic, "/")
	if len(parts) < 5 {
		b.logger.Error("invalid command topic", "topic", topic)
		return
	}

	// Find the device ID part - it's after "device/"
	deviceIDStr := ""
	for i, p := range parts {
		if p == "device" && i+1 < len(parts) {
			deviceIDStr = parts[i+1]
			break
		}
	}

	deviceID, err := strconv.Atoi(deviceIDStr)
	if err != nil {
		b.logger.Error("invalid device ID in command topic", "topic", topic, "error", err)
		return
	}

	command := strings.TrimSpace(string(payload))
	var state string
	switch command {
	case "ON":
		state = "live"
	case "OFF":
		state = "pause"
	default:
		b.logger.Error("invalid command payload", "payload", command)
		return
	}

	device, ok := b.knownDevices[deviceID]
	if !ok {
		b.logger.Error("unknown device in command", "device_id", deviceID)
		return
	}

	b.logger.Info("executing command", "device_id", deviceID, "device_name", device.Name, "state", state)

	if err := b.hologram.SetDeviceState(context.Background(), device.OrgID, deviceID, state); err != nil {
		b.logger.Error("failed to set device state", "device_id", deviceID, "error", err)
		return
	}

	// Update local state and publish immediately
	if state == "live" {
		device.State = "LIVE"
	} else {
		device.State = "PAUSED"
	}
	b.knownDevices[deviceID] = device

	if err := b.discovery.PublishStates([]hologram.Device{device}); err != nil {
		b.logger.Error("failed to publish updated state", "device_id", deviceID, "error", err)
	}
}
