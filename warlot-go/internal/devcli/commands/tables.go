package commands

import (
	"errors"
	"flag"

	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli"
)

// RunTables dispatches to list|browse|schema|count subcommands.
func RunTables(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: warlotdev tables <list|browse|schema|count> [flags]")
	}
	switch args[0] {
	case "list":
		return runTablesList(args[1:])
	case "browse":
		return runTablesBrowse(args[1:])
	case "schema":
		return runTablesSchema(args[1:])
	case "count":
		return runTablesCount(args[1:])
	default:
		return errors.New("unknown tables subcommand; use list|browse|schema|count")
	}
}

func runTablesList(args []string) error {
	fs := flag.NewFlagSet("tables list", flag.ContinueOnError)
	projectID := fs.String("project", "", "Project ID")
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

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	out, err := cl.ListTables(ctx, *projectID)
	if err != nil {
		return err
	}
	devcli.PrintJSON(out)
	return nil
}

func runTablesBrowse(args []string) error {
	fs := flag.NewFlagSet("tables browse", flag.ContinueOnError)
	projectID := fs.String("project", "", "Project ID")
	table := fs.String("table", "", "Table name")
	limit := fs.Int("limit", 10, "Limit")
	offset := fs.Int("offset", 0, "Offset")
	g := devcli.ParseGlobalFlagsArgs(fs, args)

	defer func() {
		if r := recover(); r != nil {
			devcli.Panicf("missing required flag: %v", r)
		}
	}()

	devcli.MustNonEmpty(*projectID, "-project")
	devcli.MustNonEmpty(*table, "-table")
	devcli.MustNonEmpty(g.HolderID, "-holder")
	devcli.MustNonEmpty(g.ProjectName, "-pname")
	devcli.MustNonEmpty(g.APIKey, "-apikey")

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	out, err := cl.BrowseRows(ctx, *projectID, *table, *limit, *offset)
	if err != nil {
		return err
	}
	devcli.PrintJSON(out)
	return nil
}

func runTablesSchema(args []string) error {
	fs := flag.NewFlagSet("tables schema", flag.ContinueOnError)
	projectID := fs.String("project", "", "Project ID")
	table := fs.String("table", "", "Table name")
	g := devcli.ParseGlobalFlagsArgs(fs, args)

	defer func() {
		if r := recover(); r != nil {
			devcli.Panicf("missing required flag: %v", r)
		}
	}()

	devcli.MustNonEmpty(*projectID, "-project")
	devcli.MustNonEmpty(*table, "-table")
	devcli.MustNonEmpty(g.HolderID, "-holder")
	devcli.MustNonEmpty(g.ProjectName, "-pname")
	devcli.MustNonEmpty(g.APIKey, "-apikey")

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	out, err := cl.GetTableSchema(ctx, *projectID, *table)
	if err != nil {
		return err
	}
	devcli.PrintJSON(out)
	return nil
}

func runTablesCount(args []string) error {
	fs := flag.NewFlagSet("tables count", flag.ContinueOnError)
	projectID := fs.String("project", "", "Project ID")
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

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	out, err := cl.GetTableCount(ctx, *projectID)
	if err != nil {
		return err
	}
	devcli.PrintJSON(out)
	return nil
}
