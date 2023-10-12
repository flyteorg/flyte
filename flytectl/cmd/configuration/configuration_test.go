package configuration

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"testing"

	admin2 "github.com/flyteorg/flyteidl/clients/go/admin"

	"github.com/flyteorg/flytectl/pkg/configutil"

	initConfig "github.com/flyteorg/flytectl/cmd/config/subcommand/config"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/pkg/util"

	"github.com/stretchr/testify/assert"
)

func TestCreateInitCommand(t *testing.T) {
	configCmd := CreateConfigCommand()
	assert.Equal(t, configCmd.Use, "config")
	assert.Equal(t, configCmd.Short, "Runs various config commands, look at the help of this command to get a list of available commands..")
	fmt.Println(configCmd.Commands())
	assert.Equal(t, 4, len(configCmd.Commands()))
	cmdNouns := configCmd.Commands()
	// Sort by Use value.
	sort.Slice(cmdNouns, func(i, j int) bool {
		return cmdNouns[i].Use < cmdNouns[j].Use
	})

	assert.Equal(t, "discover", cmdNouns[0].Use)
	assert.Equal(t, "Searches for a config in one of the default search paths.", cmdNouns[0].Short)
	assert.Equal(t, "docs", cmdNouns[1].Use)
	assert.Equal(t, "Generate configuration documetation in rst format", cmdNouns[1].Short)

	assert.Equal(t, "init", cmdNouns[2].Use)
	assert.Equal(t, initCmdShort, cmdNouns[2].Short)
	assert.Equal(t, "validate", cmdNouns[3].Use)
	assert.Equal(t, "Validates the loaded config.", cmdNouns[3].Short)
}

func TestSetupConfigFunc(t *testing.T) {
	var yes = strings.NewReader("Yes")
	var no = strings.NewReader("No")
	var empty = strings.NewReader("")
	mockOutStream := new(io.Writer)
	ctx := context.Background()
	_ = os.Remove(configutil.FlytectlConfig)

	_ = util.SetupFlyteDir()

	mockClient := admin2.InitializeMockClientset()
	cmdCtx := cmdCore.NewCommandContext(mockClient, *mockOutStream)
	err := configInitFunc(ctx, []string{}, cmdCtx)
	initConfig.DefaultConfig.Host = ""
	assert.Nil(t, err)

	initConfig.DefaultConfig.Force = false
	assert.Nil(t, initFlytectlConfig(yes))
	assert.Nil(t, initFlytectlConfig(no))

	initConfig.DefaultConfig.Force = true
	assert.Nil(t, initFlytectlConfig(empty))

	initConfig.DefaultConfig.Host = "flyte.org"
	assert.Nil(t, initFlytectlConfig(no))
	initConfig.DefaultConfig.Host = "localhost:30081"
	assert.Nil(t, initFlytectlConfig(no))
	assert.Nil(t, initFlytectlConfig(yes))
}

func TestTrimFunc(t *testing.T) {
	assert.Equal(t, trimEndpoint("dns:///localhost"), "localhost")
	assert.Equal(t, trimEndpoint("http://localhost"), "localhost")
	assert.Equal(t, trimEndpoint("https://localhost"), "localhost")
}

func TestValidateEndpointName(t *testing.T) {
	assert.Equal(t, true, validateEndpointName("8093405779.ap-northeast-2.elb.amazonaws.com:81"))
	assert.Equal(t, true, validateEndpointName("8093405779.ap-northeast-2.elb.amazonaws.com"))
	assert.Equal(t, false, validateEndpointName("8093405779.ap-northeast-2.elb.amazonaws.com:81/console"))
	assert.Equal(t, true, validateEndpointName("localhost"))
	assert.Equal(t, true, validateEndpointName("127.0.0.1"))
	assert.Equal(t, true, validateEndpointName("127.0.0.1:30086"))
	assert.Equal(t, true, validateEndpointName("112.11.1.1"))
	assert.Equal(t, true, validateEndpointName("112.11.1.1:8080"))
	assert.Equal(t, false, validateEndpointName("112.11.1.1:8080/console"))
	assert.Equal(t, false, validateEndpointName("flyte"))
}

func TestForceFlagInCreateConfigCommand(t *testing.T) {
	cmd := CreateConfigCommand()
	assert.False(t, initConfig.DefaultConfig.Force)
	err := cmd.Flags().Parse([]string{"--force"})
	assert.Nil(t, err)
	assert.True(t, initConfig.DefaultConfig.Force)
}
