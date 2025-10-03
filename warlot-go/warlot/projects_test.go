package warlot

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestProjects_Init_Issue_Resolve(t *testing.T) {
	srv, cl := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/warlotSql/projects/init":
			mustStatus(t, r, "/warlotSql/projects/init")
			_ = r.Body.Close()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(InitProjectResponse{
				ProjectID: "proj-123",
				DBID:      "0xDB",
			})
		case "/auth/issue":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(IssueKeyResponse{APIKey: "key-abc", URL: "https://api.example.com/proj-123"})
		case "/warlotSql/projects/resolve":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ResolveProjectResponse{ProjectID: "proj-123", DBID: "0xDB"})
		default:
			http.NotFound(w, r)
		}
	})
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Init
	initRes, err := cl.InitProject(ctx, InitProjectRequest{
		HolderID: "0xH", ProjectName: "p", OwnerAddress: "0xU",
	})
	if err != nil {
		t.Fatal(err)
	}
	if initRes.ProjectID != "proj-123" {
		t.Fatalf("got %s", initRes.ProjectID)
	}

	// Issue
	iss, err := cl.IssueAPIKey(ctx, IssueKeyRequest{
		ProjectID: "proj-123", ProjectHolder: "0xH", ProjectName: "p", User: "0xU",
	})
	if err != nil {
		t.Fatal(err)
	}
	if iss.APIKey == "" {
		t.Fatalf("no apikey")
	}

	// Resolve
	res, err := cl.ResolveProject(ctx, ResolveProjectRequest{HolderID: "0xH", ProjectName: "p"})
	if err != nil {
		t.Fatal(err)
	}
	if res.ProjectID != "proj-123" {
		t.Fatalf("resolve mismatch: %s", res.ProjectID)
	}
}
