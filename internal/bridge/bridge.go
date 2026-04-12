// Package bridge orchestrates the Hologram API polling and MQTT publishing loop.
package bridge

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

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
	ctx                context.Context
	mu                 sync.RWMutex
	knownDevices       map[int]hologram.Device
	lastSuccessfulPoll time.Time
	logger             *slog.Logger
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
	b.ctx = ctx

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

// Healthy returns true when the MQTT client is connected and polling is
// succeeding. Before the first poll completes, only the MQTT connection
// is checked. After that, the last successful poll must be within
// 2x the poll interval.
func (b *Bridge) Healthy() bool {
	if !b.mqtt.IsConnected() {
		return false
	}
	b.mu.RLock()
	lastPoll := b.lastSuccessfulPoll
	b.mu.RUnlock()
	if lastPoll.IsZero() {
		return true // haven't had a chance to poll yet
	}
	return time.Since(lastPoll) < 2*b.config.PollInterval
}

func (b *Bridge) poll(ctx context.Context) error {
	timer := prometheus.NewTimer(pollDuration)
	defer timer.ObserveDuration()

	devices, err := b.hologram.ListDevices(ctx)
	if err != nil {
		pollsTotal.WithLabelValues("error").Inc()
		return fmt.Errorf("fetching devices: %w", err)
	}

	b.mu.Lock()
	defer b.mu.Unlock()

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
	newKnown := make(map[int]hologram.Device, len(devices))
	for _, d := range devices {
		newKnown[d.ID] = d
	}

	removedCount := len(removed)
	newCount := len(devices) - (len(b.knownDevices) - removedCount)
	b.knownDevices = newKnown
	b.lastSuccessfulPoll = time.Now()

	pollsTotal.WithLabelValues("success").Inc()
	devicesTotal.Set(float64(len(devices)))

	b.logger.Info("poll complete", "devices", len(devices), "new", newCount, "removed", removedCount)
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

	commandsTotal.WithLabelValues(state).Inc()

	b.mu.RLock()
	device, ok := b.knownDevices[deviceID]
	b.mu.RUnlock()

	if !ok {
		b.logger.Error("unknown device in command", "device_id", deviceID)
		return
	}

	b.logger.Info("executing command", "device_id", deviceID, "device_name", device.Name, "state", state)

	if err := b.hologram.SetDeviceState(b.ctx, device.OrgID, deviceID, state); err != nil {
		b.logger.Error("failed to set device state", "device_id", deviceID, "error", err)
		return
	}

	// Build a deep copy with updated state, store it, and publish
	b.mu.Lock()
	device = b.knownDevices[deviceID]
	updated := copyDeviceWithState(device, state)
	b.knownDevices[deviceID] = updated
	b.mu.Unlock()

	if err := b.discovery.PublishStates([]hologram.Device{updated}); err != nil {
		b.logger.Error("failed to publish updated state", "device_id", deviceID, "error", err)
	}
}

// HealthHandler returns an http.HandlerFunc that reports bridge health.
func (b *Bridge) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if b.Healthy() {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("unhealthy"))
		}
	}
}

// copyDeviceWithState creates a deep copy of a device with the cellular link
// state updated, so the copy is safe to use outside the mutex.
func copyDeviceWithState(d hologram.Device, state string) hologram.Device {
	newState := "PAUSED"
	if state == "live" {
		newState = "LIVE"
	}

	if d.Links != nil && len(d.Links.Cellular) > 0 {
		newLinks := make([]hologram.CellularLink, len(d.Links.Cellular))
		copy(newLinks, d.Links.Cellular)
		newLinks[0].State = newState
		d.Links = &hologram.DeviceLinks{Cellular: newLinks}
	}

	return d
}
