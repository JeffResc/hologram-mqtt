package discovery

import (
	"fmt"

	"github.com/jeffresc/hologram-mqtt/internal/hologram"
)

// Version is set at build time via ldflags.
var Version = "dev"

func newOrigin() HAOrigin {
	return HAOrigin{
		Name: "hologram-mqtt",
		SW:   Version,
		URL:  "https://github.com/jeffresc/hologram-mqtt",
	}
}

func newDevice(d hologram.Device) HADevice {
	model := d.Model
	if model == "" || model == "Unknown" {
		model = d.Type
	}
	if model == "" || model == "Unknown" {
		model = "SIM"
	}
	mfr := d.Manufacturer
	if mfr == "" {
		mfr = "Hologram"
	}
	return HADevice{
		Identifiers:  []string{fmt.Sprintf("hologram_%d", d.ID)},
		Name:         d.Name,
		Manufacturer: mfr,
		Model:        model,
		SerialNumber: d.IMEI,
	}
}

func newAvailability(topicPrefix string, deviceID int) []Availability {
	return []Availability{
		{Topic: topicPrefix + "/status"},
		{Topic: fmt.Sprintf("%s/device/%d/availability", topicPrefix, deviceID)},
	}
}

type sensorDef struct {
	name           string
	objectSuffix   string
	icon           string
	valueTemplate  string
	entityCategory string
	deviceClass    string
}

var sensors = []sensorDef{
	{name: "State", objectSuffix: "state", icon: "mdi:state-machine", valueTemplate: "{{ value_json.state }}", entityCategory: "diagnostic"},
	{name: "IMEI", objectSuffix: "imei", icon: "mdi:cellphone-key", valueTemplate: "{{ value_json.imei }}", entityCategory: "diagnostic"},
	{name: "ICCID", objectSuffix: "iccid", icon: "mdi:sim", valueTemplate: "{{ value_json.iccid }}", entityCategory: "diagnostic"},
	{name: "IMSI", objectSuffix: "imsi", icon: "mdi:sim-alert", valueTemplate: "{{ value_json.imsi }}", entityCategory: "diagnostic"},
	{name: "EID", objectSuffix: "eid", icon: "mdi:sim-outline", valueTemplate: "{{ value_json.eid }}", entityCategory: "diagnostic"},
	{name: "Carrier", objectSuffix: "carrier", icon: "mdi:antenna", valueTemplate: "{{ value_json.carrier }}", entityCategory: "diagnostic"},
	{name: "Plan", objectSuffix: "plan", icon: "mdi:card-account-details", valueTemplate: "{{ value_json.plan }}", entityCategory: "diagnostic"},
	{name: "Phone Number", objectSuffix: "phone_number", icon: "mdi:phone", valueTemplate: "{{ value_json.phone_number }}", entityCategory: "diagnostic"},
	{name: "Last Connection", objectSuffix: "last_connection", icon: "mdi:clock-outline", valueTemplate: "{{ value_json.last_connection }}", entityCategory: "diagnostic"},
	{name: "Network", objectSuffix: "network", icon: "mdi:access-point-network", valueTemplate: "{{ value_json.network }}", entityCategory: "diagnostic"},
	{name: "Radio Technology", objectSuffix: "radio_tech", icon: "mdi:signal-4g", valueTemplate: "{{ value_json.radio_tech }}", entityCategory: "diagnostic"},
	{name: "Profile State", objectSuffix: "profile_state", icon: "mdi:shield-check", valueTemplate: "{{ value_json.profile_state }}", entityCategory: "diagnostic"},
	{name: "Current Billing Data Used", objectSuffix: "cur_billing_data", icon: "mdi:chart-bar", valueTemplate: "{{ value_json.cur_billing_data_used }}", entityCategory: "diagnostic"},
	{name: "Last Billing Data Used", objectSuffix: "last_billing_data", icon: "mdi:chart-bar", valueTemplate: "{{ value_json.last_billing_data_used }}", entityCategory: "diagnostic"},
}

