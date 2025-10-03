package commands

import (
	"flag"

	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli"
	"github.com/steven3002/warlot-golang-sdk/warlot-go/warlot"
)

// RunResolve resolves a project by holder and project name.
func RunResolve(args []string) error {
	fs := flag.NewFlagSet("resolve", flag.ContinueOnError)
	g := devcli.ParseGlobalFlagsArgs(fs, args)

	defer func() {
		if r := recover(); r != nil {
			devcli.Panicf("missing required flag: %v", r)
		}
	}()

	devcli.MustNonEmpty(g.HolderID, "-holder")
	devcli.MustNonEmpty(g.ProjectName, "-pname")

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	out, err := cl.ResolveProject(ctx, warlot.ResolveProjectRequest{
		HolderID:    g.HolderID,
		ProjectName: g.ProjectName,
	})
	if err != nil {
		return err
	}
	devcli.PrintJSON(out)
	return nil
}
