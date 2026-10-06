package k8s

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	structpb "google.golang.org/protobuf/types/known/structpb"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	idlCore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/workqueue"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	"github.com/flyteorg/stow/local"
)

func init() {
	labeled.SetMetricKeys(contextutils.WorkflowIDKey)
}

func newMockExecutor(ctx context.Context, t testing.TB) (Executor, array.AdvanceIteration) {
	kubeClient := &mocks.KubeClient{}
	kubeClient.EXPECT().GetClient().Return(mocks.NewFakeKubeClient())
	kubeClient.EXPECT().GetCache().Return(mocks.NewFakeKubeCache())
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

type fileTaskContext struct {
	core.TaskExecutionContext
	store     *storage.DataStore
	prefix    storage.DataReference
	statePath string
}

func (c *fileTaskContext) TaskReader() core.TaskReader               { return c }
func (c *fileTaskContext) DataStore() *storage.DataStore             { return c.store }
func (c *fileTaskContext) PluginStateReader() core.PluginStateReader { return c }
func (c *fileTaskContext) PluginStateWriter() core.PluginStateWriter { return c }
func (c *fileTaskContext) GetStateVersion() uint8                    { return 0 }
func (c *fileTaskContext) Path(context.Context) (storage.DataReference, error) {
	return c.prefix + "/task.pb", nil
}
func (c *fileTaskContext) Read(ctx context.Context) (*idlCore.TaskTemplate, error) {
	task := &idlCore.TaskTemplate{}
	return task, c.store.ReadProtobuf(ctx, c.prefix+"/task.pb", task)
}
func (c *fileTaskContext) InputReader() io.InputReader {
	ctx := context.Background()
	return ioutils.NewRemoteFileInputReader(ctx, c.store, ioutils.NewInputFilePaths(ctx, c.store, c.prefix))
}
func (c *fileTaskContext) OutputWriter() io.OutputWriter {
	ctx := context.Background()
	raw := ioutils.NewRawOutputPaths(ctx, c.prefix+"/raw")
	paths := ioutils.NewCheckpointRemoteFilePaths(ctx, c.store, c.prefix, raw, "")
	return ioutils.NewRemoteFileOutputWriter(ctx, c.store, paths)
}
func (c *fileTaskContext) Get(value interface{}) (uint8, error) {
	data, err := os.ReadFile(c.statePath)
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return 0, json.Unmarshal(data, value)
}
func (c *fileTaskContext) Put(_ uint8, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return os.WriteFile(c.statePath, data, 0600)
}
func (c *fileTaskContext) Reset() error { return os.Remove(c.statePath) }

func TestExecutorPodBuildError(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	store, err := storage.NewDataStore(&storage.Config{
		Type: storage.TypeLocal, InitContainer: "repro",
		Stow: storage.StowConfig{Kind: local.Kind, Config: map[string]string{local.ConfigKeyPath: root}},
	}, promutils.NewTestScope())
	require.NoError(t, err)
	podSpec, err := utils.MarshalObjToStruct(&v1.PodSpec{
		Containers: []v1.Container{{Name: "main", Image: "busybox"}},
	})
	require.NoError(t, err)
	custom, err := structpb.NewStruct(map[string]interface{}{"parallelism": "1", "minSuccessRatio": 1.0})
	require.NoError(t, err)
	task := &idlCore.TaskTemplate{
		Type: "container_array", TaskTypeVersion: 1, Custom: custom,
		Target: &idlCore.TaskTemplate_K8SPod{K8SPod: &idlCore.K8SPod{PodSpec: podSpec}},
	}
	tCtx := &fileTaskContext{
		TaskExecutionContext: getMockTaskExecutionContext(ctx, 1),
		store:                store,
		prefix:               "file://repro",
		statePath:            filepath.Join(root, "state.json"),
	}
	inputs := coreutils.MustMakeLiteral(map[string]interface{}{"x": []interface{}{1}}).GetMap()
	require.NoError(t, store.WriteProtobuf(ctx, tCtx.prefix+"/task.pb", storage.Options{}, task))
	require.NoError(t, store.WriteProtobuf(ctx, tCtx.prefix+"/inputs.pb", storage.Options{}, inputs))
	executor, err := NewExecutor(nil, GetConfig(), promutils.NewTestScope())
	require.NoError(t, err)

	for round := 1; round <= 4; round++ {
		transition, err := executor.Handle(ctx, tCtx)
		if round == 4 {
			require.ErrorContains(t, err, "config missing [primary_container_name] key")
			t.Logf("round=%d returned_error=%s", round, err)
		} else {
			require.NoError(t, err)
			t.Logf("round=%d phase=%s", round, transition.Info().Phase())
		}
	}
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
	case v1.PodPending:
		fallthrough
	case "":
		return v1.PodRunning
	case v1.PodRunning:
		return v1.PodSucceeded
	case v1.PodSucceeded:
		return v1.PodSucceeded
	}

	return ""
}
