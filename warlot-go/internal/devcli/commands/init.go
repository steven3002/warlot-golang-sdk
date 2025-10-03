package commands

import (
	"flag"

	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli"
	"github.com/steven3002/warlot-golang-sdk/warlot-go/warlot"
)

// RunInit initializes a new project.
func RunInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	owner := fs.String("owner", "", "Owner address (user)")
	includePass := fs.Bool("include-pass", true, "Include pass artifacts")
	deletable := fs.Bool("deletable", true, "Deletable project")
	g := devcli.ParseGlobalFlagsArgs(fs, args)

	defer func() {
		if r := recover(); r != nil {
			devcli.Panicf("missing required flag: %v", r)
		}
	}()

	devcli.MustNonEmpty(g.HolderID, "-holder")
	devcli.MustNonEmpty(g.ProjectName, "-pname")
	devcli.MustNonEmpty(*owner, "-owner")

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	out, err := cl.InitProject(ctx, warlot.InitProjectRequest{
		HolderID:      g.HolderID,
		ProjectName:   g.ProjectName,
		OwnerAddress:  *owner,
		EpochSet:      0,
		CycleEnd:      0,
		WritersLen:    0,
		TrackBackLen:  0,
		DraftEpochDur: 0,
		IncludePass:   *includePass,
		Deletable:     *deletable,
	})
	if err != nil {
		return err
	}
	devcli.PrintJSON(out)
	return nil
}
