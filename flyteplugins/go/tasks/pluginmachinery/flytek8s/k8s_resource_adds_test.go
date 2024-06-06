package flytek8s

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
)

func TestGetExecutionEnvVars(t *testing.T) {
	mock := mockTaskExecutionIdentifier{}
	envVars := GetExecutionEnvVars(mock, "")
	assert.Len(t, envVars, 12)
}

func TestGetTolerationsForResources(t *testing.T) {
	var empty []v12.Toleration
	var emptyConfig map[v12.ResourceName][]v12.Toleration

	tolGPU := v12.Toleration{
		Key:      "flyte/gpu",
		Value:    "dedicated",
		Operator: v12.TolerationOpEqual,
		Effect:   v12.TaintEffectNoSchedule,
	}

	tolEphemeralStorage := v12.Toleration{
		Key:      "ephemeral-storage",
		Value:    "dedicated",
		Operator: v12.TolerationOpExists,
		Effect:   v12.TaintEffectNoSchedule,
	}

	type args struct {
		resources v12.ResourceRequirements
	}
	tests := []struct {
		name        string
		args        args
		setVal      map[v12.ResourceName][]v12.Toleration
		setDefaults []v12.Toleration
		want        []v12.Toleration
	}{
		{
			"no-tolerations-limits",
			args{
				v12.ResourceRequirements{
					Limits: v12.ResourceList{
						v12.ResourceCPU:              resource.MustParse("1024m"),
						v12.ResourceEphemeralStorage: resource.MustParse("100M"),
					},
				},
			},
			emptyConfig,
			nil,
			empty,
		},
		{
			"no-tolerations-req",
			args{
				v12.ResourceRequirements{
					Requests: v12.ResourceList{
						v12.ResourceCPU:              resource.MustParse("1024m"),
						v12.ResourceEphemeralStorage: resource.MustParse("100M"),
					},
				},
			},
			emptyConfig,
			nil,
			empty,
		},
		{
			"no-tolerations-both",
			args{
				v12.ResourceRequirements{
					Limits: v12.ResourceList{
						v12.ResourceCPU:              resource.MustParse("1024m"),
						v12.ResourceEphemeralStorage: resource.MustParse("100M"),
					},
					Requests: v12.ResourceList{
						v12.ResourceCPU:              resource.MustParse("1024m"),
						v12.ResourceEphemeralStorage: resource.MustParse("100M"),
					},
				},
			},
			emptyConfig,
			nil,
			empty,
		},
		{
			"tolerations-limits",
			args{
				v12.ResourceRequirements{
					Limits: v12.ResourceList{
						v12.ResourceCPU:              resource.MustParse("1024m"),
						v12.ResourceEphemeralStorage: resource.MustParse("100M"),
					},
				},
			},
			map[v12.ResourceName][]v12.Toleration{
				v12.ResourceEphemeralStorage: {tolEphemeralStorage},
				ResourceNvidiaGPU:            {tolGPU},
			},
			nil,
			[]v12.Toleration{tolEphemeralStorage},
		},
		{
			"tolerations-req",
			args{
				v12.ResourceRequirements{
					Requests: v12.ResourceList{
						v12.ResourceCPU:              resource.MustParse("1024m"),
						v12.ResourceEphemeralStorage: resource.MustParse("100M"),
					},
				},
			},
			map[v12.ResourceName][]v12.Toleration{
				v12.ResourceEphemeralStorage: {tolEphemeralStorage},
				ResourceNvidiaGPU:            {tolGPU},
			},
			nil,
			[]v12.Toleration{tolEphemeralStorage},
		},
		{
			"tolerations-both",
			args{
				v12.ResourceRequirements{
					Limits: v12.ResourceList{
						v12.ResourceCPU:              resource.MustParse("1024m"),
						v12.ResourceEphemeralStorage: resource.MustParse("100M"),
					},
					Requests: v12.ResourceList{
						v12.ResourceCPU:              resource.MustParse("1024m"),
						v12.ResourceEphemeralStorage: resource.MustParse("100M"),
					},
				},
			},
			map[v12.ResourceName][]v12.Toleration{
				v12.ResourceEphemeralStorage: {tolEphemeralStorage},
				ResourceNvidiaGPU:            {tolGPU},
			},
			nil,
			[]v12.Toleration{tolEphemeralStorage},
		},
		{
			"no-tolerations-both",
			args{
				v12.ResourceRequirements{
					Limits: v12.ResourceList{
						v12.ResourceCPU:              resource.MustParse("1024m"),
						v12.ResourceEphemeralStorage: resource.MustParse("100M"),
						ResourceNvidiaGPU:            resource.MustParse("1"),
					},
					Requests: v12.ResourceList{
						v12.ResourceCPU:              resource.MustParse("1024m"),
						v12.ResourceEphemeralStorage: resource.MustParse("100M"),
					},
				},
			},
			map[v12.ResourceName][]v12.Toleration{
				v12.ResourceEphemeralStorage: {tolEphemeralStorage},
				ResourceNvidiaGPU:            {tolGPU},
			},
			nil,
			[]v12.Toleration{tolEphemeralStorage, tolGPU},
		},
		{
			"default-tolerations",
			args{},
			nil,
			[]v12.Toleration{tolEphemeralStorage},
			[]v12.Toleration{tolEphemeralStorage},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{ResourceTolerations: tt.setVal, DefaultTolerations: tt.setDefaults}))
			if got := GetPodTolerations(true, tt.args.resources); len(got) != len(tt.want) {
				t.Errorf("GetPodTolerations() = %v, want %v", got, tt.want)
			} else {
				for _, tol := range tt.want {
					assert.Contains(t, got, tol)
				}
			}
		})
	}
}

