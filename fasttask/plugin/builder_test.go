package plugin

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/unionai/flyte/fasttask/plugin/interfaces"
	"github.com/unionai/flyte/fasttask/plugin/pb"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils/secrets"

	flytestdlibConfig "github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/promutils"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	coremocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	pluginsCoreMock "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

type kubeClient struct {
	client.Client

	createCalls int
	deleteCalls int

	returnErrorOnCreate error
}

func (k *kubeClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	k.createCalls++
	if k.returnErrorOnCreate != nil {
		return k.returnErrorOnCreate
	}

	return nil
}

func (k *kubeClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	k.deleteCalls++
	return nil
}

type kubeCache struct {
	cache.Cache

	pods []v1.Pod
}

func (k *kubeCache) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	for _, pod := range k.pods {
		if pod.Name == key.Name {
			return nil
		}
	}

	return errors.NewNotFound(v1.Resource("pods"), key.Name)
}

func (k *kubeCache) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	if podList, ok := list.(*v1.PodList); ok {
		podList.Items = k.pods
	}
	return nil
}

func TestGetFastTaskEnvironmentSpec(t *testing.T) {
	ctx := context.Background()

	namespace := "testNamespace"
	containerName := "testContainer"

	providedPod := &v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{Name: containerName}},
		},
	}
	providedBytes, err := json.Marshal(providedPod)
	require.NoError(t, err)

	tests := []struct {
		name          string
		executionEnv  *idlcore.ExecutionEnv
		expectError   bool
		errorContains string
	}{
		{
			name: "provided PodTemplateSpec",
			executionEnv: &idlcore.ExecutionEnv{
				Environment: &idlcore.ExecutionEnv_Spec{
					Spec: func() *structpb.Struct {
						envSpec, _ := structpb.NewStruct(map[string]interface{}{
							"replica_count":     1,
							"pod_template_spec": providedBytes,
						})
						return envSpec
					}(),
				},
			},
		},
		{
			name: "unmarshal environment spec failure",
			executionEnv: &idlcore.ExecutionEnv{
				Environment: &idlcore.ExecutionEnv_Spec{
					Spec: func() *structpb.Struct {
						invalidSpec, _ := structpb.NewStruct(map[string]interface{}{
							"replica_count": "invalid", // Should be an integer
						})
						return invalidSpec
					}(),
				},
			},
			expectError:   true,
			errorContains: "failed to unmarshal environment spec",
		},
		{
			name: "unmarshal pod template spec failure",
			executionEnv: &idlcore.ExecutionEnv{
				Environment: &idlcore.ExecutionEnv_Spec{
					Spec: func() *structpb.Struct {
						// For this test, we need to create a valid struct but with invalid JSON in the pod_template_spec field
						// We'll encode an invalid JSON string as base64 to make it pass the initial struct unmarshal
						encodedInvalidJson := base64.StdEncoding.EncodeToString([]byte("invalid-json"))
						envSpec, _ := structpb.NewStruct(map[string]interface{}{
							"replica_count":     1,
							"pod_template_spec": encodedInvalidJson,
						})
						return envSpec
					}(),
				},
			},
			expectError:   true,
			errorContains: "failed to unmarshal pod template spec",
		},
		{
			name: "executionEnv_Extant returns error",
			executionEnv: &idlcore.ExecutionEnv{
				Environment: &idlcore.ExecutionEnv_Extant{},
			},
			expectError:   true,
			errorContains: "executing fasttask from extant is not implemented",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tCtx := &pluginsCoreMock.TaskExecutionContext{}
			metadataMock := &pluginsCoreMock.TaskExecutionMetadata{}
			metadataMock.On("GetAnnotations").Return(map[string]string{})
			metadataMock.On("GetLabels").Return(map[string]string{})
			metadataMock.On("GetNamespace").Return(namespace)
			tCtx.OnTaskExecutionMetadata().Return(metadataMock)

			taskTemplate := &idlcore.TaskTemplate{Metadata: &idlcore.TaskMetadata{}}
			taskReaderMock := &pluginsCoreMock.TaskReader{}
			taskReaderMock.On("Read", mock.Anything).Return(taskTemplate, nil)
			tCtx.OnTaskReader().Return(taskReaderMock)

			// Call the function
			result, err := getFastTaskEnvironmentSpec(ctx, tCtx, test.executionEnv)

			// Validate results
			if test.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.errorContains)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)

				// Validate the successful case
				if test.name == "provided PodTemplateSpec" {
					var podTemplate v1.PodTemplateSpec
					err := json.Unmarshal(result.PodTemplateSpec, &podTemplate)
					require.NoError(t, err)

					assert.Equal(t, namespace, podTemplate.GetNamespace())
					assert.Len(t, podTemplate.Spec.Containers, 1)
					assert.Equal(t, containerName, podTemplate.Spec.Containers[0].Name)
					assert.Equal(t, int32(1), result.GetReplicaCount())
				}
			}
		})
	}
}

