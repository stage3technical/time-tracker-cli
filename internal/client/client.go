package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client performs authenticated HTTP requests against the Time Tracker API.
type Client struct {
	BaseURL      string
	Token        string
	ExtraHeaders map[string]string
	HTTPClient   *http.Client
}

// Response holds the result of an API call.
type Response struct {
	StatusCode int
	Body       []byte
}

// APIError represents a non-success HTTP response from the API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Body)
}

// New creates a Client with the given base URL and bearer token.
func New(baseURL, token string) *Client {
	return &Client{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		Token:      token,
		HTTPClient: http.DefaultClient,
	}
}

// Do sends an HTTP request. path may start with / or not. body may be nil.
func (c *Client) Do(method, path string, query url.Values, body []byte) (*Response, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u := c.BaseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	for key, value := range c.ExtraHeaders {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &Response{
		StatusCode: resp.StatusCode,
		Body:       data,
	}

	if resp.StatusCode >= 400 {
		return result, &APIError{StatusCode: resp.StatusCode, Body: string(data)}
	}
	return result, nil
}

// Get is a convenience wrapper for GET requests.
func (c *Client) Get(path string, query url.Values) (*Response, error) {
	return c.Do(http.MethodGet, path, query, nil)
}

// ExitCode maps HTTP status to a process exit code for scripting.
func ExitCode(status int) int {
	switch status {
	case http.StatusNotFound:
		return 4
	case http.StatusUnauthorized:
		return 5
	case http.StatusForbidden:
		return 6
	case http.StatusConflict:
		return 9
	case http.StatusBadRequest:
		return 2
	default:
		if status >= 500 {
			return 1
		}
		if status >= 400 {
			return 1
		}
		return 0
	}
}

// ParseDetail tries to extract a human-readable message from an API error body.
func ParseDetail(body string) string {
	var payload struct {
		Detail json.RawMessage `json:"detail"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return body
	}
	if len(payload.Detail) == 0 {
		return body
	}
	return string(payload.Detail)
}
