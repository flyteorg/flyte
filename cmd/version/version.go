package version

import (
	"context"
	"encoding/json"
	"fmt"

	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	stdlibversion "github.com/flyteorg/flytestdlib/version"
	"github.com/spf13/cobra"
)

// Long descriptions are whitespace sensitive when generating docs using sphinx.
const (
	versionCmdShort = `Used for fetching flyte version`
	versionCmdLong  = `
Example version.
::

 bin/flytectl version
`
)

type versionOutput struct {
	// Specifies the Name of app
	App string `json:"App,omitempty"`
	// Specifies the GIT sha of the build
	Build string `json:"Build,omitempty"`
	// Version for the build, should follow a semver
	Version string `json:"Version,omitempty"`
	// Build timestamp
	BuildTime string `json:"BuildTime,omitempty"`
}

// GetVersionCommand will return version command
func GetVersionCommand(rootCmd *cobra.Command) map[string]cmdCore.CommandEntry {
	getResourcesFuncs := map[string]cmdCore.CommandEntry{
		"version": {CmdFunc: getVersion, Aliases: []string{"versions"}, ProjectDomainNotRequired: true,
			Short: versionCmdShort,
			Long:  versionCmdLong},
	}
	return getResourcesFuncs
}

func getVersion(ctx context.Context, args []string, cmdCtx cmdCore.CommandContext) error {

	v, err := cmdCtx.AdminClient().GetVersion(ctx, &admin.GetVersionRequest{})
	if err != nil {
		return fmt.Errorf("err %v: ", err)
	}

	// Print Flytectl
	if err := printVersion(versionOutput{
		Build:     stdlibversion.Build,
		BuildTime: stdlibversion.BuildTime,
		Version:   stdlibversion.Version,
		App:       "flytectl",
	}); err != nil {
		return err
	}

	// Print Flyteadmin
	if err := printVersion(versionOutput{
		Build:     v.ControlPlaneVersion.Build,
		BuildTime: v.ControlPlaneVersion.BuildTime,
		Version:   v.ControlPlaneVersion.Version,
		App:       "controlPlane",
	}); err != nil {
		return err
	}
	return nil
}

func printVersion(response versionOutput) error {
	b, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return fmt.Errorf("err %v: ", err)
	}
	fmt.Print(string(b))
	return nil
}
