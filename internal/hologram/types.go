// Package hologram provides a client for the Hologram.io REST API.
package hologram

// Device represents a Hologram device from the API.
type Device struct {
	ID                 int              `json:"id"`
	OrgID              int              `json:"orgid"`
	Name               string           `json:"name"`
	IMEI               string           `json:"imei"`
	IMEISV             string           `json:"imei_sv"`
	SIMNumber          string           `json:"sim_number"`
	IMSI               string           `json:"imsi"`
	MSISDN             string           `json:"msisdn"`
	State              string           `json:"state"`
	PhoneNumber        string           `json:"phone_number"`
	Carrier            string           `json:"carrier"`
	LastConnectionTime *int64           `json:"last_connection_time"`
	NetworkUsed        string           `json:"network_used"`
	DeviceType         string           `json:"device_type"`
	Manufacturer       string           `json:"manufacturer"`
	Tags               []string         `json:"tags"`
	Plan               *Plan            `json:"plan"`
	RecentSessionInfo  *SessionInfo     `json:"recent_session_info"`
	Links              *DeviceLinks     `json:"links"`
}

// Plan represents a Hologram data plan.
type Plan struct {
	Name string `json:"name"`
	Zone string `json:"zone"`
}

// SessionInfo represents recent session data for a device.
type SessionInfo struct {
	BytesUp       int64  `json:"bytes_up"`
	BytesDown     int64  `json:"bytes_down"`
	RadioTech     string `json:"radio_access_technology"`
	NetworkName   string `json:"network_name"`
}

// DeviceLinks holds linked resources from the API response.
type DeviceLinks struct {
	Cellular []CellularLink `json:"cellular"`
}

// CellularLink represents a SIM/cellular link associated with a device.
type CellularLink struct {
	ID              int    `json:"id"`
	SIM             string `json:"sim"`
	IMSI            string `json:"imsi"`
	MSISDN          string `json:"msisdn"`
	State           string `json:"state"`
	LastConnectTime string `json:"last_connect_time"`
	CarrierID       int    `json:"carrierid"`
}

// DeviceListResponse is the paginated response from GET /devices.
type DeviceListResponse struct {
	Success   bool     `json:"success"`
	Data      []Device `json:"data"`
	Continues bool     `json:"continues"`
	Limit     int      `json:"limit"`
	Size      int      `json:"size"`
}

// BatchStateRequest is the request body for POST /devices/batch/state.
type BatchStateRequest struct {
	State     string `json:"state"`
	DeviceIDs []int  `json:"deviceids"`
	OrgID     int    `json:"orgid"`
}

// BatchStateResponse is the response from POST /devices/batch/state.
type BatchStateResponse struct {
	Success bool `json:"success"`
	Data    struct {
		JobID string `json:"jobid"`
	} `json:"data"`
}
