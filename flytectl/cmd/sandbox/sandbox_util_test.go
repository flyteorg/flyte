package sandbox

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	u "github.com/flyteorg/flytectl/cmd/testutils"

	f "github.com/flyteorg/flytectl/pkg/filesystemutils"

	"github.com/stretchr/testify/assert"
)

var (
	cmdCtx cmdCore.CommandContext
)

func cleanup(client *client.Client) error {
	containers, err := client.ContainerList(context.Background(), types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return err
	}
	for _, v := range containers {
		if strings.Contains(v.Names[0], SandboxClusterName) {
			if err := client.ContainerRemove(context.Background(), v.ID, types.ContainerRemoveOptions{}); err != nil {
				return err
			}
		}
	}
	return nil
}

func setupSandbox() {
	mockAdminClient := u.MockClient
	cmdCtx = cmdCore.NewCommandContext(mockAdminClient, u.MockOutStream)
	_, err := os.Stat(f.FilePathJoin(f.UserHomeDir(), ".flyte"))
	if os.IsNotExist(err) {
		_ = os.MkdirAll(f.FilePathJoin(f.UserHomeDir(), ".flyte"), 0755)
	}
	_ = setupFlytectlConfig()
}

func TestConfigCleanup(t *testing.T) {
	_, err := os.Stat(f.FilePathJoin(f.UserHomeDir(), ".flyte"))
	if os.IsNotExist(err) {
		_ = os.MkdirAll(f.FilePathJoin(f.UserHomeDir(), ".flyte"), 0755)
	}
	_ = ioutil.WriteFile(FlytectlConfig, []byte("string"), 0600)
	_ = ioutil.WriteFile(Kubeconfig, []byte("string"), 0600)

	err = configCleanup()
	assert.Nil(t, err)

	_, err = os.Stat(FlytectlConfig)
	check := os.IsNotExist(err)
	assert.Equal(t, check, true)

	_, err = os.Stat(Kubeconfig)
	check = os.IsNotExist(err)
	assert.Equal(t, check, true)
	_ = configCleanup()
}

func TestSetupFlytectlConfig(t *testing.T) {
	_, err := os.Stat(f.FilePathJoin(f.UserHomeDir(), ".flyte"))
	if os.IsNotExist(err) {
		_ = os.MkdirAll(f.FilePathJoin(f.UserHomeDir(), ".flyte"), 0755)
	}
	err = setupFlytectlConfig()
	assert.Nil(t, err)
	_, err = os.Stat(FlytectlConfig)
	assert.Nil(t, err)
	check := os.IsNotExist(err)
	assert.Equal(t, check, false)
	_ = configCleanup()
}

func TestTearDownSandbox(t *testing.T) {
	setupSandbox()
	cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	err := teardownSandboxCluster(context.Background(), []string{}, cmdCtx)
	assert.Nil(t, err)
	assert.Nil(t, cleanup(cli))
}

func TestStartContainer(t *testing.T) {
	setupSandbox()
	cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	assert.Nil(t, cleanup(cli))
	err := startSandboxCluster(context.Background(), []string{}, cmdCtx)
	assert.Nil(t, err)
}