func TestGetOrCreateEnvironment(t *testing.T) {
	ctx := context.Background()

	// Set up common test data
	executionEnvID := interfaces.ExecutionEnvID{
		Project: "project",
		Domain:  "domain",
		Name:    "foo",
		Version: "0",
	}

	// Create pod template spec for testing
	podTemplateSpec := &v1.PodTemplateSpec{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Name: "primary"},
			},
		},
	}
	podTemplateSpecBytes, err := json.Marshal(podTemplateSpec)
	require.NoError(t, err)

	// Execution environment with specification
	executionEnv := &idlcore.ExecutionEnv{
		Environment: &idlcore.ExecutionEnv_Spec{
			Spec: func() *structpb.Struct {
				envSpec, _ := structpb.NewStruct(map[string]interface{}{
					"replica_count":          1,
					"primary_container_name": "primary",
					"parallelism":            1,
					"pod_template_spec":      podTemplateSpecBytes,
				})
				return envSpec
			}(),
		},
	}

	executionEnvMultipleReplicas := &idlcore.ExecutionEnv{
		Environment: &idlcore.ExecutionEnv_Spec{
			Spec: func() *structpb.Struct {
				envSpec, _ := structpb.NewStruct(map[string]interface{}{
					"replica_count":          2,
					"primary_container_name": "primary",
					"parallelism":            1,
					"pod_template_spec":      podTemplateSpecBytes,
				})
				return envSpec
			}(),
		},
	}

	// Test cases
	tests := []struct {
		name                 string
		executionEnv         *idlcore.ExecutionEnv
		existingEnvironments map[string]interfaces.Environment
		mockCreatePodError   error
		expectCreateCalls    int
		expectScaleUp        bool
		expectedState        interfaces.State
		expectError          bool
		expectFailureMessage bool
	}{
		{
			name:                 "Success - Create New Environment",
			executionEnv:         executionEnv,
			existingEnvironments: map[string]interfaces.Environment{},
			expectCreateCalls:    1,
			expectedState:        interfaces.HEALTHY,
		},
		{
			name:                 "Success - Create New Environment - multiple replicas",
			executionEnv:         executionEnvMultipleReplicas,
			existingEnvironments: map[string]interfaces.Environment{},
			expectCreateCalls:    1,
			expectScaleUp:        true,
			expectedState:        interfaces.HEALTHY,
		},
		{
			name: "Success - Return Existing Healthy Environment",
			existingEnvironments: map[string]interfaces.Environment{
				executionEnvID.String(): &environmentImpl{
					state: interfaces.HEALTHY,
				},
			},
			expectCreateCalls: 0, // No new pods should be created
			expectedState:     interfaces.HEALTHY,
		},
		{
			name: "Success - Return Existing Tombstoned Environment",
			existingEnvironments: map[string]interfaces.Environment{
				executionEnvID.String(): &environmentImpl{
					state: interfaces.TOMBSTONED,
				},
			},
			expectCreateCalls: 0, // No new pods should be created
			expectedState:     interfaces.TOMBSTONED,
		},
		{
			name: "Success - Recover Orphaned Environment",
			existingEnvironments: map[string]interfaces.Environment{
				executionEnvID.String(): &environmentImpl{
					state: interfaces.ORPHANED,
				},
			},
			expectCreateCalls: 0, // We test the recovery process, not pod creation
			expectedState:     interfaces.HEALTHY,
		},
		{
			name:                 "Failure - Pod Creation Errors",
			executionEnv:         executionEnvMultipleReplicas,
			existingEnvironments: map[string]interfaces.Environment{},
			mockCreatePodError:   fmt.Errorf("pod creation failed"),
			expectCreateCalls:    2,
			expectedState:        interfaces.TOMBSTONED,
			expectFailureMessage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create TaskExecutionContext with mocks
			tCtx := &pluginsCoreMock.TaskExecutionContext{}
			metadataMock := &pluginsCoreMock.TaskExecutionMetadata{}
			metadataMock.On("GetAnnotations").Return(map[string]string{})
			metadataMock.On("GetLabels").Return(map[string]string{})
			metadataMock.On("GetNamespace").Return("testNamespace")
			tCtx.OnTaskExecutionMetadata().Return(metadataMock)

			taskTemplate := &idlcore.TaskTemplate{Metadata: &idlcore.TaskMetadata{}}
			taskReaderMock := &pluginsCoreMock.TaskReader{}
			taskReaderMock.On("Read", mock.Anything).Return(taskTemplate, nil)
			tCtx.OnTaskReader().Return(taskReaderMock)

			store := newEnvironmentStore()
			for id, env := range tt.existingEnvironments {
				store.GetOrCreate(id, env)
			}

			// Set up metrics
			scope := promutils.NewTestScope()
			metrics := newBuilderMetrics(scope)

			// Create a mock kubeClient to track pod creation calls
			kubeClient := &kubeClient{}
			kubeClient.returnErrorOnCreate = tt.mockCreatePodError

			kubeCache := &kubeCache{}

			kubeClientImpl := &pluginsCoreMock.KubeClient{}
			kubeClientImpl.OnGetClient().Return(kubeClient)
			kubeClientImpl.OnGetCache().Return(kubeCache)

			// Create test builder
			builder := &environmentBuilderImpl{
				kubeClient:  kubeClientImpl,
				store:       store,
				metrics:     metrics,
				randSource:  rand.New(rand.NewSource(time.Now().UnixNano())),
				scaleUpChan: make(chan string, 1),
			}

			// Call GetOrCreateEnvironment
			env, err := builder.GetOrCreateEnvironment(ctx, tCtx, executionEnvID, tt.executionEnv)

			// Validate results
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, env)

				// Check environment state
				assert.Equal(t, int(tt.expectedState), int(env.State()))

				// Check pod creation calls
				assert.Equal(t, tt.expectCreateCalls, kubeClient.createCalls)

				if tt.expectScaleUp {
					assert.Equal(t, 1, len(builder.scaleUpChan))
				}

				if tt.expectFailureMessage {
					assert.NotEmpty(t, env.FailureMessage())
				} else {
					assert.Empty(t, env.FailureMessage())
				}
			}
		})
	}
}