// SensorConfigs generates HA discovery configs for all sensor entities of a device.
func SensorConfigs(d hologram.Device, topicPrefix, discoveryPrefix string) []EntityConfig {
	device := newDevice(d)
	origin := newOrigin()
	avail := newAvailability(topicPrefix, d.ID)
	stateTopic := fmt.Sprintf("%s/device/%d/attributes", topicPrefix, d.ID)

	var configs []EntityConfig
	for _, s := range sensors {
		nodeID := fmt.Sprintf("hologram_%d", d.ID)
		objectID := fmt.Sprintf("hologram_%d_%s", d.ID, s.objectSuffix)
		topic := fmt.Sprintf("%s/sensor/%s/%s/config", discoveryPrefix, nodeID, s.objectSuffix)

		name := s.name
		payload := DiscoveryPayload{
			Name:             &name,
			UniqueID:         objectID,
			ObjectID:         objectID,
			StateTopic:       stateTopic,
			ValueTemplate:    s.valueTemplate,
			Icon:             s.icon,
			EntityCategory:   s.entityCategory,
			DeviceClass:      s.deviceClass,
			Device:           device,
			Origin:           origin,
			Availability:     avail,
			AvailabilityMode: "all",
		}

		configs = append(configs, EntityConfig{Topic: topic, Payload: payload})
	}
	return configs
}

// BinarySensorConfig generates the HA discovery config for the connectivity binary sensor.
func BinarySensorConfig(d hologram.Device, topicPrefix, discoveryPrefix string) EntityConfig {
	device := newDevice(d)
	origin := newOrigin()
	avail := newAvailability(topicPrefix, d.ID)
	nodeID := fmt.Sprintf("hologram_%d", d.ID)
	objectID := fmt.Sprintf("hologram_%d_connectivity", d.ID)
	stateTopic := fmt.Sprintf("%s/device/%d/connectivity", topicPrefix, d.ID)
	topic := fmt.Sprintf("%s/binary_sensor/%s/connectivity/config", discoveryPrefix, nodeID)

	name := "Connectivity"
	return EntityConfig{
		Topic: topic,
		Payload: DiscoveryPayload{
			Name:             &name,
			UniqueID:         objectID,
			ObjectID:         objectID,
			StateTopic:       stateTopic,
			DeviceClass:      "connectivity",
			PayloadOn:        "ON",
			PayloadOff:       "OFF",
			Device:           device,
			Origin:           origin,
			Availability:     avail,
			AvailabilityMode: "all",
		},
	}
}

// SwitchConfig generates the HA discovery config for the pause/resume switch.
func SwitchConfig(d hologram.Device, topicPrefix, discoveryPrefix string) EntityConfig {
	device := newDevice(d)
	origin := newOrigin()
	avail := newAvailability(topicPrefix, d.ID)
	nodeID := fmt.Sprintf("hologram_%d", d.ID)
	objectID := fmt.Sprintf("hologram_%d_active", d.ID)
	stateTopic := fmt.Sprintf("%s/device/%d/switch/state", topicPrefix, d.ID)
	commandTopic := fmt.Sprintf("%s/device/%d/switch/set", topicPrefix, d.ID)
	topic := fmt.Sprintf("%s/switch/%s/active/config", discoveryPrefix, nodeID)

	name := "Active"
	return EntityConfig{
		Topic: topic,
		Payload: DiscoveryPayload{
			Name:             &name,
			UniqueID:         objectID,
			ObjectID:         objectID,
			StateTopic:       stateTopic,
			CommandTopic:     commandTopic,
			Icon:             "mdi:power",
			PayloadOn:        "ON",
			PayloadOff:       "OFF",
			Device:           device,
			Origin:           origin,
			Availability:     avail,
			AvailabilityMode: "all",
		},
	}
}

// AllConfigs returns all discovery configs for a device.
func AllConfigs(d hologram.Device, topicPrefix, discoveryPrefix string) []EntityConfig {
	configs := SensorConfigs(d, topicPrefix, discoveryPrefix)
	configs = append(configs, BinarySensorConfig(d, topicPrefix, discoveryPrefix))
	configs = append(configs, SwitchConfig(d, topicPrefix, discoveryPrefix))
	return configs
}
