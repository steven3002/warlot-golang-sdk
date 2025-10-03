package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli"
	"github.com/steven3002/warlot-golang-sdk/warlot-go/warlot"
)

// RunSQL executes a SQL statement (optionally streaming rows).
func RunSQL(args []string) error {
	fs := flag.NewFlagSet("sql", flag.ContinueOnError)
	projectID := fs.String("project", "", "Project ID (required)")
	query := fs.String("q", "", "SQL query (required)")
	paramsJSON := fs.String("params", "", "Params JSON array, e.g. [\"Laptop\",999.99]")
	idempotency := fs.String("idempotency", "", "Idempotency key for writes")
	stream := fs.Bool("stream", false, "Stream SELECT rows (prints one JSON object per row)")
	g := devcli.ParseGlobalFlagsArgs(fs, args)

	defer func() {
		if r := recover(); r != nil {
			devcli.Panicf("missing required flag: %v", r)
		}
	}()

	devcli.MustNonEmpty(*projectID, "-project")
	devcli.MustNonEmpty(g.HolderID, "-holder")
	devcli.MustNonEmpty(g.ProjectName, "-pname")
	devcli.MustNonEmpty(g.APIKey, "-apikey")
	devcli.MustNonEmpty(*query, "-q")

	var params []any
	if strings.TrimSpace(*paramsJSON) != "" {
		if err := json.Unmarshal([]byte(*paramsJSON), &params); err != nil {
			return fmt.Errorf("invalid -params JSON: %w", err)
		}
	}

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	callOpts := []warlot.CallOption{}
	if *idempotency != "" {
		callOpts = append(callOpts, warlot.WithIdempotencyKey(*idempotency))
	}

	if *stream {
		sc, err := cl.ExecSQLStream(ctx, *projectID, warlot.SQLRequest{SQL: *query, Params: params}, callOpts...)
		if err != nil {
			return err
		}
		defer sc.Close()

		var row map[string]any
		for sc.Next(&row) {
			devcli.PrintJSON(row)
			row = nil
		}
		if err := sc.Err(); err != nil {
			return fmt.Errorf("stream read error: %w", err)
		}
		return nil
	}

	proj := cl.Project(*projectID)
	res, err := proj.SQL(ctx, *query, params, callOpts...)
	if err != nil {
		return err
	}
	devcli.PrintJSON(res)
	return nil
}
