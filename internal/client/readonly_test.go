package client

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
)

type stubDoer struct {
	method string
	path   string
	called bool
}

func (s *stubDoer) Do(method, path string, _ url.Values, _ []byte) (*Response, error) {
	s.called = true
	s.method = method
	s.path = path
	return &Response{StatusCode: http.StatusOK, Body: []byte(`{}`)}, nil
}

func (s *stubDoer) Get(path string, query url.Values) (*Response, error) {
	return s.Do(http.MethodGet, path, query, nil)
}

func TestReadOnlyClient_AllowsGET(t *testing.T) {
	inner := &stubDoer{}
	ro := NewReadOnly(inner)
	if _, err := ro.Get("/api/v1/persons", nil); err != nil {
		t.Fatalf("GET: %v", err)
	}
	if !inner.called || inner.method != http.MethodGet {
		t.Fatalf("inner not called with GET: %+v", inner)
	}
}

func TestReadOnlyClient_BlocksPOST(t *testing.T) {
	inner := &stubDoer{}
	ro := NewReadOnly(inner)
	_, err := ro.Do(http.MethodPost, "/api/v1/projects", nil, []byte(`{}`))
	if !errors.Is(err, ErrReadOnly) {
		t.Fatalf("expected ErrReadOnly, got %v", err)
	}
	if inner.called {
		t.Fatal("inner should not have been called")
	}
}

func TestReadOnlyClient_BlocksPUT(t *testing.T) {
	inner := &stubDoer{}
	ro := NewReadOnly(inner)
	_, err := ro.Do(http.MethodPut, "/api/v1/persons/x", nil, []byte(`{}`))
	if !errors.Is(err, ErrReadOnly) {
		t.Fatalf("expected ErrReadOnly, got %v", err)
	}
}

func TestReadOnlyClient_BlocksDELETE(t *testing.T) {
	inner := &stubDoer{}
	ro := NewReadOnly(inner)
	_, err := ro.Do(http.MethodDelete, "/api/v1/projects/x", nil, nil)
	if !errors.Is(err, ErrReadOnly) {
		t.Fatalf("expected ErrReadOnly, got %v", err)
	}
}
