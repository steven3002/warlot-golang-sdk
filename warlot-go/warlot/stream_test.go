package warlot

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestExecSQLStream_RowScanner(t *testing.T) {
	// craft a minimal JSON that RowScanner expects:
	// {"ok":true,"rows":[{...},{...}]}
	body := `{"ok":true,"rows":[{"i":1},{"i":2},{"i":3}]}`

	srv, cl := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/sql") {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, body)
			return
		}
		http.NotFound(w, r)
	})
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sc, err := cl.ExecSQLStream(ctx, "proj", SQLRequest{SQL: "SELECT 1"})
	if err != nil {
		t.Fatal(err)
	}
	defer sc.Close()

	var got []int
	for {
		var m map[string]any
		if !sc.Next(&m) {
			break
		}
		got = append(got, int(m["i"].(float64)))
	}
	if sc.Err() != nil {
		t.Fatalf("scanner err: %v", sc.Err())
	}
	if len(got) != 3 || got[0] != 1 || got[2] != 3 {
		t.Fatalf("got=%v", got)
	}
}
