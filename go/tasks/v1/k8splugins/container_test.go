package k8splugins

import (
	"context"
	"testing"

	"github.com/lyft/flyteplugins/go/tasks/v1/errors"
	"github.com/lyft/flyteplugins/go/tasks/v1/flytek8s"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/lyft/flytestdlib/storage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lyft/flyteplugins/go/tasks/v1/types/mocks"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/lyft/flyteplugins/go/tasks/v1/types"
)

var resourceRequirements = &v1.ResourceRequirements{
	Limits: v1.ResourceList{
		v1.ResourceCPU:     resource.MustParse("1024m"),
		v1.ResourceStorage: resource.MustParse("100M"),
	},
}

func dummyContainerTaskContext(resources *v1.ResourceRequirements) types.TaskContext {
	taskCtx := &mocks.TaskContext{}
	taskCtx.On("GetNamespace").Return("test-namespace")
	taskCtx.On("GetAnnotations").Return(map[string]string{"annotation-1": "val1"})
	taskCtx.On("GetLabels").Return(map[string]string{"label-1": "val1"})
	taskCtx.On("GetOwnerReference").Return(metav1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskCtx.On("GetDataDir").Return(storage.DataReference("/data/"))
	taskCtx.On("GetInputsFile").Return(storage.DataReference("/input"))
	taskCtx.On("GetK8sServiceAccount").Return("service-account")

	tID := &mocks.TaskExecutionID{}
	tID.On("GetID").Return(core.TaskExecutionIdentifier{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
	})
	tID.On("GetGeneratedName").Return("some-acceptable-name")
	taskCtx.On("GetTaskExecutionID").Return(tID)

	to := &mocks.TaskOverrides{}
	to.On("GetResources").Return(resources)
	taskCtx.On("GetOverrides").Return(to)

	return taskCtx
}

func TestContainerTaskExecutor_BuildIdentityResource(t *testing.T) {
	c := containerTaskExecutor{}
	x := &mocks.TaskContext{}
	r, err := c.BuildIdentityResource(context.TODO(), x)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	_, ok := r.(*v1.Pod)
	assert.True(t, ok)
	assert.Equal(t, flytek8s.PodKind, r.GetObjectKind().GroupVersionKind().Kind)
}

func TestContainerTaskExecutor_BuildResource(t *testing.T) {
	c := containerTaskExecutor{}
	x := dummyContainerTaskContext(resourceRequirements)
	command := []string{"command"}
	args := []string{"{{.Input}}"}

	task := &core.TaskTemplate{
		Type: "test",
		Target: &core.TaskTemplate_Container{
			Container: &core.Container{
				Command: command,
				Args:    args,
			},
		},
	}
	r, err := c.BuildResource(context.TODO(), x, task, nil)
	assert.NoError(t, err)
	assert.NotNil(t, r)
	j, ok := r.(*v1.Pod)
	assert.True(t, ok)

	assert.NotEmpty(t, j.Spec.Containers)
	assert.Equal(t, resourceRequirements.Limits[v1.ResourceCPU], j.Spec.Containers[0].Resources.Limits[v1.ResourceCPU])

	// TODO: Once configurable, test when setting storage is supported on the cluster vs not.
	storageRes := j.Spec.Containers[0].Resources.Limits[v1.ResourceStorage]
	assert.Equal(t, int64(0), (&storageRes).Value())

	assert.Equal(t, command, j.Spec.Containers[0].Command)
	assert.Equal(t, []string{"/input"}, j.Spec.Containers[0].Args)

	assert.Equal(t, "service-account", j.Spec.ServiceAccountName)
}

func TestContainerTaskExecutor_GetTaskStatus(t *testing.T) {
	c := containerTaskExecutor{}
	j := &v1.Pod{
		Status: v1.PodStatus{},
	}

	ctx := context.TODO()
	t.Run("running", func(t *testing.T) {
		s, i, err := c.GetTaskStatus(ctx, nil, j)
		assert.NoError(t, err)
		assert.NotNil(t, i)
		assert.Equal(t, types.TaskPhaseRunning, s.Phase)
	})

	t.Run("queued", func(t *testing.T) {
		j.Status.Phase = v1.PodPending
		s, i, err := c.GetTaskStatus(ctx, nil, j)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseQueued, s.Phase)
		assert.Nil(t, i)
	})

	t.Run("failNoCondition", func(t *testing.T) {
		j.Status.Phase = v1.PodFailed
		s, i, err := c.GetTaskStatus(ctx, nil, j)
		assert.NoError(t, err)
		assert.NotNil(t, i)
		assert.Equal(t, types.TaskPhaseRetryableFailure, s.Phase)
		ec, ok := errors.GetErrorCode(s.Err)
		assert.True(t, ok)
		assert.Equal(t, errors.TaskFailedUnknownError, ec)
	})

	t.Run("failConditionUnschedulable", func(t *testing.T) {
		j.Status.Phase = v1.PodFailed
		j.Status.Reason = "Unschedulable"
		j.Status.Message = "some message"
		j.Status.Conditions = []v1.PodCondition{
			{
				Type: v1.PodReasonUnschedulable,
			},
		}
		s, i, err := c.GetTaskStatus(ctx, nil, j)
		assert.NoError(t, err)
		assert.NotNil(t, i)
		assert.Equal(t, types.TaskPhaseRetryableFailure, s.Phase)
		ec, ok := errors.GetErrorCode(s.Err)
		assert.True(t, ok)
		assert.Equal(t, "Unschedulable", ec)
	})

	t.Run("success", func(t *testing.T) {
		j.Status.Phase = v1.PodSucceeded
		s, i, err := c.GetTaskStatus(ctx, nil, j)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseSucceeded, s.Phase)
		assert.NotNil(t, i)
	})
}

func TestConvertPodFailureToError(t *testing.T) {
	t.Run("unknown-error", func(t *testing.T) {
		err := ConvertPodFailureToError(v1.PodStatus{})
		assert.Error(t, err)
		ec, ok := errors.GetErrorCode(err)
		assert.True(t, ok)
		assert.Equal(t, ec, errors.TaskFailedUnknownError)
	})

	t.Run("known-error", func(t *testing.T) {
		err := ConvertPodFailureToError(v1.PodStatus{Reason: "hello"})
		assert.Error(t, err)
		ec, ok := errors.GetErrorCode(err)
		assert.True(t, ok)
		assert.Equal(t, ec, "hello")
	})
}

func advancePodPhases(ctx context.Context, runtimeClient client.Client) error {
	podList := &v1.PodList{}
	err := runtimeClient.List(ctx, &client.ListOptions{
		Raw: &metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       flytek8s.PodKind,
				APIVersion: v1.SchemeGroupVersion.String(),
			},
		},
	}, podList)
	if err != nil {
		return err
	}

	for _, pod := range podList.Items {
		pod.Status.Phase = nextHappyPodPhase(pod.Status.Phase)
		pod.Status.ContainerStatuses = []v1.ContainerStatus{
			{ContainerID: "cont_123"},
		}
		err = runtimeClient.Update(ctx, pod.DeepCopy())
		if err != nil {
			return err
		}
	}

	return nil
}

func nextHappyPodPhase(phase v1.PodPhase) v1.PodPhase {
	switch phase {
	case v1.PodUnknown:
		fallthrough
	case v1.PodPending:
		fallthrough
	case "":
		return v1.PodRunning
	case v1.PodRunning:
		return v1.PodSucceeded
	case v1.PodSucceeded:
		return v1.PodSucceeded
	}

	return v1.PodUnknown
}
