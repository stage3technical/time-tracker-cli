package client

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

// ErrReadOnly is returned when a non-GET request is attempted in read-only mode.
var ErrReadOnly = errors.New("read-only mode: only GET requests are allowed")

// HTTPDoer performs HTTP requests against the Time Tracker API.
type HTTPDoer interface {
	Do(method, path string, query url.Values, body []byte) (*Response, error)
	Get(path string, query url.Values) (*Response, error)
}

// ReadOnlyClient wraps an HTTPDoer and blocks all non-GET methods except /health.
type ReadOnlyClient struct {
	inner HTTPDoer
}

// NewReadOnly wraps c so only GET requests (and GET /health) are permitted.
func NewReadOnly(c HTTPDoer) *ReadOnlyClient {
	return &ReadOnlyClient{inner: c}
}

func (c *ReadOnlyClient) Do(method, path string, query url.Values, body []byte) (*Response, error) {
	if !isReadOnlyAllowed(method, path) {
		return nil, ErrReadOnly
	}
	return c.inner.Do(method, path, query, body)
}

func (c *ReadOnlyClient) Get(path string, query url.Values) (*Response, error) {
	return c.Do(http.MethodGet, path, query, nil)
}

func isReadOnlyAllowed(method, path string) bool {
	if strings.EqualFold(method, http.MethodGet) {
		return true
	}
	normalized := path
	if !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}
	return normalized == "/health"
}
