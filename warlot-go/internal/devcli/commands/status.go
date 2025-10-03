package commands

import (
	"flag"

	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli"
)

// RunStatus retrieves project status.
func RunStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
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

	out, err := cl.GetProjectStatus(ctx, *projectID)
	if err != nil {
		return err
	}
	devcli.PrintJSON(out)
	return nil
}
