package upgrade

import (
	"sort"
	"testing"

	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flytectl/cmd/testutils"
	"github.com/flyteorg/flyte/flytectl/pkg/github"
	"github.com/flyteorg/flyte/flytectl/pkg/platformutil"
	"github.com/flyteorg/flyte/flytectl/pkg/util"
	stdlibversion "github.com/flyteorg/flyte/flytestdlib/version"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var (
	version = "v0.2.20"
	tempExt = "flyte.ext"
)

func TestUpgradeCommand(t *testing.T) {
	rootCmd := &cobra.Command{
		Long:              "Flytectl is a CLI tool written in Go to interact with the FlyteAdmin service.",
		Short:             "Flytectl CLI tool",
		Use:               "flytectl",
		DisableAutoGenTag: true,
	}
	upgradeCmd := SelfUpgrade(rootCmd)
	cmdCore.AddCommands(rootCmd, upgradeCmd)
	assert.Equal(t, len(rootCmd.Commands()), 1)
	cmdNouns := rootCmd.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})

	assert.Equal(t, cmdNouns[0].Use, "upgrade")
	assert.Equal(t, cmdNouns[0].Short, upgradeCmdShort)
	assert.Equal(t, cmdNouns[0].Long, upgradeCmdLong)
}

func TestUpgrade(t *testing.T) {
	_ = util.WriteIntoFile([]byte("data"), tempExt)
	stdlibversion.Version = version
	github.FlytectlReleaseConfig.OverrideExecutable = tempExt
	t.Run("Successful upgrade", func(t *testing.T) {
		message, err := upgrade(github.FlytectlReleaseConfig)
		assert.Nil(t, err)
		assert.Contains(t, message, "Successfully updated to version")
	})
}

func TestCheckGoosForRollback(t *testing.T) {
	stdlibversion.Version = version
	linux := platformutil.Linux
	windows := platformutil.Windows
	darwin := platformutil.Darwin
	github.FlytectlReleaseConfig.OverrideExecutable = tempExt
	t.Run("checkGOOSForRollback on linux", func(t *testing.T) {
		assert.Equal(t, true, isRollBackSupported(linux))
		assert.Equal(t, false, isRollBackSupported(windows))
		assert.Equal(t, true, isRollBackSupported(darwin))
	})
}

func TestIsUpgradeable(t *testing.T) {
	stdlibversion.Version = version
	github.FlytectlReleaseConfig.OverrideExecutable = tempExt
	linux := platformutil.Linux
	windows := platformutil.Windows
	darwin := platformutil.Darwin
	t.Run("IsUpgradeable on linux", func(t *testing.T) {
		check, err := isUpgradeSupported(linux)
		assert.Nil(t, err)
		assert.Equal(t, true, check)
	})
	t.Run("IsUpgradeable on darwin", func(t *testing.T) {
		check, err := isUpgradeSupported(darwin)
		assert.Nil(t, err)
		assert.Equal(t, true, check)
	})
	t.Run("IsUpgradeable on darwin using brew", func(t *testing.T) {
		check, err := isUpgradeSupported(darwin)
		assert.Nil(t, err)
		assert.Equal(t, true, check)
	})
	t.Run("isUpgradeSupported failed", func(t *testing.T) {
		stdlibversion.Version = "v"
		check, err := isUpgradeSupported(linux)
		assert.NotNil(t, err)
		assert.Equal(t, false, check)
		stdlibversion.Version = version
	})
	t.Run("isUpgradeSupported windows", func(t *testing.T) {
		check, err := isUpgradeSupported(windows)
		assert.Nil(t, err)
		assert.Equal(t, false, check)
	})
}

func TestSelfUpgrade(t *testing.T) {
	stdlibversion.Version = version
	github.FlytectlReleaseConfig.OverrideExecutable = tempExt
	goos = platformutil.Linux
	t.Run("Successful upgrade", func(t *testing.T) {
		s := testutils.Setup()
		stdlibversion.Build = ""
		stdlibversion.BuildTime = ""
		stdlibversion.Version = version

		assert.Nil(t, selfUpgrade(s.Ctx, []string{}, s.CmdCtx))
	})
}

func TestSelfUpgradeError(t *testing.T) {
	stdlibversion.Version = version
	github.FlytectlReleaseConfig.OverrideExecutable = tempExt
	goos = platformutil.Linux
	t.Run("Successful upgrade", func(t *testing.T) {
		s := testutils.Setup()
		stdlibversion.Build = ""
		stdlibversion.BuildTime = ""
		stdlibversion.Version = "v"

		assert.NotNil(t, selfUpgrade(s.Ctx, []string{}, s.CmdCtx))
	})

}

func TestSelfUpgradeRollback(t *testing.T) {
	stdlibversion.Version = version
	github.FlytectlReleaseConfig.OverrideExecutable = tempExt
	goos = platformutil.Linux
	t.Run("Successful rollback", func(t *testing.T) {
		s := testutils.Setup()
		var args = []string{rollBackSubCommand}
		stdlibversion.Build = ""
		stdlibversion.BuildTime = ""
		stdlibversion.Version = version
		assert.Nil(t, selfUpgrade(s.Ctx, args, s.CmdCtx))
	})

	t.Run("Successful rollback failed", func(t *testing.T) {
		s := testutils.Setup()
		var args = []string{rollBackSubCommand}
		stdlibversion.Build = ""
		stdlibversion.BuildTime = ""
		stdlibversion.Version = "v100.0.0"
		assert.NotNil(t, selfUpgrade(s.Ctx, args, s.CmdCtx))
	})

	t.Run("Successful rollback for windows", func(t *testing.T) {
		s := testutils.Setup()
		var args = []string{rollBackSubCommand}
		stdlibversion.Build = ""
		stdlibversion.BuildTime = ""
		stdlibversion.Version = version
		goos = platformutil.Windows
		assert.Nil(t, selfUpgrade(s.Ctx, args, s.CmdCtx))
	})

	t.Run("Successful rollback for windows", func(t *testing.T) {
		s := testutils.Setup()
		var args = []string{rollBackSubCommand}
		stdlibversion.Build = ""
		stdlibversion.BuildTime = ""
		stdlibversion.Version = version
		github.FlytectlReleaseConfig.OverrideExecutable = "/"
		assert.Nil(t, selfUpgrade(s.Ctx, args, s.CmdCtx))
	})

}