var testTaskExecutionIdentifier = core.TaskExecutionIdentifier{
	TaskId: &core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "proj",
		Domain:       "domain",
		Name:         "name",
	},
	RetryAttempt: 1,
	NodeExecutionId: &core.NodeExecutionIdentifier{
		NodeId: "nodeId",
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "proj",
			Domain:  "domain",
			Name:    "name",
		},
	},
}

type mockTaskExecutionIdentifier struct{}

func (m mockTaskExecutionIdentifier) GetID() core.TaskExecutionIdentifier {
	return testTaskExecutionIdentifier
}

func (m mockTaskExecutionIdentifier) GetGeneratedNameWith(minLength, maxLength int) (string, error) {
	return "task-exec-name", nil
}

func (m mockTaskExecutionIdentifier) GetGeneratedName() string {
	return "task-exec-name"
}

func (m mockTaskExecutionIdentifier) GetUniqueNodeID() string {
	return "unique-node-id"
}

func TestDecorateEnvVars(t *testing.T) {
	ctx := context.Background()
	ctx = contextutils.WithWorkflowID(ctx, "fake_workflow")

	defaultEnv := []v12.EnvVar{
		{
			Name:  "x",
			Value: "y",
		},
	}
	additionalEnv := map[string]string{
		"k": "v",
	}
	var emptyEnvVar map[string]string
	envVarsFromEnv := map[string]string{
		"k": "value",
	}

	originalEnvVal := os.Getenv("value")
	err := os.Setenv("value", "v")
	if err != nil {
		t.Fatalf("failed to set env var 'value'; %v", err)
	}
	defer os.Setenv("value", originalEnvVal)

	expected := append(defaultEnv, GetContextEnvVars(ctx)...)
	expected = append(expected, GetExecutionEnvVars(mockTaskExecutionIdentifier{}, "")...)

	aggregated := append(expected, v12.EnvVar{Name: "k", Value: "v"})
	type args struct {
		envVars []v12.EnvVar
		id      pluginsCore.TaskExecutionID
	}
	tests := []struct {
		name                  string
		args                  args
		additionEnvVar        map[string]string
		additionEnvVarFromEnv map[string]string
		executionEnvVar       map[string]string
		consoleURL            string
		want                  []v12.EnvVar
	}{
		{"no-additional", args{envVars: defaultEnv, id: mockTaskExecutionIdentifier{}}, emptyEnvVar, emptyEnvVar, emptyEnvVar, "", expected},
		{"with-additional", args{envVars: defaultEnv, id: mockTaskExecutionIdentifier{}}, additionalEnv, emptyEnvVar, emptyEnvVar, "", aggregated},
		{"from-env", args{envVars: defaultEnv, id: mockTaskExecutionIdentifier{}}, emptyEnvVar, envVarsFromEnv, emptyEnvVar, "", aggregated},
		{"from-execution-metadata", args{envVars: defaultEnv, id: mockTaskExecutionIdentifier{}}, emptyEnvVar, emptyEnvVar, additionalEnv, "", aggregated},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
				DefaultEnvVars:        tt.additionEnvVar,
				DefaultEnvVarsFromEnv: tt.additionEnvVarFromEnv,
			}))
			if got, _ := DecorateEnvVars(ctx, tt.args.envVars, tt.executionEnvVar, tt.args.id, tt.consoleURL); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecorateEnvVars() = %v, want %v", got, tt.want)
			}
		})
	}
}
