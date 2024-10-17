package plugin

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	coremocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"

	"github.com/unionai/flyte/fasttask/plugin/pb"
)

type kubeClient struct {
	client.Client

	createCalls int
	deleteCalls int
}

func (k *kubeClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	k.createCalls++
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

func TestCreate(t *testing.T) {
	executionEnvID := core.ExecutionEnvID{
		Project: "project",
		Domain:  "domain",
		Name:    "foo",
		Version: "0",
	}

	fastTaskEnvironment := &pb.FastTaskEnvironment{
		QueueId: executionEnvID.String(),
	}

	fastTaskEnvStruct := &_struct.Struct{}
	err := utils.MarshalStruct(fastTaskEnvironment, fastTaskEnvStruct)
	assert.Nil(t, err)

	podTemplateSpec := &v1.PodTemplateSpec{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{
					Name: "primary",
				},
			},
		},
	}
	podTemplateSpecBytes, err := json.Marshal(podTemplateSpec)
	assert.Nil(t, err)

	fastTaskEnvSpec := &pb.FastTaskEnvironmentSpec{
		Parallelism:          1,
		PrimaryContainerName: "primary",
		PodTemplateSpec:      podTemplateSpecBytes,
		ReplicaCount:         2,
		TerminationCriteria: &pb.FastTaskEnvironmentSpec_TtlSeconds{
			TtlSeconds: 300,
		},
	}

	ctx := context.TODO()
	tests := []struct {
		name                string
		environmentSpec     *pb.FastTaskEnvironmentSpec
		environments        map[string]*environment
		expectedEnvironment *_struct.Struct
		expectedCreateCalls int
	}{
		{
			name:                "Success",
			environmentSpec:     fastTaskEnvSpec,
			environments:        map[string]*environment{},
			expectedEnvironment: fastTaskEnvStruct,
			expectedCreateCalls: 2,
		},
		{
			name:            "Exists",
			environmentSpec: fastTaskEnvSpec,
			environments: map[string]*environment{
				executionEnvID.String(): &environment{
					extant: fastTaskEnvStruct,
					state:  HEALTHY,
				},
			},
			expectedEnvironment: fastTaskEnvStruct,
			expectedCreateCalls: 0,
		},
		{
			name:            "Orphaned",
			environmentSpec: fastTaskEnvSpec,
			environments: map[string]*environment{
				executionEnvID.String(): &environment{
					extant:   fastTaskEnvStruct,
					replicas: []string{"bar"},
					state:    ORPHANED,
				},
			},
			expectedEnvironment: fastTaskEnvStruct,
			expectedCreateCalls: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scope := promutils.NewTestScope()

			fastTaskEnvSpecStruct := &_struct.Struct{}
			err := utils.MarshalStruct(test.environmentSpec, fastTaskEnvSpecStruct)
			assert.Nil(t, err)

			// initialize InMemoryBuilder
			kubeClient := &kubeClient{}
			kubeCache := &kubeCache{}

			kubeClientImpl := &coremocks.KubeClient{}
			kubeClientImpl.OnGetClient().Return(kubeClient)
			kubeClientImpl.OnGetCache().Return(kubeCache)

			builder := NewEnvironmentBuilder(kubeClientImpl, scope)
			builder.environments = test.environments

			// call `Create`
			environment, err := builder.Create(ctx, executionEnvID, fastTaskEnvSpecStruct)
			assert.Nil(t, err)
			assert.True(t, proto.Equal(test.expectedEnvironment, environment))
			assert.Equal(t, test.expectedCreateCalls, kubeClient.createCalls)
		})
	}
}