func TestDetectOrphanedEnvironments(t *testing.T) {

	// Set up common test data
	executionEnvID := interfaces.ExecutionEnvID{
		Project: "project",
		Domain:  "domain",
		Name:    "foo",
		Version: "0",
		Org:     "org",
	}

	pods := []v1.Pod{
		v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "bar",
				Labels: map[string]string{
					EXECUTION_ENV_NAME:    executionEnvID.Name,
					EXECUTION_ENV_VERSION: executionEnvID.Version,
					PROJECT_LABEL:         executionEnvID.Project,
					DOMAIN_LABEL:          executionEnvID.Domain,
					ORGANIZATION_LABEL:    executionEnvID.Org,
				},
				Annotations: map[string]string{
					TTL_SECONDS: "60",
				},
			},
		},
		v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "baz",
				Labels: map[string]string{
					EXECUTION_ENV_NAME:    executionEnvID.Name,
					EXECUTION_ENV_VERSION: executionEnvID.Version,
					PROJECT_LABEL:         executionEnvID.Project,
					DOMAIN_LABEL:          executionEnvID.Domain,
					ORGANIZATION_LABEL:    executionEnvID.Org,
				},
				Annotations: map[string]string{
					TTL_SECONDS: "60",
				},
			},
		},
	}

	workersMap := sync.Map{}
	workersMap.Store("bar", &workerImpl{})
	workersMap.Store("baz", &workerImpl{})

	workersMap1 := sync.Map{}
	workersMap1.Store("bar", &workerImpl{})

	ctx := context.TODO()
	tests := []struct {
		name                     string
		environments             map[string]interfaces.Environment
		expectedEnvironmentCount int
		expectedReplicaCount     int
	}{
		{
			name: "Noop",
			environments: map[string]interfaces.Environment{
				executionEnvID.String(): &environmentImpl{
					workers: &workersMap,
					state:   interfaces.HEALTHY,
				},
			},
			expectedEnvironmentCount: 1,
			expectedReplicaCount:     2,
		},
		{
			name:                     "CreateOrphanedEnvironment",
			environments:             map[string]interfaces.Environment{},
			expectedEnvironmentCount: 1,
			expectedReplicaCount:     2,
		},
		{
			name: "ExistingOrphanedEnvironment",
			environments: map[string]interfaces.Environment{
				executionEnvID.String(): &environmentImpl{
					workers: &workersMap1,
					state:   interfaces.ORPHANED,
				},
			},
			expectedEnvironmentCount: 1,
			expectedReplicaCount:     2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// initialize InMemoryBuilder
			kubeClient := &kubeClient{}
			kubeCache := &kubeCache{
				pods: pods,
			}

			kubeClientImpl := &pluginsCoreMock.KubeClient{}
			kubeClientImpl.OnGetClient().Return(kubeClient)
			kubeClientImpl.OnGetCache().Return(kubeCache)

			store := newEnvironmentStore()
			for id, env := range test.environments {
				store.GetOrCreate(id, env)
			}

			// Set up metrics
			scope := promutils.NewTestScope()
			metrics := newBuilderMetrics(scope)

			// Create test builder
			builder := &environmentBuilderImpl{
				kubeClient:  kubeClientImpl,
				store:       store,
				metrics:     metrics,
				randSource:  rand.New(rand.NewSource(time.Now().UnixNano())),
				scaleUpChan: make(chan string, 1),
			}

			// call `Create`
			err := builder.detectOrphanedEnvironments(ctx, kubeCache)
			assert.Nil(t, err)
			assert.Equal(t, test.expectedEnvironmentCount, len(builder.store.List()))
			totalReplicas := 0
			for _, environment := range builder.store.List() {
				environment.RangeWorkers(func(workerID string, worker interfaces.Worker) bool {
					totalReplicas++
					return true
				})
			}
			assert.Equal(t, test.expectedReplicaCount, totalReplicas)
		})
	}
}

