package hologram

import (
	"encoding/json"
	"testing"
)

func FuzzDeviceListResponse(f *testing.F) {
	f.Add([]byte(`{"success":true,"data":[],"continues":false,"limit":100,"size":0}`))
	f.Add([]byte(`{"success":true,"data":[{"id":1,"orgid":10,"name":"Test"}],"continues":false}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))
	f.Add([]byte(`{"data":null}`))
	f.Add([]byte(`{"data":[{"links":{"cellular":[{"plan":null}]}}]}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var resp DeviceListResponse
		// Must not panic regardless of input
		_ = json.Unmarshal(data, &resp)

		// If parsing succeeded, exercise the accessor methods
		for _, d := range resp.Data {
			_ = d.EffectiveState()
			_ = d.PrimaryCellularLink()
		}
	})
}

func FuzzBatchStateResponse(f *testing.F) {
	f.Add([]byte(`{"success":true,"data":{"jobid":"abc123"}}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`null`))
	f.Add([]byte(`{"success":false,"data":{}}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var resp BatchStateResponse
		// Must not panic regardless of input
		_ = json.Unmarshal(data, &resp)
	})
}