func TestDetectOrphanedEnvironments(t *testing.T) {
	pods := []v1.Pod{
		v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "bar",
				Labels: map[string]string{
					EXECUTION_ENV_NAME:    "foo",
					EXECUTION_ENV_VERSION: "0",
					PROJECT_LABEL:         "project",
					DOMAIN_LABEL:          "domain",
					ORGANIZATION_LABEL:    "",
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
					EXECUTION_ENV_NAME:    "foo",
					EXECUTION_ENV_VERSION: "0",
					PROJECT_LABEL:         "project",
					DOMAIN_LABEL:          "domain",
					ORGANIZATION_LABEL:    "",
				},
				Annotations: map[string]string{
					TTL_SECONDS: "60",
				},
			},
		},
	}

	ctx := context.TODO()
	tests := []struct {
		name                     string
		environments             map[string]*environment
		expectedEnvironmentCount int
		expectedReplicaCount     int
	}{
		{
			name: "Noop",
			environments: map[string]*environment{
				"project_domain_foo_0": &environment{
					replicas: []string{"bar", "baz"},
					state:    HEALTHY,
				},
			},
			expectedEnvironmentCount: 1,
			expectedReplicaCount:     2,
		},
		{
			name:                     "CreateOrphanedEnvironment",
			environments:             map[string]*environment{},
			expectedEnvironmentCount: 1,
			expectedReplicaCount:     2,
		},
		{
			name: "ExistingOrphanedEnvironment",
			environments: map[string]*environment{
				"project_domain_foo_0": &environment{
					replicas: []string{"bar"},
					state:    ORPHANED,
				},
			},
			expectedEnvironmentCount: 1,
			expectedReplicaCount:     2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scope := promutils.NewTestScope()

			// initialize InMemoryBuilder
			kubeClient := &kubeClient{}
			kubeCache := &kubeCache{
				pods: pods,
			}

			kubeClientImpl := &coremocks.KubeClient{}
			kubeClientImpl.OnGetClient().Return(kubeClient)
			kubeClientImpl.OnGetCache().Return(kubeCache)

			builder := NewEnvironmentBuilder(kubeClientImpl, scope)
			builder.environments = test.environments

			// call `Create`
			err := builder.detectOrphanedEnvironments(ctx, kubeCache)
			assert.Nil(t, err)
			assert.Equal(t, test.expectedEnvironmentCount, len(builder.environments))
			totalReplicas := 0
			for _, environment := range builder.environments {
				totalReplicas += len(environment.replicas)
			}
			assert.Equal(t, test.expectedReplicaCount, totalReplicas)
		})
	}
}

func TestGCEnvironments(t *testing.T) {
	podTemplateSpec := &v1.PodTemplateSpec{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{
					Name: "primary",
				},
			},
		},
	}
	podTemplateSpecBytes, err := json.Marshal(podTemplateSpec)
	assert.Nil(t, err)

	fastTaskEnvSpec := &pb.FastTaskEnvironmentSpec{
		PodTemplateSpec: podTemplateSpecBytes,
		TerminationCriteria: &pb.FastTaskEnvironmentSpec_TtlSeconds{
			TtlSeconds: 300,
		},
	}

	ctx := context.TODO()
	tests := []struct {
		name                     string
		environments             map[string]*environment
		expectedDeleteCalls      int
		expectedEnvironmentCount int
	}{
		{
			name: "Noop",
			environments: map[string]*environment{
				"foo": &environment{
					lastAccessedAt: time.Now(),
					replicas:       []string{"bar", "baz"},
					spec:           fastTaskEnvSpec,
					state:          HEALTHY,
				},
			},
			expectedDeleteCalls:      0,
			expectedEnvironmentCount: 1,
		},
		{
			name: "TimeoutTtl",
			environments: map[string]*environment{
				"foo": &environment{
					lastAccessedAt: time.Now().Add(-301 * time.Second),
					replicas:       []string{"bar", "baz"},
					spec:           fastTaskEnvSpec,
					state:          TOMBSTONED,
				},
			},
			expectedDeleteCalls:      2,
			expectedEnvironmentCount: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scope := promutils.NewTestScope()

			// initialize InMemoryBuilder
			kubeClient := &kubeClient{}
			kubeCache := &kubeCache{}

			kubeClientImpl := &coremocks.KubeClient{}
			kubeClientImpl.OnGetClient().Return(kubeClient)
			kubeClientImpl.OnGetCache().Return(kubeCache)

			builder := NewEnvironmentBuilder(kubeClientImpl, scope)
			builder.environments = test.environments

			// call `Create`
			err := builder.gcEnvironments(ctx)
			assert.Nil(t, err)
			assert.Equal(t, test.expectedDeleteCalls, kubeClient.deleteCalls)
			assert.Equal(t, test.expectedEnvironmentCount, len(builder.environments))
		})
	}
}