// withConfigOverride saves the current config, runs the provided function, and restores the config after
func withConfigOverride(t *testing.T, fn func()) {
	if fn == nil {
		return
	}
	original := *GetConfig()
	t.Cleanup(func() {
		*GetConfig() = original
	})
	fn()
}

func TestAddObjectMetadata(t *testing.T) {
	ctx := context.Background()

	taskMetadata := &coremocks.TaskExecutionMetadata{}
	taskMetadata.OnGetNamespace().Return("test-namespace")
	taskMetadata.OnGetAnnotations().Return(map[string]string{
		"metadataAnnotation": "metadataAnnotation",
	})
	taskMetadata.OnGetLabels().Return(map[string]string{
		"metadataLabel": "metadataLabel",
	})

	taskReader := &coremocks.TaskReader{}
	taskReader.OnRead(ctx).Return(&idlcore.TaskTemplate{
		SecurityContext: &idlcore.SecurityContext{
			Secrets: []*idlcore.Secret{
				{
					Group:            "my_group",
					Key:              "my_key",
					MountRequirement: idlcore.Secret_ENV_VAR,
				},
			},
		},
	}, nil)

	tCtx := &coremocks.TaskExecutionContext{}
	tCtx.OnTaskExecutionMetadata().Return(taskMetadata)
	tCtx.OnTaskReader().Return(taskReader)

	cfg := &config.K8sPluginConfig{
		DefaultAnnotations: map[string]string{
			"defaultAnnotation": "defaultAnnotation",
		},
		DefaultLabels: map[string]string{
			"defaultLabel": "defaultLabel",
		},
	}

	spec := &v1.PodTemplateSpec{}
	err := addObjectMetadata(ctx, tCtx, spec, cfg)

	assert.Nil(t, err)
	assert.Equal(t, map[string]string{
		"defaultAnnotation":  "defaultAnnotation",
		"metadataAnnotation": "metadataAnnotation",
		"flyte.secrets/s0":   "m4zg54lqhiqce2lzl4txe22voarau12fpe4caitnpfpwwzlzeifg122vnz1f53tfof1ws3tfnvsw34b1ebcu3vs6kzavecq",
	}, spec.GetAnnotations())
	assert.Equal(t, map[string]string{
		secrets.PodLabel: secrets.PodLabelValue,
		"defaultLabel":   "defaultLabel",
		"metadataLabel":  "metadataLabel",
	}, spec.GetLabels())
	assert.Equal(t, "test-namespace", spec.GetNamespace())
	assert.Len(t, spec.GetOwnerReferences(), 0)
	assert.Len(t, spec.GetFinalizers(), 0)
}

