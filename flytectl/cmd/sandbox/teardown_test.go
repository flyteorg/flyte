package sandbox

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/flyteorg/flyte/flytectl/cmd/testutils"
	"github.com/flyteorg/flyte/flytectl/pkg/configutil"
	"github.com/flyteorg/flyte/flytectl/pkg/docker"
	"github.com/flyteorg/flyte/flytectl/pkg/docker/mocks"
	"github.com/flyteorg/flyte/flytectl/pkg/k8s"
	k8sMocks "github.com/flyteorg/flyte/flytectl/pkg/k8s/mocks"
	"github.com/flyteorg/flyte/flytectl/pkg/util"
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
	mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return(containers, nil)
	mockDocker.OnContainerRemove(ctx, mock.Anything, container.RemoveOptions{Force: true}).Return(nil)
	mockK8sContextMgr := &k8sMocks.ContextOps{}
	mockK8sContextMgr.OnRemoveContext(mock.Anything).Return(nil)
	k8s.ContextMgr = mockK8sContextMgr

	docker.Client = mockDocker
	err := teardownSandboxCluster(ctx, []string{}, s.CmdCtx)
	assert.Nil(t, err)
}
