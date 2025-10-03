package warlot

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestHeaders_Auth_Forwarded(t *testing.T) {
	var gotAPI, gotHolder, gotProj, gotX string

	srv, cl := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		gotAPI = r.Header.Get("x-api-key")
		gotHolder = r.Header.Get("x-holder-id")
		gotProj = r.Header.Get("x-project-name")
		gotX = r.Header.Get("x-extra")
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	cl.APIKey = "k"
	cl.HolderID = "h"
	cl.ProjectName = "p"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Any method calls doJSON; hit a harmless path using status/commit isn’t required here.
	_, _ = cl.doRequest(ctx, http.MethodGet, "/w", cl.authHeaders(), nil)
	// a second call with extra header
	h := cl.authHeaders()
	h.Set("x-extra", "1")
	_, _ = cl.doRequest(ctx, http.MethodGet, "/w", h, nil)

	if gotAPI != "k" || gotHolder != "h" || gotProj != "p" || gotX != "1" {
		t.Fatalf("headers not forwarded: api=%q holder=%q proj=%q x=%q", gotAPI, gotHolder, gotProj, gotX)
	}
}