func TestScaleDown(t *testing.T) {
	ctx := context.Background()

	podTemplateSpec := &v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-namespace",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{Name: "test"}},
		},
	}
	podTemplateSpecBytes, err := json.Marshal(podTemplateSpec)
	assert.NoError(t, err)

	tests := []struct {
		name                   string
		setupConfig            func()
		setupEnv               func() *environmentImpl
		expectedDeletedWorkers []string
		expectedEnvDeleted     bool
		expectedDeletePodCalls int
	}{
		{
			name: "Delete expired environment based on lastAccessedAt",
			setupConfig: func() {
				GetConfig().OrphanedWorkerTTL = flytestdlibConfig.Duration{Duration: 30 * time.Second}
				GetConfig().DefaultWorkerTTL = flytestdlibConfig.Duration{Duration: 60 * time.Second}
				GetConfig().DefaultEnvironmentTTL = flytestdlibConfig.Duration{Duration: 100 * time.Second}
			},
			setupEnv: func() *environmentImpl {
				env := &environmentImpl{
					createdAt: time.Now().Unix() - 200,
					envID: interfaces.ExecutionEnvID{
						Name:    "test-env",
						Project: "test-project",
						Domain:  "test-domain",
					},
					fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{
						PodTemplateSpec: podTemplateSpecBytes,
						TerminationCriteria: &pb.FastTaskEnvironmentSpec_TtlSeconds{
							TtlSeconds: 100,
						},
					},
					lock:    sync.RWMutex{},
					state:   interfaces.HEALTHY,
					workers: &sync.Map{},
				}

				worker := &workerImpl{
					id:             "worker-1",
					lastAccessedAt: time.Now().Unix() - 150,
					state:          interfaces.HEALTHY,
					lock:           sync.RWMutex{},
				}
				env.workers.Store("worker-1", worker)

				return env
			},
			expectedDeletedWorkers: []string{"worker-1"},
			expectedEnvDeleted:     true,
			expectedDeletePodCalls: 1,
		},
		{
			name: "Delete orphaned workers beyond TTL",
			setupConfig: func() {
				GetConfig().OrphanedWorkerTTL = flytestdlibConfig.Duration{Duration: 30 * time.Second}
				GetConfig().DefaultWorkerTTL = flytestdlibConfig.Duration{Duration: 60 * time.Second}
				GetConfig().DefaultEnvironmentTTL = flytestdlibConfig.Duration{Duration: 300 * time.Second}
			},
			setupEnv: func() *environmentImpl {
				env := &environmentImpl{
					createdAt: time.Now().Unix(),
					envID: interfaces.ExecutionEnvID{
						Name:    "test-env",
						Project: "test-project",
						Domain:  "test-domain",
					},
					fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{
						PodTemplateSpec: podTemplateSpecBytes,
						ReplicaCount:    2,
						TerminationCriteria: &pb.FastTaskEnvironmentSpec_TtlSeconds{
							TtlSeconds: 300,
						},
					},
					lock:    sync.RWMutex{},
					state:   interfaces.HEALTHY,
					workers: &sync.Map{},
				}

				// Create orphaned workers
				orphanedWorker1 := &workerImpl{
					id:             "orphaned-1",
					lastAccessedAt: time.Now().Unix() - 40, // Beyond 30s TTL
					state:          interfaces.ORPHANED,
					lock:           sync.RWMutex{},
				}
				orphanedWorker2 := &workerImpl{
					id:             "orphaned-2",
					lastAccessedAt: time.Now().Unix() - 20, // Within 30s TTL
					state:          interfaces.ORPHANED,
					lock:           sync.RWMutex{},
				}
				healthyWorker := &workerImpl{
					id:             "healthy-1",
					lastAccessedAt: time.Now().Unix() - 10,
					state:          interfaces.HEALTHY,
					lock:           sync.RWMutex{},
				}

				env.workers.Store("orphaned-1", orphanedWorker1)
				env.workers.Store("orphaned-2", orphanedWorker2)
				env.workers.Store("healthy-1", healthyWorker)

				return env
			},
			expectedDeletedWorkers: []string{"orphaned-1"},
			expectedEnvDeleted:     false,
			expectedDeletePodCalls: 1,
		},
		{
			name: "Delete expired workers but maintain minimum replicas",
			setupConfig: func() {
				GetConfig().OrphanedWorkerTTL = flytestdlibConfig.Duration{Duration: 30 * time.Second}
				GetConfig().DefaultWorkerTTL = flytestdlibConfig.Duration{Duration: 60 * time.Second}
				GetConfig().DefaultEnvironmentTTL = flytestdlibConfig.Duration{Duration: 300 * time.Second}
			},
			setupEnv: func() *environmentImpl {
				env := &environmentImpl{
					createdAt: time.Now().Unix(),
					envID: interfaces.ExecutionEnvID{
						Name:    "test-env",
						Project: "test-project",
						Domain:  "test-domain",
					},
					fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{
						PodTemplateSpec: podTemplateSpecBytes,
						ReplicaCount:    2, // Minimum 2 replicas
						TerminationCriteria: &pb.FastTaskEnvironmentSpec_TtlSeconds{
							TtlSeconds: 300,
						},
					},
					lock:    sync.RWMutex{},
					state:   interfaces.HEALTHY,
					workers: &sync.Map{},
				}

				// Create 3 healthy workers, 2 expired
				worker1 := &workerImpl{
					id:             "worker-1",
					lastAccessedAt: time.Now().Unix() - 70, // Expired
					state:          interfaces.HEALTHY,
					lock:           sync.RWMutex{},
				}
				worker2 := &workerImpl{
					id:             "worker-2",
					lastAccessedAt: time.Now().Unix() - 65, // Expired
					state:          interfaces.HEALTHY,
					lock:           sync.RWMutex{},
				}
				worker3 := &workerImpl{
					id:             "worker-3",
					lastAccessedAt: time.Now().Unix() - 10, // Not expired
					state:          interfaces.HEALTHY,
					lock:           sync.RWMutex{},
				}

				env.workers.Store("worker-1", worker1)
				env.workers.Store("worker-2", worker2)
				env.workers.Store("worker-3", worker3)

				return env
			},
			expectedDeletedWorkers: []string{"worker-1"},
			expectedEnvDeleted:     false,
			expectedDeletePodCalls: 1,
		},
		{
			name: "Environment with only orphaned workers uses lastAccessedAt for TTL check",
			setupConfig: func() {
				GetConfig().OrphanedWorkerTTL = flytestdlibConfig.Duration{Duration: 30 * time.Second}
				GetConfig().DefaultWorkerTTL = flytestdlibConfig.Duration{Duration: 60 * time.Second}
				GetConfig().DefaultEnvironmentTTL = flytestdlibConfig.Duration{Duration: 100 * time.Second}
			},
			setupEnv: func() *environmentImpl {
				env := &environmentImpl{
					createdAt: time.Now().Unix() - 200, // Created 200s ago
					envID: interfaces.ExecutionEnvID{
						Name:    "test-env",
						Project: "test-project",
						Domain:  "test-domain",
					},
					fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{
						PodTemplateSpec: podTemplateSpecBytes,
						TerminationCriteria: &pb.FastTaskEnvironmentSpec_TtlSeconds{
							TtlSeconds: 100, // 100s TTL
						},
					},
					lock:    sync.RWMutex{},
					state:   interfaces.HEALTHY,
					workers: &sync.Map{},
				}

				// Create only orphaned workers with recent activity
				orphanedWorker := &workerImpl{
					id:             "orphaned-1",
					lastAccessedAt: time.Now().Unix() - 10, // Recently accessed
					state:          interfaces.ORPHANED,
					lock:           sync.RWMutex{},
				}
				env.workers.Store("orphaned-1", orphanedWorker)

				return env
			},
			expectedDeletedWorkers: []string{},
			expectedEnvDeleted:     false, // Should NOT delete because worker was recently accessed
			expectedDeletePodCalls: 0,
		},
		{
			name: "Empty environment uses createdAt for TTL check",
			setupConfig: func() {
				GetConfig().DefaultEnvironmentTTL = flytestdlibConfig.Duration{Duration: 100 * time.Second}
			},
			setupEnv: func() *environmentImpl {
				env := &environmentImpl{
					createdAt: time.Now().Unix() - 150, // Created 150s ago
					envID: interfaces.ExecutionEnvID{
						Name:    "test-env",
						Project: "test-project",
						Domain:  "test-domain",
					},
					fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{
						PodTemplateSpec: podTemplateSpecBytes,
						TerminationCriteria: &pb.FastTaskEnvironmentSpec_TtlSeconds{
							TtlSeconds: 100, // 100s TTL
						},
					},
					lock:    sync.RWMutex{},
					state:   interfaces.HEALTHY,
					workers: &sync.Map{},
				}

				return env
			},
			expectedDeletedWorkers: []string{},
			expectedEnvDeleted:     true, // Should delete because createdAt exceeds TTL
			expectedDeletePodCalls: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withConfigOverride(t, tt.setupConfig)

			env := tt.setupEnv()

			kubeClient := &kubeClient{}
			mockKubeClient := &pluginsCoreMock.KubeClient{}
			mockKubeClient.OnGetClient().Return(kubeClient)

			store := newEnvironmentStore()
			store.GetOrCreate(env.EnvID().String(), env)

			scope := promutils.NewTestScope()
			builder := &environmentBuilderImpl{
				kubeClient: mockKubeClient,
				store:      store,
				metrics:    newBuilderMetrics(scope),
			}

			err := builder.scaleDown(ctx, env)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedDeletePodCalls, kubeClient.deleteCalls)

			if tt.expectedEnvDeleted {
				assert.Nil(t, store.Get(env.EnvID().String()), "Environment should be deleted from store")
			} else {
				assert.NotNil(t, store.Get(env.EnvID().String()), "Environment should remain in store")

				for _, workerID := range tt.expectedDeletedWorkers {
					_, exists := env.workers.Load(workerID)
					assert.False(t, exists, "Worker %s should be deleted", workerID)
				}
			}
		})
	}
}

