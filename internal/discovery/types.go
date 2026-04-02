// Package discovery publishes Home Assistant MQTT auto-discovery configurations.
package discovery

// DiscoveryPayload is the config payload published to HA discovery topics.
type DiscoveryPayload struct {
	Name             *string        `json:"name"`
	UniqueID         string         `json:"unique_id"`
	ObjectID         string         `json:"object_id"`
	StateTopic       string         `json:"state_topic,omitempty"`
	CommandTopic     string         `json:"command_topic,omitempty"`
	ValueTemplate    string         `json:"value_template,omitempty"`
	DeviceClass      string         `json:"device_class,omitempty"`
	EntityCategory   string         `json:"entity_category,omitempty"`
	Icon             string         `json:"icon,omitempty"`
	PayloadOn        string         `json:"payload_on,omitempty"`
	PayloadOff       string         `json:"payload_off,omitempty"`
	PayloadAvailable string         `json:"payload_available,omitempty"`
	PayloadNotAvail  string         `json:"payload_not_available,omitempty"`
	Availability     []Availability `json:"availability,omitempty"`
	AvailabilityMode string         `json:"availability_mode,omitempty"`
	Device           HADevice       `json:"device"`
	Origin           HAOrigin       `json:"origin"`
}

// HADevice represents a Home Assistant device in the discovery payload.
type HADevice struct {
	Identifiers  []string `json:"identifiers"`
	Name         string   `json:"name"`
	Manufacturer string   `json:"manufacturer"`
	Model        string   `json:"model"`
	SWVersion    string   `json:"sw_version,omitempty"`
	SerialNumber string   `json:"serial_number,omitempty"`
}

// HAOrigin identifies the application publishing discovery messages.
type HAOrigin struct {
	Name string `json:"name"`
	SW   string `json:"sw"`
	URL  string `json:"url"`
}

// Availability defines an MQTT availability topic.
type Availability struct {
	Topic string `json:"topic"`
}

// EntityConfig pairs a discovery topic with its payload.
type EntityConfig struct {
	Topic   string
	Payload DiscoveryPayload
}
