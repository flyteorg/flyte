package sandbox

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flytectl/pkg/configutil"
	"github.com/flyteorg/flytectl/pkg/docker"
	"github.com/flyteorg/flytectl/pkg/docker/mocks"
	"github.com/flyteorg/flytectl/pkg/k8s"
	k8sMocks "github.com/flyteorg/flytectl/pkg/k8s/mocks"
	"github.com/flyteorg/flytectl/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTearDownClusterFunc(t *testing.T) {
	var containers []types.Container
	_ = util.SetupFlyteDir()
	_ = util.WriteIntoFile([]byte("data"), configutil.FlytectlConfig)
	s := testutils.Setup()
	ctx := s.Ctx
	mockDocker := &mocks.Docker{}
	mockDocker.OnContainerList(ctx, types.ContainerListOptions{All: true}).Return(containers, nil)
	mockDocker.OnContainerRemove(ctx, mock.Anything, types.ContainerRemoveOptions{Force: true}).Return(nil)
	mockK8sContextMgr := &k8sMocks.ContextOps{}
	mockK8sContextMgr.OnRemoveContext(mock.Anything).Return(nil)
	k8s.ContextMgr = mockK8sContextMgr

	docker.Client = mockDocker
	err := teardownSandboxCluster(ctx, []string{}, s.CmdCtx)
	assert.Nil(t, err)
}
