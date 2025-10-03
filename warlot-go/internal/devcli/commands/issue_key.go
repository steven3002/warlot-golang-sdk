package commands

import (
	"flag"

	"github.com/steven3002/warlot-golang-sdk/warlot-go/internal/devcli"
	"github.com/steven3002/warlot-golang-sdk/warlot-go/warlot"
)

// RunIssueKey issues an API key for a project.
func RunIssueKey(args []string) error {
	fs := flag.NewFlagSet("issue-key", flag.ContinueOnError)
	projectID := fs.String("project", "", "Project ID")
	userAddr := fs.String("user", "", "User address (owner)")
	g := devcli.ParseGlobalFlagsArgs(fs, args)

	defer func() {
		if r := recover(); r != nil {
			devcli.Panicf("missing required flag: %v", r)
		}
	}()

	devcli.MustNonEmpty(*projectID, "-project")
	devcli.MustNonEmpty(g.HolderID, "-holder")
	devcli.MustNonEmpty(g.ProjectName, "-pname")
	devcli.MustNonEmpty(*userAddr, "-user")

	cl := devcli.NewClient(g)
	ctx, cancel := devcli.Ctx(g)
	defer cancel()

	out, err := cl.IssueAPIKey(ctx, warlot.IssueKeyRequest{
		ProjectID:     *projectID,
		ProjectHolder: g.HolderID,
		ProjectName:   g.ProjectName,
		User:          *userAddr,
	})
	if err != nil {
		return err
	}
	devcli.PrintJSON(out)
	return nil
}