func TestGet(t *testing.T) {
	executionEnvID := core.ExecutionEnvID{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
		Version: "version",
	}

	fastTaskEnvironment := &pb.FastTaskEnvironment{
		QueueId: executionEnvID.String(),
	}

	fastTaskEnvStruct := &_struct.Struct{}
	err := utils.MarshalStruct(fastTaskEnvironment, fastTaskEnvStruct)
	assert.Nil(t, err)

	ctx := context.TODO()
	tests := []struct {
		name                string
		environments        map[string]*environment
		expectedEnvironment *_struct.Struct
	}{
		{
			name: "Exists",
			environments: map[string]*environment{
				executionEnvID.String(): &environment{
					extant: fastTaskEnvStruct,
					state:  HEALTHY,
				},
			},
			expectedEnvironment: fastTaskEnvStruct,
		},
		{
			name:                "DoesNotExist",
			environments:        map[string]*environment{},
			expectedEnvironment: nil,
		},
		{
			name: "Tombstoned",
			environments: map[string]*environment{
				executionEnvID.String(): &environment{
					state: TOMBSTONED,
				},
			},
			expectedEnvironment: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scope := promutils.NewTestScope()

			// initialize InMemoryBuilder
			kubeClient := &kubeClient{}
			kubeCache := &kubeCache{}

			kubeClientImpl := &coremocks.KubeClient{}
			kubeClientImpl.OnGetClient().Return(kubeClient)
			kubeClientImpl.OnGetCache().Return(kubeCache)

			builder := NewEnvironmentBuilder(kubeClientImpl, scope)
			builder.environments = test.environments

			// call `Get`
			environment := builder.Get(ctx, executionEnvID)
			assert.True(t, proto.Equal(test.expectedEnvironment, environment))
		})
	}
}

func TestRepairEnvironments(t *testing.T) {
	fastTaskEnvironment := &pb.FastTaskEnvironment{
		QueueId: "foo",
	}

	fastTaskEnvStruct := &_struct.Struct{}
	err := utils.MarshalStruct(fastTaskEnvironment, fastTaskEnvStruct)
	assert.Nil(t, err)

	podTemplateSpec := &v1.PodTemplateSpec{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				v1.Container{
					Name: "primary",
				},
			},
		},
	}
	podTemplateSpecBytes, err := json.Marshal(podTemplateSpec)
	assert.Nil(t, err)

	fastTaskEnvSpec := &pb.FastTaskEnvironmentSpec{
		Parallelism:          1,
		PrimaryContainerName: "primary",
		PodTemplateSpec:      podTemplateSpecBytes,
		ReplicaCount:         2,
		TerminationCriteria: &pb.FastTaskEnvironmentSpec_TtlSeconds{
			TtlSeconds: 300,
		},
	}

	ctx := context.TODO()
	tests := []struct {
		name                string
		environments        map[string]*environment
		existingPods        []v1.Pod
		expectedCreateCalls int
	}{
		{
			name: "Noop",
			environments: map[string]*environment{
				"foo": &environment{
					extant:   fastTaskEnvStruct,
					replicas: []string{"bar", "baz"},
					spec:     fastTaskEnvSpec,
					state:    HEALTHY,
				},
			},
			existingPods: []v1.Pod{
				v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "bar",
					},
				},
				v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "baz",
					},
				},
			},
			expectedCreateCalls: 0,
		},
		{
			name: "RepairMissingPod",
			environments: map[string]*environment{
				"foo": &environment{
					extant:   fastTaskEnvStruct,
					replicas: []string{"bar", "baz"},
					spec:     fastTaskEnvSpec,
					state:    REPAIRING,
				},
			},
			existingPods: []v1.Pod{
				v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "bar",
					},
				},
			},
			expectedCreateCalls: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scope := promutils.NewTestScope()

			// initialize InMemoryBuilder
			kubeClient := &kubeClient{}
			kubeCache := &kubeCache{
				pods: test.existingPods,
			}

			kubeClientImpl := &coremocks.KubeClient{}
			kubeClientImpl.OnGetClient().Return(kubeClient)
			kubeClientImpl.OnGetCache().Return(kubeCache)

			builder := NewEnvironmentBuilder(kubeClientImpl, scope)
			builder.environments = test.environments

			// call `Create`
			err := builder.repairEnvironments(ctx)
			assert.Nil(t, err)
			assert.Equal(t, test.expectedCreateCalls, kubeClient.createCalls)

			// verify all environments are now healthy
			for _, environment := range builder.environments {
				assert.Equal(t, HEALTHY, environment.state)
			}
		})
	}
}
