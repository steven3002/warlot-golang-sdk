package warlot

import (
	"context"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

func TestMigrator_Up(t *testing.T) {
	// fake FS with two migration files
	mfs := fstest.MapFS{
		"migrations/001_init.sql":     {Data: []byte(`CREATE TABLE IF NOT EXISTS _migrations (id TEXT PRIMARY KEY, applied_at TEXT);`)},
		"migrations/010_products.sql": {Data: []byte(`CREATE TABLE IF NOT EXISTS products (id INTEGER);`)},
	}

	// backend simulator state
	applied := map[string]bool{}

	srv, cl := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/sql") {
			var req SQLRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			defer r.Body.Close()
			sql := strings.ToLower(req.SQL)
			w.Header().Set("Content-Type", "application/json")

			switch {
			case strings.HasPrefix(sql, "create table"):
				json.NewEncoder(w).Encode(SQLResponse{OK: true, RowCount: intPtr(0)})

			case strings.HasPrefix(sql, "insert into _migrations"):
				// capture id value (name)
				if len(req.Params) >= 1 {
					if id, ok := req.Params[0].(string); ok {
						applied[id] = true
					}
				}
				json.NewEncoder(w).Encode(SQLResponse{OK: true, RowCount: intPtr(1)})

			case strings.HasPrefix(sql, "select id from _migrations"):
				// return already applied rows
				rows := []map[string]any{}
				for id := range applied {
					rows = append(rows, map[string]any{"id": id})
				}
				json.NewEncoder(w).Encode(SQLResponse{OK: true, Rows: rows})

			default:
				// treat all others as ok
				json.NewEncoder(w).Encode(SQLResponse{OK: true})
			}
			return
		}
		http.NotFound(w, r)
	})
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	proj := cl.Project("proj-123")

	// First run applies both (sorted)
	got, err := migrate.Up(ctx, proj, mapFSTrim{mfs}, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("applied=%v", got)
	}

	// Second run is idempotent (nothing to apply)
	got, err = migrate.Up(ctx, proj, mapFSTrim{mfs}, "migrations")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no new migrations, got %v", got)
	}
}

// mapFSTrim adapts fstest.MapFS to fs.FS (no change; just explicit type)
type mapFSTrim struct{ fs.FS }
