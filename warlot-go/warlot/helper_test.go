package warlot

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	srv := httptest.NewServer(handler)
	cl := New(
		WithBaseURL(srv.URL),
		WithRetries(2),
		WithBackoff(50*time.Millisecond, 200*time.Millisecond),
	)
	return srv, cl
}

func mustStatus(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if r.URL.Path != want {
		t.Fatalf("path = %s, want %s", r.URL.Path, want)
	}
}
