package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientDoSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-jwt" {
			t.Errorf("auth header = %q", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/api/v1/persons" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":"1"}]`))
	}))
	defer srv.Close()

	c := New(srv.URL, "test-jwt")
	resp, err := c.Get("/api/v1/persons", nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if string(resp.Body) != `[{"id":"1"}]` {
		t.Errorf("body = %s", resp.Body)
	}
}

func TestClientDoError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail":"not found"}`))
	}))
	defer srv.Close()

	c := New(srv.URL, "token")
	resp, err := c.Get("/missing", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error type = %T", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d", apiErr.StatusCode)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("response status = %d", resp.StatusCode)
	}
}

func TestExitCode(t *testing.T) {
	tests := []struct {
		status int
		want   int
	}{
		{404, 4},
		{401, 5},
		{403, 6},
		{409, 9},
		{400, 2},
		{500, 1},
		{200, 0},
	}
	for _, tc := range tests {
		if got := ExitCode(tc.status); got != tc.want {
			t.Errorf("ExitCode(%d) = %d want %d", tc.status, got, tc.want)
		}
	}
}

func TestParseDetail(t *testing.T) {
	msg := ParseDetail(`{"detail":"Person not found"}`)
	if msg != `"Person not found"` {
		t.Errorf("detail = %q", msg)
	}
}
