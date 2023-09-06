package k8s

import (
	"context"
	"strconv"
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/workqueue"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	labeled.SetMetricKeys(contextutils.WorkflowIDKey)
}

func newMockExecutor(ctx context.Context, t testing.TB) (Executor, array.AdvanceIteration) {
	kubeClient := &mocks.KubeClient{}
	kubeClient.OnGetClient().Return(mocks.NewFakeKubeClient())
	kubeClient.OnGetCache().Return(mocks.NewFakeKubeCache())
	e, err := NewExecutor(kubeClient, &Config{
		MaxErrorStringLength: 200,
		OutputAssembler: workqueue.Config{
			Workers:            2,
			MaxRetries:         0,
			IndexCacheMaxItems: 100,
		},
		ErrorAssembler: workqueue.Config{
			Workers:            2,
			MaxRetries:         0,
			IndexCacheMaxItems: 100,
		},
	}, promutils.NewTestScope())
	assert.NoError(t, err)

	assert.NoError(t, e.Start(ctx))
	return e, func(ctx context.Context, tCtx core.TaskExecutionContext) error {
		return advancePodPhases(context.Background(), tCtx.DataStore(), tCtx.OutputWriter(), kubeClient.GetClient())
	}
}

func TestEndToEnd(t *testing.T) {
	ctx := context.Background()
	executor, iter := newMockExecutor(ctx, t)
	array.RunArrayTestsEndToEnd(t, executor, iter)
}

func advancePodPhases(ctx context.Context, store *storage.DataStore, outputWriter io.OutputWriter, runtimeClient client.Client) error {
	podList := &v1.PodList{}
	err := runtimeClient.List(ctx, podList, &client.ListOptions{
		Raw: &metav1.ListOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "pod",
				APIVersion: v1.SchemeGroupVersion.String(),
			},
		},
	})
	if err != nil {
		return err
	}

	for _, pod := range podList.Items {
		newPhase := nextHappyPodPhase(pod.Status.Phase)
		primaryContainerName := pod.Annotations["primary_container_name"]
		if len(primaryContainerName) <= 0 {
			primaryContainerName = "foo"
		}
		pod.Status.ContainerStatuses = []v1.ContainerStatus{
			v1.ContainerStatus{
				Name:        primaryContainerName,
				ContainerID: primaryContainerName,
				State: v1.ContainerState{
					Running: &v1.ContainerStateRunning{},
				},
			},
		}

		if pod.Status.Phase != newPhase && newPhase == v1.PodSucceeded {
			idx := -1
			env := pod.Spec.Containers[0].Env
			for _, v := range env {
				if v.Name == "FLYTE_K8S_ARRAY_INDEX" {
					idx, err = strconv.Atoi(v.Value)
					if err != nil {
						return err
					}

					break
				}
			}

			pod.Status.ContainerStatuses[0].State = v1.ContainerState{
				Terminated: &v1.ContainerStateTerminated{},
			}

			ref := outputWriter.GetOutputPath()
			if idx > -1 {
				ref, err = store.ConstructReference(ctx, outputWriter.GetOutputPrefixPath(), strconv.Itoa(idx), "outputs.pb")
				if err != nil {
					return err
				}
			}

			err = store.WriteProtobuf(ctx, ref, storage.Options{},
				coreutils.MustMakeLiteral(map[string]interface{}{
					"x": 5,
				}).GetMap())
			if err != nil {
				return err
			}
		}

		pod.Status.Phase = newPhase

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
