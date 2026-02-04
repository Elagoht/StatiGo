// Package client provides an HTTP client with retry logic and structured logging.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Config holds HTTP client configuration.
type Config struct {
	BaseURL         string
	Timeout         time.Duration
	ConnectTimeout  time.Duration
	TLSTimeout      time.Duration
	IdleConnTimeout time.Duration
	MaxRetries      int
	RetryWaitMin    time.Duration
	RetryWaitMax    time.Duration
	BearerToken     string
	UserAgent       string
	Headers         map[string]string
}

// DefaultConfig returns sensible default configuration.
func DefaultConfig() Config {
	return Config{
		Timeout:         30 * time.Second,
		ConnectTimeout:  10 * time.Second,
		TLSTimeout:      10 * time.Second,
		IdleConnTimeout: 90 * time.Second,
		MaxRetries:      3,
		RetryWaitMin:    1 * time.Second,
		RetryWaitMax:    30 * time.Second,
		UserAgent:       "Statigo/1.0",
	}
}

// Client is an HTTP client with retry and logging capabilities.
type Client struct {
	httpClient *http.Client
	config     Config
	logger     *slog.Logger
}

// New creates a new HTTP client.
func New(config Config, logger *slog.Logger) *Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   config.ConnectTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   config.TLSTimeout,
		IdleConnTimeout:       config.IdleConnTimeout,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		ResponseHeaderTimeout: config.Timeout,
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   config.Timeout,
			Transport: transport,
		},
		config: config,
		logger: logger,
	}
}

// Get performs a GET request and decodes the JSON response.
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	return c.doJSON(ctx, http.MethodGet, path, nil, result)
}

// Post performs a POST request with a JSON body and decodes the response.
func (c *Client) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doJSON(ctx, http.MethodPost, path, body, result)
}

// Put performs a PUT request with a JSON body and decodes the response.
func (c *Client) Put(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.doJSON(ctx, http.MethodPut, path, body, result)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) error {
	return c.doJSON(ctx, http.MethodDelete, path, nil, nil)
}

// doJSON performs an HTTP request with JSON encoding/decoding.
func (c *Client) doJSON(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	url := c.config.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.config.UserAgent != "" {
		req.Header.Set("User-Agent", c.config.UserAgent)
	}

	if c.config.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.BearerToken)
	}

	for key, value := range c.config.Headers {
		req.Header.Set(key, value)
	}

	// Perform request with retries
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate backoff
			wait := c.config.RetryWaitMin * time.Duration(1<<uint(attempt-1))
			if wait > c.config.RetryWaitMax {
				wait = c.config.RetryWaitMax
			}

			c.logger.Debug("retrying request",
				slog.String("method", method),
				slog.String("url", url),
				slog.Int("attempt", attempt),
				slog.Duration("wait", wait),
			)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}

		resp, lastErr = c.httpClient.Do(req)
		if lastErr == nil && resp.StatusCode < 500 {
			break
		}

		if resp != nil {
			resp.Body.Close()
		}
	}

	if lastErr != nil {
		return fmt.Errorf("request failed after %d retries: %w", c.config.MaxRetries, lastErr)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Body:       string(bodyBytes),
		}
	}

	// Decode response
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// Do performs a raw HTTP request.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// Set default headers
	if c.config.UserAgent != "" && req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.config.UserAgent)
	}

	if c.config.BearerToken != "" && req.Header.Get("Authorization") == "" {
		req.Header.Set("Authorization", "Bearer "+c.config.BearerToken)
	}

	return c.httpClient.Do(req)
}

// HTTPError represents an HTTP error response.
type HTTPError struct {
	StatusCode int
	Body       string
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}
