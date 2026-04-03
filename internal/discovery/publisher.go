package discovery

import (
	"encoding/json"
	"fmt"
	"log/slog"

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
	state := d.EffectiveState()
	connectivity := "OFF"
	if state == "LIVE" {
		connectivity = "ON"
	}
	if err := p.mqtt.Publish(prefix+"/connectivity", 1, true, []byte(connectivity)); err != nil {
		return err
	}

	// Publish switch state (ON = LIVE/active, OFF = PAUSED)
	switchState := "ON"
	if state != "LIVE" {
		switchState = "OFF"
	}
	if err := p.mqtt.Publish(prefix+"/switch/state", 1, true, []byte(switchState)); err != nil {
		return err
	}

	return nil
}

// deviceAttributes is the JSON published to the attributes topic.
type deviceAttributes struct {
	State              string `json:"state"`
	IMEI               string `json:"imei"`
	ICCID              string `json:"iccid"`
	IMSI               string `json:"imsi"`
	Carrier            string `json:"carrier"`
	Plan               string `json:"plan"`
	PhoneNumber        string `json:"phone_number"`
	LastConnection     string `json:"last_connection"`
	Network            string `json:"network"`
	RadioTech          string `json:"radio_tech"`
	CurBillingDataUsed int64  `json:"cur_billing_data_used"`
	LastBillingDataUsed int64 `json:"last_billing_data_used"`
	SessionActive      bool   `json:"session_active"`
	DeviceID           int    `json:"device_id"`
	OrgID              int    `json:"org_id"`
	LinkID             int    `json:"link_id"`
	EID                string `json:"eid"`
	ProfileState       string `json:"profile_state"`
}

func buildAttributes(d hologram.Device) deviceAttributes {
	attrs := deviceAttributes{
		State:    d.EffectiveState(),
		IMEI:     d.IMEI,
		DeviceID: d.ID,
		OrgID:    d.OrgID,
	}

	// Pull from last session (top-level on device)
	if d.LastSession != nil {
		attrs.Network = d.LastSession.NetworkName
		attrs.RadioTech = d.LastSession.RadioTech
		attrs.SessionActive = d.LastSession.Active
		if d.LastSession.SessionBegin != "" && d.LastSession.SessionBegin != "0000-00-00 00:00:00" {
			attrs.LastConnection = d.LastSession.SessionBegin
		}
	}

	// Pull from the primary cellular link
	if link := d.PrimaryCellularLink(); link != nil {
		attrs.LinkID = link.ID
		attrs.ICCID = link.SIM
		attrs.IMSI = fmt.Sprintf("%d", link.IMSI)
		attrs.PhoneNumber = link.MSISDN
		attrs.Carrier = link.CarrierID.String()
		attrs.CurBillingDataUsed = link.CurBillingDataUsed
		attrs.LastBillingDataUsed = link.LastBillingDataUsed
		attrs.EID = link.EID
		attrs.ProfileState = link.ProfileState

		if link.Plan != nil {
			attrs.Plan = link.Plan.Name
		}
		if attrs.LastConnection == "" && link.LastConnectTime != "" {
			attrs.LastConnection = link.LastConnectTime
		}
		if attrs.Network == "" && link.LastNetworkUsed != "" {
			attrs.Network = link.LastNetworkUsed
		}
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