func TestScaleUp(t *testing.T) {
	ctx := context.Background()

	podTemplateSpec := &v1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-namespace",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{{Name: "test"}},
		},
	}
	podTemplateSpecBytes, _ := json.Marshal(podTemplateSpec)

	tests := []struct {
		name                   string
		setupEnv               func() *environmentImpl
		expectedCreatePodCalls int
		expectedWorkerCount    int
	}{
		{
			name: "Scale up to minimum replicas",
			setupEnv: func() *environmentImpl {
				env := &environmentImpl{
					createdAt: time.Now().Unix(),
					envID: interfaces.ExecutionEnvID{
						Name:    "test-env",
						Project: "test-project",
						Domain:  "test-domain",
					},
					fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{
						PodTemplateSpec: podTemplateSpecBytes,
						ReplicaCount:    3, // Want 3 replicas
					},
					lock:    sync.RWMutex{},
					state:   interfaces.HEALTHY,
					workers: &sync.Map{},
				}

				// Start with 1 worker
				worker := &workerImpl{
					id:    "worker-1",
					state: interfaces.HEALTHY,
					lock:  sync.RWMutex{},
				}
				env.workers.Store("worker-1", worker)

				return env
			},
			expectedCreatePodCalls: 2,
			expectedWorkerCount:    3,
		},
		{
			name: "Already at desired replica count",
			setupEnv: func() *environmentImpl {
				env := &environmentImpl{
					createdAt: time.Now().Unix(),
					envID: interfaces.ExecutionEnvID{
						Name:    "test-env",
						Project: "test-project",
						Domain:  "test-domain",
					},
					fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{
						PodTemplateSpec: podTemplateSpecBytes,
						ReplicaCount:    2,
					},
					lock:    sync.RWMutex{},
					state:   interfaces.HEALTHY,
					workers: &sync.Map{},
				}

				// Already have 2 workers
				worker1 := &workerImpl{
					id:    "worker-1",
					state: interfaces.HEALTHY,
					lock:  sync.RWMutex{},
				}
				worker2 := &workerImpl{
					id:    "worker-2",
					state: interfaces.HEALTHY,
					lock:  sync.RWMutex{},
				}
				env.workers.Store("worker-1", worker1)
				env.workers.Store("worker-2", worker2)

				return env
			},
			expectedCreatePodCalls: 0,
			expectedWorkerCount:    2,
		},
		{
			name: "Skip scale up for unhealthy environment",
			setupEnv: func() *environmentImpl {
				env := &environmentImpl{
					createdAt: time.Now().Unix(),
					envID: interfaces.ExecutionEnvID{
						Name:    "test-env",
						Project: "test-project",
						Domain:  "test-domain",
					},
					fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{
						PodTemplateSpec: podTemplateSpecBytes,
						ReplicaCount:    3,
					},
					lock:    sync.RWMutex{},
					state:   interfaces.TOMBSTONED,
					workers: &sync.Map{},
				}

				return env
			},
			expectedCreatePodCalls: 0,
			expectedWorkerCount:    0,
		},
		{
			name: "Scale up by one when between min and max",
			setupEnv: func() *environmentImpl {
				minReplicas := int32(2)
				env := &environmentImpl{
					createdAt: time.Now().Unix(),
					envID: interfaces.ExecutionEnvID{
						Name:    "test-env",
						Project: "test-project",
						Domain:  "test-domain",
					},
					fastTaskEnvironmentSpec: &pb.FastTaskEnvironmentSpec{
						PodTemplateSpec: podTemplateSpecBytes,
						ReplicaCount:    5,                                        // Max replicas
						MinReplicaCount: &wrappers.Int32Value{Value: minReplicas}, // Min replicas
					},
					lock:    sync.RWMutex{},
					state:   interfaces.HEALTHY,
					workers: &sync.Map{},
				}

				// Have 3 workers (above min of 2, below max of 5)
				for i := 1; i <= 3; i++ {
					worker := &workerImpl{
						id:    fmt.Sprintf("worker-%d", i),
						state: interfaces.HEALTHY,
						lock:  sync.RWMutex{},
					}
					env.workers.Store(worker.id, worker)
				}

				return env
			},
			expectedCreatePodCalls: 1,
			expectedWorkerCount:    4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := tt.setupEnv()

			kubeClient := &kubeClient{}
			mockKubeClient := &pluginsCoreMock.KubeClient{}
			mockKubeClient.OnGetClient().Return(kubeClient)

			store := newEnvironmentStore()
			store.GetOrCreate(env.EnvID().String(), env)

			scope := promutils.NewTestScope()
			builder := &environmentBuilderImpl{
				kubeClient: mockKubeClient,
				store:      store,
				metrics:    newBuilderMetrics(scope),
				randSource: rand.New(rand.NewSource(12345)),
			}

			err := builder.scaleUp(ctx, env)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedCreatePodCalls, kubeClient.createCalls)

			workerCount := 0
			env.workers.Range(func(key, value interface{}) bool {
				workerCount++
				return true
			})
			assert.Equal(t, tt.expectedWorkerCount, workerCount)
		})
	}
}
