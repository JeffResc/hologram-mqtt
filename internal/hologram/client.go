package hologram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	defaultBaseURL  = "https://dashboard.hologram.io/api/1"
	defaultPageSize = 100
)

// Client defines the interface for interacting with the Hologram API.
type Client interface {
	ListDevices(ctx context.Context) ([]Device, error)
	SetDeviceState(ctx context.Context, orgID, deviceID int, state string) error
}

// Option configures the hologram client.
type Option func(*httpClient)

// WithBaseURL overrides the API base URL (useful for testing).
func WithBaseURL(url string) Option {
	return func(c *httpClient) {
		c.baseURL = url
	}
}

// WithRetryConfig overrides the retry settings (useful for testing).
func WithRetryConfig(retryMax int, retryWaitMin, retryWaitMax time.Duration) Option {
	return func(c *httpClient) {
		c.client.RetryMax = retryMax
		c.client.RetryWaitMin = retryWaitMin
		c.client.RetryWaitMax = retryWaitMax
	}
}

// WithOrgID sets the organization ID to filter API requests.
func WithOrgID(orgID int) Option {
	return func(c *httpClient) {
		c.orgID = orgID
	}
}

type httpClient struct {
	baseURL string
	apiKey  string
	orgID   int
	client  *retryablehttp.Client
	logger  *slog.Logger
}

// NewClient creates a new Hologram API client.
func NewClient(apiKey string, logger *slog.Logger, opts ...Option) Client {
	rc := retryablehttp.NewClient()
	rc.RetryMax = 3
	rc.RetryWaitMin = 5 * time.Second
	rc.RetryWaitMax = 20 * time.Second
	rc.HTTPClient.Timeout = 30 * time.Second
	rc.Logger = nil // suppress default logger
	rc.CheckRetry = rateLimitRetryPolicy

	c := &httpClient{
		baseURL: defaultBaseURL,
		apiKey:  apiKey,
		client:  rc,
		logger:  logger,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// rateLimitRetryPolicy retries only on 429 Too Many Requests.
func rateLimitRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if err != nil {
		return false, err
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return true, nil
	}
	return false, nil
}

// ListDevices fetches all devices, handling pagination automatically.
func (c *httpClient) ListDevices(ctx context.Context) ([]Device, error) {
	var all []Device
	startAfter := 0

	for {
		url := fmt.Sprintf("%s/devices?limit=%d", c.baseURL, defaultPageSize)
		if c.orgID > 0 {
			url += fmt.Sprintf("&orgid=%d", c.orgID)
		}
		if startAfter > 0 {
			url += fmt.Sprintf("&startafter=%d", startAfter)
		}

		resp, err := c.doRequest(ctx, http.MethodGet, url, "")
		if err != nil {
			return nil, fmt.Errorf("listing devices: %w", err)
		}

		c.logger.Debug("API response received", "bytes", len(resp))

		var listResp DeviceListResponse
		if err := json.Unmarshal(resp, &listResp); err != nil {
			return nil, fmt.Errorf("decoding device list: %w", err)
		}

		if !listResp.Success {
			return nil, fmt.Errorf("API returned success=false for device list")
		}

		all = append(all, listResp.Data...)

		if !listResp.Continues || len(listResp.Data) == 0 {
			break
		}

		startAfter = listResp.Data[len(listResp.Data)-1].ID
	}

	c.logger.Info("fetched devices from Hologram API", "count", len(all))
	return all, nil
}

// SetDeviceState pauses or resumes a device.
// State must be "pause" or "live".
func (c *httpClient) SetDeviceState(ctx context.Context, orgID, deviceID int, state string) error {
	if state != "pause" && state != "live" {
		return fmt.Errorf("invalid state %q: must be \"pause\" or \"live\"", state)
	}

	body := fmt.Sprintf(`{"state":%q,"deviceids":[%d],"orgid":%d}`, state, deviceID, orgID)
	url := c.baseURL + "/devices/batch/state"

	respBody, err := c.doRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("setting device state: %w", err)
	}

	var batchResp BatchStateResponse
	if err := json.Unmarshal(respBody, &batchResp); err != nil {
		return fmt.Errorf("decoding state response: %w", err)
	}

	if !batchResp.Success {
		return fmt.Errorf("API returned success=false for state change")
	}

	c.logger.Info("device state changed", "device_id", deviceID, "state", state)
	return nil
}

// doRequest performs an HTTP request using the retryable client.
func (c *httpClient) doRequest(ctx context.Context, method, url, body string) ([]byte, error) {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth("apikey", c.apiKey)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "unexpected status: " + strconv.Itoa(resp.StatusCode) + " body: " + string(respBody),
		}
	}

	return respBody, nil
}

// APIError represents an error from the Hologram API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("hologram API error (status %d): %s", e.StatusCode, e.Message)
}
