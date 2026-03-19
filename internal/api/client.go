// Package api provides a typed HTTP client for the APsystems OpenAPI.
// It handles authentication, JSON decoding, retries, and rate-limit back-off.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/apsystems/mcp-server/internal/auth"
	"github.com/apsystems/mcp-server/internal/models"
)

const (
	defaultBaseURL = "https://api.apsystemsema.com:9282"
	defaultTimeout = 30 * time.Second
	maxRetries     = 3
	retryDelay     = 2 * time.Second
)

// Client is a thread-safe APsystems API client.
type Client struct {
	baseURL    string
	appID      string
	appSecret  string
	httpClient *http.Client
	logger     *slog.Logger

	// Simple sliding-window rate limiter: tracks last request time.
	mu       sync.Mutex
	lastCall time.Time
	minGap   time.Duration // minimum gap between requests
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the default API base URL.
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") }
}

// WithLogger sets a structured logger.
func WithLogger(l *slog.Logger) Option {
	return func(c *Client) { c.logger = l }
}

// WithRateLimit sets the minimum gap between consecutive requests.
func WithRateLimit(d time.Duration) Option {
	return func(c *Client) { c.minGap = d }
}

// WithHTTPClient sets a custom *http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// NewClient creates a new APsystems API client.
func NewClient(appID, appSecret string, opts ...Option) *Client {
	c := &Client{
		baseURL:   defaultBaseURL,
		appID:     appID,
		appSecret: appSecret,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		logger: slog.Default(),
		minGap: 200 * time.Millisecond, // default: max 5 req/s
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Do executes an authenticated request against the APsystems API.
//
//   - method: HTTP method (GET, POST, etc.)
//   - path: URL path segment after the base URL, e.g. "/user/api/v2/systems/details/ABC"
//   - query: optional query parameters (energy_level, date_range, etc.)
//   - out: pointer to the struct the JSON "data" field should be decoded into; may be nil.
//
// It returns the decoded "data" as raw JSON and any error.
func (c *Client) Do(ctx context.Context, method, path string, query map[string]string, out interface{}) (json.RawMessage, error) {
	// Rate-limit: wait if we're going too fast.
	c.throttle()

	fullURL := c.baseURL + path
	if len(query) > 0 {
		v := url.Values{}
		for k, val := range query {
			v.Set(k, val)
		}
		fullURL += "?" + v.Encode()
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Info("retrying request",
				"attempt", attempt,
				"path", path,
			)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay * time.Duration(attempt)):
			}
		}

		raw, err := c.doOnce(ctx, method, fullURL)
		if err != nil {
			lastErr = err
			// Only retry on transient errors.
			if isTransient(err) {
				continue
			}
			return nil, err
		}

		// Decode the wrapper.
		var resp models.APIResponse
		if err := json.Unmarshal(raw, &resp); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}

		if resp.Code != 0 {
			desc := models.APIErrorCodes[resp.Code]
			if desc == "" {
				desc = "unknown error"
			}
			apiErr := &APIError{Code: resp.Code, Message: desc}
			// Retry on rate-limit / busy errors.
			if resp.Code == 7002 || resp.Code == 7003 {
				lastErr = apiErr
				continue
			}
			return nil, apiErr
		}

		// Decode "data" into the caller's struct if provided.
		if out != nil && resp.Data != nil {
			if err := json.Unmarshal(resp.Data, out); err != nil {
				return nil, fmt.Errorf("decode data field: %w", err)
			}
		}

		return resp.Data, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// doOnce performs a single HTTP request with authentication.
func (c *Client) doOnce(ctx context.Context, method, fullURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	if err := auth.SignRequest(req, c.appID, c.appSecret); err != nil {
		return nil, fmt.Errorf("sign request: %w", err)
	}

	c.logger.Debug("api request",
		"method", method,
		"url", fullURL,
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= 500 {
		return nil, &TransientError{StatusCode: resp.StatusCode, Body: string(body)}
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, body)
	}

	return body, nil
}

// throttle ensures a minimum gap between consecutive API calls.
func (c *Client) throttle() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.lastCall.IsZero() {
		elapsed := time.Since(c.lastCall)
		if elapsed < c.minGap {
			time.Sleep(c.minGap - elapsed)
		}
	}
	c.lastCall = time.Now()
}

func isTransient(err error) bool {
	_, ok := err.(*TransientError)
	return ok
}

// APIError represents a non-zero response code from the APsystems API.
type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("apsystems api error %d: %s", e.Code, e.Message)
}

// TransientError wraps a 5xx or network error that is safe to retry.
type TransientError struct {
	StatusCode int
	Body       string
}

func (e *TransientError) Error() string {
	return fmt.Sprintf("transient http %d: %s", e.StatusCode, e.Body)
}
