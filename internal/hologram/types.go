// Package hologram provides a client for the Hologram.io REST API.
package hologram

import "encoding/json"

// Device represents a Hologram device from the API.
type Device struct {
	ID           int          `json:"id"`
	OrgID        int          `json:"orgid"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	IMEI         string       `json:"imei"`
	IMEISV       string       `json:"imei_sv"`
	Model        string       `json:"model"`
	Manufacturer string       `json:"manufacturer"`
	PhoneNumber  string       `json:"phonenumber"`
	Hidden       int          `json:"hidden"`
	SIMCardID    int          `json:"simcardid"`
	Tunnelable   int          `json:"tunnelable"`
	IsHyper      bool         `json:"is_hyper"`
	Whencreated  string       `json:"whencreated"`
	Tags         []string     `json:"tags"`
	Links        *DeviceLinks `json:"links"`
	LastSession  *LastSession `json:"lastsession"`
}

// PrimaryCellularLink returns the first cellular link if present.
func (d *Device) PrimaryCellularLink() *CellularLink {
	if d.Links != nil && len(d.Links.Cellular) > 0 {
		return &d.Links.Cellular[0]
	}
	return nil
}

// EffectiveState returns the primary cellular link state.
func (d *Device) EffectiveState() string {
	if link := d.PrimaryCellularLink(); link != nil {
		return link.State
	}
	return ""
}

// LastSession represents the most recent session data for a device.
type LastSession struct {
	LinkID       json.Number `json:"linkid"`
	Bytes        int64       `json:"bytes"`
	SessionBegin string      `json:"session_begin"`
	SessionEnd   string      `json:"session_end"`
	IMEI         string      `json:"imei"`
	CellID       string      `json:"cellid"`
	NetworkName  string      `json:"network_name"`
	RadioTech    string      `json:"radio_access_technology"`
	Active       bool        `json:"active"`
}

// DeviceLinks holds linked resources from the API response.
type DeviceLinks struct {
	Cellular []CellularLink `json:"cellular"`
}

// CellularLink represents a SIM/cellular link associated with a device.
type CellularLink struct {
	ID                  int         `json:"id"`
	DeviceID            int         `json:"deviceid"`
	SIM                 string      `json:"sim"`
	IMSI                int64       `json:"imsi"`
	MSISDN              string      `json:"msisdn"`
	State               string      `json:"state"`
	LastConnectTime     string      `json:"last_connect_time"`
	LastNetworkUsed     string      `json:"last_network_used"`
	CarrierID           json.Number `json:"carrierid"`
	Plan                *Plan       `json:"plan"`
	Apn                 string      `json:"apn"`
	OverageLimit        int64       `json:"overage_limit"`
	SMSLimit            int         `json:"smslimit"`
	DataThreshold       int64       `json:"data_threshold"`
	WhenExpires         string      `json:"whenexpires"`
	Whenclaimed         string      `json:"whenclaimed"`
	CurBillingDataUsed  int64       `json:"cur_billing_data_used"`
	LastBillingDataUsed int64       `json:"last_billing_data_used"`
	EID                 string      `json:"eid"`
	EUICCType           string      `json:"euicc_type"`
	ProfileState        string      `json:"profile_state"`
	FallbackAttribute   string      `json:"fallback_attribute"`
}

// Plan represents a Hologram data plan.
type Plan struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Zone string `json:"zone"`
	Data int64  `json:"data"`
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
