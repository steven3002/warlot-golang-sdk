package warlot

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestTables_Status_Commit_Pager(t *testing.T) {
	srv, cl := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.HasSuffix(r.URL.Path, "/tables"):
			json.NewEncoder(w).Encode(ListTablesResponse{Tables: []string{"products"}})
		case strings.Contains(r.URL.Path, "/tables/products/rows"):
			q := r.URL.Query()
			limit := q.Get("limit")
			offset := q.Get("offset")
			_ = limit
			_ = offset
			json.NewEncoder(w).Encode(BrowseRowsResponse{
				Limit:  2,
				Offset: 0,
				Table:  "products",
				Rows: []map[string]any{
					{"id": 1}, {"id": 2},
				},
			})
		case strings.HasSuffix(r.URL.Path, "/schema"):
			json.NewEncoder(w).Encode(map[string]any{"name": "products", "columns": []string{"id"}})
		case strings.HasSuffix(r.URL.Path, "/count"):
			json.NewEncoder(w).Encode(TableCountResponse{ProjectID: "proj-123", TableCount: 1})
		case strings.HasSuffix(r.URL.Path, "/status"):
			json.NewEncoder(w).Encode(map[string]any{"ok": true})
		case strings.HasSuffix(r.URL.Path, "/commit"):
			json.NewEncoder(w).Encode(map[string]any{"committed": true})
		default:
			http.NotFound(w, r)
		}
	})
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	proj := cl.Project("proj-123")

	// list
	lt, err := proj.Tables(ctx)
	if err != nil || len(lt.Tables) != 1 {
		t.Fatalf("tables: %+v err=%v", lt, err)
	}

	// browse + pager
	pgr := &Pager{Project: proj, Table: "products", Limit: 2}
	rows, err := pgr.Next(ctx)
	if err != nil || len(rows) != 2 {
		t.Fatalf("pager first: %v %v", len(rows), err)
	}
	// emulate end
	pgr.Done = true
	rows, err = pgr.Next(ctx)
	if err != nil || rows != nil {
		t.Fatalf("pager end: rows=%v err=%v", rows, err)
	}

	// schema
	sc, err := proj.Schema(ctx, "products")
	if err != nil || sc["name"] != "products" {
		t.Fatalf("schema: %v err=%v", sc, err)
	}

	// count
	cnt, err := proj.Count(ctx)
	if err != nil || cnt.TableCount != 1 {
		t.Fatalf("count: %+v err=%v", cnt, err)
	}

	// status
	st, err := proj.Status(ctx)
	if err != nil || st["ok"] != true {
		t.Fatalf("status: %+v err=%v", st, err)
	}

	// commit
	cm, err := proj.Commit(ctx)
	if err != nil || cm["committed"] != true {
		t.Fatalf("commit: %+v err=%v", cm, err)
	}
}
