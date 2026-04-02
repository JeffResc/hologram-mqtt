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
		json.NewEncoder(w).Encode(resp)
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
		json.NewEncoder(w).Encode(resp)
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
		w.Write([]byte(`{"success":false,"error":"forbidden"}`))
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
		json.NewEncoder(w).Encode(resp)
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
		json.NewEncoder(w).Encode(resp)
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
