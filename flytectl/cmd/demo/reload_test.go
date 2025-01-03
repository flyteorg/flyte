package demo

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	cmdCore "github.com/flyteorg/flyte/flytectl/cmd/core"
	"github.com/flyteorg/flyte/flytectl/pkg/docker"
	"github.com/flyteorg/flyte/flytectl/pkg/docker/mocks"
	"github.com/flyteorg/flyte/flytectl/pkg/k8s"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

var fakePod = corev1.Pod{
	Status: corev1.PodStatus{
		Phase:      corev1.PodRunning,
		Conditions: []corev1.PodCondition{},
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:   "dummyflytepod",
		Labels: map[string]string{"app.kubernetes.io/name": "flyte-binary"},
	},
}

func sandboxSetup(ctx context.Context, legacy bool) {
	mockDocker := &mocks.Docker{}
	docker.Client = mockDocker
	mockDocker.OnContainerList(ctx, container.ListOptions{All: true}).Return([]types.Container{
		{
			ID: docker.FlyteSandboxClusterName,
			Names: []string{
				docker.FlyteSandboxClusterName,
			},
		},
	}, nil)

	// This first set of mocks is for the check for the bootstrap agent. This is
	// Expected to fail in legacy sandboxes
	var checkLegacySandboxExecExitCode int
	if legacy {
		checkLegacySandboxExecExitCode = 1
	}
	mockDocker.OnContainerExecCreateMatch(
		ctx,
		docker.FlyteSandboxClusterName,
		types.ExecConfig{
			AttachStderr: true,
			Tty:          true,
			WorkingDir:   "/",
			AttachStdout: true,
			Cmd:          []string{"sh", "-c", fmt.Sprintf("which %s > /dev/null", internalBootstrapAgent)},
		},
	).Return(types.IDResponse{ID: "0"}, nil)
	mockDocker.OnContainerExecAttachMatch(ctx, "0", types.ExecStartCheck{}).Return(types.HijackedResponse{
		Reader: bufio.NewReader(bytes.NewReader([]byte{})),
	}, nil)
	mockDocker.OnContainerExecInspectMatch(ctx, "0").Return(types.ContainerExecInspect{ExitCode: checkLegacySandboxExecExitCode}, nil)

	// Register additional mocks for the actual execution of the bootstrap agent
	// in non-legacy sandboxes
	if !legacy {
		mockDocker.OnContainerExecCreateMatch(
			ctx,
			docker.FlyteSandboxClusterName,
			types.ExecConfig{
				AttachStderr: true,
				Tty:          true,
				WorkingDir:   "/",
				AttachStdout: true,
				Cmd:          []string{internalBootstrapAgent},
			},
		).Return(types.IDResponse{ID: "1"}, nil)
		mockDocker.OnContainerExecAttachMatch(ctx, "1", types.ExecStartCheck{}).Return(types.HijackedResponse{
			Reader: bufio.NewReader(bytes.NewReader([]byte{})),
		}, nil)
	}
}

func TestReloadLegacy(t *testing.T) {
	ctx := context.Background()
	commandCtx := cmdCore.CommandContext{}
	sandboxSetup(ctx, false)
	err := reloadDemoCluster(ctx, []string{}, commandCtx)
	assert.Nil(t, err)
}

func TestDemoReloadLegacy(t *testing.T) {
	ctx := context.Background()
	commandCtx := cmdCore.CommandContext{}
	sandboxSetup(ctx, true)
	t.Run("No errors", func(t *testing.T) {
		client := testclient.NewSimpleClientset()
		_, err := client.CoreV1().Pods("flyte").Create(ctx, &fakePod, v1.CreateOptions{})
		assert.NoError(t, err)
		k8s.Client = client
		err = reloadDemoCluster(ctx, []string{}, commandCtx)
		assert.NoError(t, err)
	})

	t.Run("Multiple pods will error", func(t *testing.T) {
		client := testclient.NewSimpleClientset()
		_, err := client.CoreV1().Pods("flyte").Create(ctx, &fakePod, v1.CreateOptions{})
		assert.NoError(t, err)
		fakePod.SetName("othername")
		_, err = client.CoreV1().Pods("flyte").Create(ctx, &fakePod, v1.CreateOptions{})
		assert.NoError(t, err)
		k8s.Client = client
		err = reloadDemoCluster(ctx, []string{}, commandCtx)
		assert.Errorf(t, err, "should only have one pod")
	})
}
