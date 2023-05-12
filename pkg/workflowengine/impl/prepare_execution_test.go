package impl

import (
	"testing"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

const testRole = "role"
const testK8sServiceAccount = "sa"

const testRoleSc = "roleSc"
const testK8sServiceAccountSc = "saSc"

var roleNameKey = "iam.amazonaws.com/role"

func TestAddMapValues(t *testing.T) {
	overrides := map[string]string{
		"1": "a",
		"2": "b",
	}
	defaultValues := map[string]string{
		"1": "c",
		"3": "d",
	}
	t.Run("defaultValues nil", func(t *testing.T) {
		vals := addMapValues(overrides, nil)
		assert.EqualValues(t, vals, overrides)
	})
	t.Run("overrides nil", func(t *testing.T) {
		vals := addMapValues(nil, defaultValues)
		assert.EqualValues(t, vals, defaultValues)
	})
	t.Run("overlapping keys correctly set", func(t *testing.T) {
		vals := addMapValues(overrides, defaultValues)
		assert.EqualValues(t, map[string]string{
			"1": "a",
			"2": "b",
			"3": "d",
		}, vals)
	})
}

func TestAddPermissions(t *testing.T) {
	securityCtx := &core.SecurityContext{
		RunAs: &core.Identity{
			IamRole:           testRoleSc,
			K8SServiceAccount: testK8sServiceAccountSc,
		},
	}
	securityCtxFromAuth := &core.SecurityContext{
		RunAs: &core.Identity{
			IamRole:           testRole,
			K8SServiceAccount: testK8sServiceAccount,
		},
	}
	t.Run("security ctx from auth", func(t *testing.T) {
		flyteWf := v1alpha1.FlyteWorkflow{}
		addPermissions(securityCtxFromAuth, roleNameKey, &flyteWf)
		assert.EqualValues(t, flyteWf.Annotations, map[string]string{
			roleNameKey: testRole,
		})
		assert.Equal(t, testK8sServiceAccount, flyteWf.ServiceAccountName)
		assert.True(t, proto.Equal(&flyteWf.SecurityContext, securityCtxFromAuth))
	})

	t.Run("override using security ctx", func(t *testing.T) {
		flyteWf := v1alpha1.FlyteWorkflow{}
		addPermissions(securityCtx, roleNameKey, &flyteWf)
		assert.EqualValues(t, flyteWf.Annotations, map[string]string{
			roleNameKey: testRoleSc,
		})
		assert.Equal(t, testK8sServiceAccountSc, flyteWf.ServiceAccountName)
		assert.True(t, proto.Equal(&flyteWf.SecurityContext, securityCtx))
	})
}

func TestAddExecutionOverrides(t *testing.T) {
	t.Run("task plugin overrides", func(t *testing.T) {
		overrides := []*admin.PluginOverride{
			{
				TaskType:              "taskType1",
				PluginId:              []string{"Plugin1", "Plugin2"},
				MissingPluginBehavior: admin.PluginOverride_USE_DEFAULT,
			},
		}
		workflow := &v1alpha1.FlyteWorkflow{}
		addExecutionOverrides(overrides, nil, nil, nil, workflow)
		assert.EqualValues(t, workflow.ExecutionConfig.TaskPluginImpls, map[string]v1alpha1.TaskPluginOverride{
			"taskType1": {
				PluginIDs:             []string{"Plugin1", "Plugin2"},
				MissingPluginBehavior: admin.PluginOverride_USE_DEFAULT,
			},
		})
	})
	t.Run("max parallelism", func(t *testing.T) {
		workflowExecutionConfig := &admin.WorkflowExecutionConfig{
			MaxParallelism: 100,
		}
		workflow := &v1alpha1.FlyteWorkflow{}
		addExecutionOverrides(nil, workflowExecutionConfig, nil, nil, workflow)
		assert.EqualValues(t, workflow.ExecutionConfig.MaxParallelism, uint32(100))
	})
	t.Run("recovery execution", func(t *testing.T) {
		recoveryExecutionID := &core.WorkflowExecutionIdentifier{
			Project: "p",
			Domain:  "d",
			Name:    "n",
		}
		workflow := &v1alpha1.FlyteWorkflow{}
		addExecutionOverrides(nil, nil, recoveryExecutionID, nil, workflow)
		assert.True(t, proto.Equal(recoveryExecutionID, workflow.ExecutionConfig.RecoveryExecution.WorkflowExecutionIdentifier))
	})
	t.Run("task resources", func(t *testing.T) {
		workflow := &v1alpha1.FlyteWorkflow{}
		addExecutionOverrides(nil, nil, nil, &interfaces.TaskResources{
			Defaults: runtimeInterfaces.TaskResourceSet{
				CPU:    resource.MustParse("1"),
				Memory: resource.MustParse("100Gi"),
			},
			Limits: runtimeInterfaces.TaskResourceSet{
				CPU:              resource.MustParse("2"),
				Memory:           resource.MustParse("200Gi"),
				Storage:          resource.MustParse("5Gi"),
				EphemeralStorage: resource.MustParse("1Gi"),
				GPU:              resource.MustParse("1"),
			},
		}, workflow)
		assert.EqualValues(t, v1alpha1.TaskResourceSpec{
			CPU:    resource.MustParse("1"),
			Memory: resource.MustParse("100Gi"),
		}, workflow.ExecutionConfig.TaskResources.Requests)

		assert.EqualValues(t, v1alpha1.TaskResourceSpec{
			CPU:              resource.MustParse("2"),
			Memory:           resource.MustParse("200Gi"),
			Storage:          resource.MustParse("5Gi"),
			EphemeralStorage: resource.MustParse("1Gi"),
			GPU:              resource.MustParse("1"),
		}, workflow.ExecutionConfig.TaskResources.Limits)
	})
	t.Run("interruptible", func(t *testing.T) {
		workflowExecutionConfig := &admin.WorkflowExecutionConfig{
			Interruptible: &wrappers.BoolValue{Value: true},
		}
		workflow := &v1alpha1.FlyteWorkflow{}
		addExecutionOverrides(nil, workflowExecutionConfig, nil, nil, workflow)
		assert.NotNil(t, workflow.ExecutionConfig.Interruptible)
		assert.True(t, *workflow.ExecutionConfig.Interruptible)
	})
	t.Run("skip cache", func(t *testing.T) {
		workflowExecutionConfig := &admin.WorkflowExecutionConfig{
			OverwriteCache: true,
		}
		workflow := &v1alpha1.FlyteWorkflow{}
		addExecutionOverrides(nil, workflowExecutionConfig, nil, nil, workflow)
		assert.True(t, workflow.ExecutionConfig.OverwriteCache)
	})
	t.Run("Override environment variables", func(t *testing.T) {
		workflowExecutionConfig := &admin.WorkflowExecutionConfig{
			Envs: &admin.Envs{Values: []*core.KeyValuePair{{Key: "key", Value: "value"}}},
		}
		workflow := &v1alpha1.FlyteWorkflow{}
		addExecutionOverrides(nil, workflowExecutionConfig, nil, nil, workflow)
		assert.Equal(t, workflow.ExecutionConfig.EnvironmentVariables, map[string]string{"key": "value"})
	})
}

func TestPrepareFlyteWorkflow(t *testing.T) {
	var flyteWorkflow v1alpha1.FlyteWorkflow
	execID := core.WorkflowExecutionIdentifier{
		Project: "p",
		Domain:  "d",
		Name:    "n",
	}
	recoveryNodeExecutionID := &core.WorkflowExecutionIdentifier{
		Project: "p",
		Domain:  "d",
		Name:    "original",
	}

	var acceptedAt = time.Now()
	err := PrepareFlyteWorkflow(interfaces.ExecutionData{
		ExecutionID: &execID,
		ExecutionParameters: interfaces.ExecutionParameters{
			AcceptedAt: acceptedAt,
			Labels: map[string]string{
				"customlabel": "labelval",
			},
			Annotations: map[string]string{
				"customannotation": "annotationval",
			},
			TaskPluginOverrides: []*admin.PluginOverride{
				{
					TaskType:              "python",
					PluginId:              []string{"plugin a"},
					MissingPluginBehavior: admin.PluginOverride_USE_DEFAULT,
				},
			},
			ExecutionConfig: &admin.WorkflowExecutionConfig{
				MaxParallelism: 50,
				Interruptible:  &wrappers.BoolValue{Value: true},
				SecurityContext: &core.SecurityContext{
					RunAs: &core.Identity{
						IamRole:           testRoleSc,
						K8SServiceAccount: testK8sServiceAccountSc,
					},
				},
			},
			RecoveryExecution: recoveryNodeExecutionID,
			EventVersion:      1,
			RoleNameKey:       roleNameKey,
			RawOutputDataConfig: &admin.RawOutputDataConfig{
				OutputLocationPrefix: "s3://bucket/key",
			},
		},
	}, &flyteWorkflow)
	assert.NoError(t, err)

	assert.EqualValues(t, flyteWorkflow.ExecutionID, v1alpha1.WorkflowExecutionIdentifier{
		WorkflowExecutionIdentifier: &execID,
	})
	assert.EqualValues(t, map[string]string{
		"customlabel": "labelval",
	}, flyteWorkflow.Labels)
	expectedAnnotations := map[string]string{
		roleNameKey:        testRoleSc,
		"customannotation": "annotationval",
	}
	assert.EqualValues(t, expectedAnnotations, flyteWorkflow.Annotations)

	assert.EqualValues(t, map[string]v1alpha1.TaskPluginOverride{
		"python": {
			PluginIDs:             []string{"plugin a"},
			MissingPluginBehavior: admin.PluginOverride_USE_DEFAULT,
		},
	}, flyteWorkflow.ExecutionConfig.TaskPluginImpls)
	assert.Equal(t, flyteWorkflow.ServiceAccountName, testK8sServiceAccountSc)
	assert.Equal(t, flyteWorkflow.ExecutionConfig.MaxParallelism, uint32(50))
	assert.NotNil(t, flyteWorkflow.ExecutionConfig.Interruptible)
	assert.True(t, *flyteWorkflow.ExecutionConfig.Interruptible)
	assert.True(t, proto.Equal(recoveryNodeExecutionID, flyteWorkflow.ExecutionConfig.RecoveryExecution.WorkflowExecutionIdentifier))
	assert.Equal(t, flyteWorkflow.WorkflowMeta.EventVersion, v1alpha1.EventVersion(1))
	assert.Equal(t, flyteWorkflow.RawOutputDataConfig, v1alpha1.RawOutputDataConfig{
		RawOutputDataConfig: &admin.RawOutputDataConfig{
			OutputLocationPrefix: "s3://bucket/key",
		},
	})
}
