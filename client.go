package gopocketbaseclient

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// Version information
const (
	Version = "1.4.2"
	Name    = "Go PocketBase Client"
)

func NewClient(baseURL, jwtToken string) *Client {
	// Create optimized transport with connection pooling
	transport := &http.Transport{
		// Connection pooling settings
		MaxIdleConns:        100,              // Maximum idle connections across all hosts
		MaxIdleConnsPerHost: 20,               // Maximum idle connections per host
		MaxConnsPerHost:     50,               // Maximum connections per host
		IdleConnTimeout:     90 * time.Second, // How long idle connections stay open

		// Timeout settings for better performance
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,  // Connection timeout
			KeepAlive: 30 * time.Second, // Keep-alive probe interval
		}).DialContext,

		// Response header timeout
		ResponseHeaderTimeout: 10 * time.Second,

		// Expect continue timeout
		ExpectContinueTimeout: 1 * time.Second,

		// TLS handshake timeout
		TLSHandshakeTimeout: 5 * time.Second,

		// Disable compression for better CPU performance (PocketBase responses are usually small)
		DisableCompression: true,
	}

	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second, // Increased overall timeout for bulk operations
		},
		Token: jwtToken,
	}
}

func (c *Client) doRequest(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = MarshalPocketBaseJSON(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequest(method, c.BaseURL+endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if err := checkHTTPStatus(resp.StatusCode, respBody); err != nil {
		return nil, err
	}

	return respBody, nil
}

// New function to check HTTP status
func checkHTTPStatus(statusCode int, respBody []byte) error {
	if statusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", statusCode, respBody)
	}
	return nil
}

// GetVersion returns the client library version
func GetVersion() string {
	return Version
}

// GetLibraryInfo returns library name and version
func GetLibraryInfo() (string, string) {
	return Name, Version
}
