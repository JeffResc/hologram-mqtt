package discovery

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jeffresc/hologram-mqtt/internal/hologram"
	"github.com/jeffresc/hologram-mqtt/internal/mqtt"
)

// Publisher handles publishing HA MQTT discovery configs and device state.
type Publisher struct {
	mqtt            mqtt.Publisher
	topicPrefix     string
	discoveryPrefix string
	logger          *slog.Logger
}

// NewPublisher creates a new discovery publisher.
func NewPublisher(mqttPub mqtt.Publisher, topicPrefix, discoveryPrefix string, logger *slog.Logger) *Publisher {
	return &Publisher{
		mqtt:            mqttPub,
		topicPrefix:     topicPrefix,
		discoveryPrefix: discoveryPrefix,
		logger:          logger,
	}
}

// PublishDiscovery publishes HA discovery config messages for all devices.
func (p *Publisher) PublishDiscovery(devices []hologram.Device) error {
	for _, d := range devices {
		configs := AllConfigs(d, p.topicPrefix, p.discoveryPrefix)
		for _, cfg := range configs {
			payload, err := json.Marshal(cfg.Payload)
			if err != nil {
				return fmt.Errorf("marshaling discovery for device %d: %w", d.ID, err)
			}
			if err := p.mqtt.Publish(cfg.Topic, 1, true, payload); err != nil {
				return fmt.Errorf("publishing discovery for device %d: %w", d.ID, err)
			}
		}
		p.logger.Debug("published discovery configs", "device_id", d.ID, "name", d.Name)
	}
	p.logger.Info("published discovery configs for all devices", "count", len(devices))
	return nil
}

// PublishStates publishes current state values for all devices.
func (p *Publisher) PublishStates(devices []hologram.Device) error {
	for _, d := range devices {
		if err := p.publishDeviceState(d); err != nil {
			p.logger.Error("failed to publish state", "device_id", d.ID, "error", err)
			continue
		}
	}
	return nil
}

func (p *Publisher) publishDeviceState(d hologram.Device) error {
	prefix := fmt.Sprintf("%s/device/%d", p.topicPrefix, d.ID)

	// Publish availability
	if err := p.mqtt.Publish(prefix+"/availability", 1, true, []byte("online")); err != nil {
		return err
	}

	// Publish attributes as JSON for sensor value_templates
	attrs := buildAttributes(d)
	attrJSON, err := json.Marshal(attrs)
	if err != nil {
		return fmt.Errorf("marshaling attributes: %w", err)
	}
	if err := p.mqtt.Publish(prefix+"/attributes", 1, true, attrJSON); err != nil {
		return err
	}

	// Publish connectivity binary sensor
	connectivity := "OFF"
	if d.State == "LIVE" {
		connectivity = "ON"
	}
	if err := p.mqtt.Publish(prefix+"/connectivity", 1, true, []byte(connectivity)); err != nil {
		return err
	}

	// Publish switch state (ON = LIVE/active, OFF = PAUSED)
	switchState := "ON"
	if d.State == "PAUSED" || d.State == "DEAD" {
		switchState = "OFF"
	}
	if err := p.mqtt.Publish(prefix+"/switch/state", 1, true, []byte(switchState)); err != nil {
		return err
	}

	return nil
}

// deviceAttributes is the JSON published to the attributes topic.
type deviceAttributes struct {
	State          string `json:"state"`
	IMEI           string `json:"imei"`
	SIMNumber      string `json:"sim_number"`
	Carrier        string `json:"carrier"`
	Plan           string `json:"plan"`
	PhoneNumber    string `json:"phone_number"`
	LastConnection string `json:"last_connection"`
	Network        string `json:"network"`
	DataUp         int64  `json:"data_up"`
	DataDown       int64  `json:"data_down"`
}

func buildAttributes(d hologram.Device) deviceAttributes {
	attrs := deviceAttributes{
		State:       d.State,
		IMEI:        d.IMEI,
		SIMNumber:   d.SIMNumber,
		Carrier:     d.Carrier.String(),
		PhoneNumber: d.PhoneNumber,
		Network:     d.NetworkUsed,
	}

	if d.Plan != nil {
		attrs.Plan = d.Plan.Name
	}

	if d.LastConnectionTime != nil {
		t := time.Unix(*d.LastConnectionTime, 0).UTC()
		attrs.LastConnection = t.Format(time.RFC3339)
	}

	if d.RecentSessionInfo != nil {
		attrs.DataUp = d.RecentSessionInfo.BytesUp
		attrs.DataDown = d.RecentSessionInfo.BytesDown
	}

	return attrs
}

// RemoveDiscovery publishes empty retained messages to remove discovery configs.
func (p *Publisher) RemoveDiscovery(devices []hologram.Device) error {
	for _, d := range devices {
		configs := AllConfigs(d, p.topicPrefix, p.discoveryPrefix)
		for _, cfg := range configs {
			if err := p.mqtt.Publish(cfg.Topic, 1, true, []byte{}); err != nil {
				return fmt.Errorf("removing discovery for device %d: %w", d.ID, err)
			}
		}

		// Clear availability
		prefix := fmt.Sprintf("%s/device/%d", p.topicPrefix, d.ID)
		_ = p.mqtt.Publish(prefix+"/availability", 1, true, []byte("offline"))

		p.logger.Info("removed discovery configs", "device_id", d.ID, "name", d.Name)
	}
	return nil
}
