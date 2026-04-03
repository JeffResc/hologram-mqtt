package hologram

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestListDevicesSinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/devices")

		user, pass, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "apikey", user)
		assert.Equal(t, "test-key", pass)

		resp := DeviceListResponse{
			Success:   true,
			Continues: false,
			Data: []Device{
				{ID: 1, Name: "Device 1", IMEI: "111", State: "LIVE", OrgID: 10},
				{ID: 2, Name: "Device 2", IMEI: "222", State: "PAUSED", OrgID: 10},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-key", testLogger(), WithBaseURL(server.URL))
	devices, err := client.ListDevices(context.Background())
	require.NoError(t, err)
	assert.Len(t, devices, 2)
	assert.Equal(t, "Device 1", devices[0].Name)
	assert.Equal(t, "Device 2", devices[1].Name)
}

func TestListDevicesPagination(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := callCount.Add(1)
		var resp DeviceListResponse
		if count == 1 {
			resp = DeviceListResponse{
				Success:   true,
				Continues: true,
				Data:      []Device{{ID: 1, Name: "Device 1", OrgID: 10}},
			}
		} else {
			assert.Contains(t, r.URL.RawQuery, "startafter=1")
			resp = DeviceListResponse{
				Success:   true,
				Continues: false,
				Data:      []Device{{ID: 2, Name: "Device 2", OrgID: 10}},
			}
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-key", testLogger(), WithBaseURL(server.URL))
	devices, err := client.ListDevices(context.Background())
	require.NoError(t, err)
	assert.Len(t, devices, 2)
	assert.Equal(t, int32(2), callCount.Load())
}

func TestListDevicesAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"success":false,"error":"forbidden"}`))
	}))
	defer server.Close()

	client := NewClient("bad-key", testLogger(), WithBaseURL(server.URL))
	_, err := client.ListDevices(context.Background())
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusForbidden, apiErr.StatusCode)
}

func TestSetDeviceState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/devices/batch/state")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req BatchStateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "pause", req.State)
		assert.Equal(t, []int{42}, req.DeviceIDs)
		assert.Equal(t, 10, req.OrgID)

		resp := BatchStateResponse{Success: true}
		resp.Data.JobID = "job-123"
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-key", testLogger(), WithBaseURL(server.URL))
	err := client.SetDeviceState(context.Background(), 10, 42, "pause")
	require.NoError(t, err)
}

func TestSetDeviceStateInvalidState(t *testing.T) {
	client := NewClient("test-key", testLogger())
	err := client.SetDeviceState(context.Background(), 10, 42, "invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid state")
}

func TestRateLimitRetry(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := callCount.Add(1)
		if count <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		resp := DeviceListResponse{
			Success:   true,
			Continues: false,
			Data:      []Device{{ID: 1, Name: "Device 1"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Use a client with short backoffs for testing
	c := &httpClient{
		baseURL: server.URL,
		apiKey:  "test-key",
		http:    &http.Client{},
		logger:  testLogger(),
	}

	devices, err := c.ListDevices(context.Background())
	require.NoError(t, err)
	assert.Len(t, devices, 1)
	assert.True(t, callCount.Load() >= 3)
}

func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := NewClient("test-key", testLogger(), WithBaseURL(server.URL))
	_, err := client.ListDevices(ctx)
	require.Error(t, err)
}

// --- Edge case tests ---

func TestListDevicesEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := DeviceListResponse{
			Success:   true,
			Continues: false,
			Data:      []Device{},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-key", testLogger(), WithBaseURL(server.URL))
	devices, err := client.ListDevices(context.Background())
	require.NoError(t, err)
	assert.Empty(t, devices)
}

func TestListDevicesSuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := DeviceListResponse{
			Success: false,
			Data:    nil,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-key", testLogger(), WithBaseURL(server.URL))
	_, err := client.ListDevices(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "success=false")
}

func TestListDevicesInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewClient("test-key", testLogger(), WithBaseURL(server.URL))
	_, err := client.ListDevices(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decoding device list")
}

func TestSetDeviceStateLive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req BatchStateRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, "live", req.State)

		resp := BatchStateResponse{Success: true}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-key", testLogger(), WithBaseURL(server.URL))
	err := client.SetDeviceState(context.Background(), 10, 42, "live")
	require.NoError(t, err)
}

func TestSetDeviceStateSuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := BatchStateResponse{Success: false}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient("test-key", testLogger(), WithBaseURL(server.URL))
	err := client.SetDeviceState(context.Background(), 10, 42, "pause")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "success=false")
}

func TestSetDeviceStateServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	client := NewClient("test-key", testLogger(), WithBaseURL(server.URL))
	err := client.SetDeviceState(context.Background(), 10, 42, "pause")
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
}

func TestAPIErrorString(t *testing.T) {
	err := &APIError{StatusCode: 403, Message: "forbidden"}
	assert.Contains(t, err.Error(), "403")
	assert.Contains(t, err.Error(), "forbidden")
}

func TestListDevicesWithNilOptionalFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a device with minimal fields (all optional fields missing/null)
		_, _ = w.Write([]byte(`{
			"success": true,
			"continues": false,
			"data": [{
				"id": 1,
				"orgid": 10,
				"name": "Minimal",
				"state": "LIVE",
				"plan": null,
				"last_connection_time": null,
				"recent_session_info": null,
				"tags": null
			}]
		}`))
	}))
	defer server.Close()

	client := NewClient("test-key", testLogger(), WithBaseURL(server.URL))
	devices, err := client.ListDevices(context.Background())
	require.NoError(t, err)
	require.Len(t, devices, 1)
	assert.Equal(t, "Minimal", devices[0].Name)
	assert.Nil(t, devices[0].Plan)
	assert.Nil(t, devices[0].LastConnectionTime)
	assert.Nil(t, devices[0].RecentSessionInfo)
}

func TestWithHTTPClient(t *testing.T) {
	// Verify the WithHTTPClient option works
	customHTTP := &http.Client{}
	client := NewClient("test-key", testLogger(), WithHTTPClient(customHTTP))
	// Just verify it doesn't panic
	assert.NotNil(t, client)
}
