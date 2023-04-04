package cmdcore

import (
	"context"
	"net/url"
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/admin"
	"github.com/flyteorg/flytestdlib/config"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func testCommandFunc(ctx context.Context, args []string, cmdCtx CommandContext) error {
	return nil
}

func TestGenerateCommandFunc(t *testing.T) {
	t.Run("dummy host name", func(t *testing.T) {
		adminCfg := admin.GetConfig(context.Background())
		adminCfg.Endpoint = config.URL{URL: url.URL{Host: "dummyHost"}}
		adminCfg.AuthType = admin.AuthTypePkce
		rootCmd := &cobra.Command{}
		cmdEntry := CommandEntry{CmdFunc: testCommandFunc, ProjectDomainNotRequired: true}
		fn := generateCommandFunc(cmdEntry)
		assert.Nil(t, fn(rootCmd, []string{}))
	})

	t.Run("host is not configured", func(t *testing.T) {
		adminCfg := admin.GetConfig(context.Background())
		adminCfg.Endpoint = config.URL{URL: url.URL{Host: ""}}
		rootCmd := &cobra.Command{}
		cmdEntry := CommandEntry{CmdFunc: testCommandFunc, ProjectDomainNotRequired: true}
		fn := generateCommandFunc(cmdEntry)
		assert.Nil(t, fn(rootCmd, []string{}))
	})
}
