package impl

import (
	"context"
	"errors"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/executioncluster"
	execClusterIfaces "github.com/flyteorg/flyteadmin/pkg/executioncluster/interfaces"
	clusterMock "github.com/flyteorg/flyteadmin/pkg/executioncluster/mocks"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyteadmin/pkg/runtime/mocks"
	"github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/workflowengine/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	flyteclient "github.com/flyteorg/flytepropeller/pkg/client/clientset/versioned"
	v1alpha12 "github.com/flyteorg/flytepropeller/pkg/client/clientset/versioned/typed/flyteworkflow/v1alpha1"
	"github.com/stretchr/testify/assert"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/mock"
	k8_api_err "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var fakeFlyteWF = FakeFlyteWorkflowV1alpha1{}

type createCallback func(*v1alpha1.FlyteWorkflow, v1.CreateOptions) (*v1alpha1.FlyteWorkflow, error)
type deleteCallback func(name string, options *v1.DeleteOptions) error
type FakeFlyteWorkflow struct {
	v1alpha12.FlyteWorkflowInterface
	createCallback createCallback
	deleteCallback deleteCallback
}

func (b *FakeFlyteWorkflow) Create(ctx context.Context, wf *v1alpha1.FlyteWorkflow, opts v1.CreateOptions) (*v1alpha1.FlyteWorkflow, error) {
	if b.createCallback != nil {
		return b.createCallback(wf, opts)
	}
	return nil, nil
}

func (b *FakeFlyteWorkflow) Delete(ctx context.Context, name string, options v1.DeleteOptions) error {
	if b.deleteCallback != nil {
		return b.deleteCallback(name, &options)
	}
	return nil
}

type flyteWorkflowsCallback func(string) v1alpha12.FlyteWorkflowInterface

type FakeFlyteWorkflowV1alpha1 struct {
	v1alpha12.FlyteworkflowV1alpha1Interface
	flyteWorkflowsCallback flyteWorkflowsCallback
}

func (b *FakeFlyteWorkflowV1alpha1) FlyteWorkflows(namespace string) v1alpha12.FlyteWorkflowInterface {
	if b.flyteWorkflowsCallback != nil {
		return b.flyteWorkflowsCallback(namespace)
	}
	return &FakeFlyteWorkflow{}
}

type FakeK8FlyteClient struct {
	flyteclient.Interface
	ID string
}

func (b *FakeK8FlyteClient) FlyteworkflowV1alpha1() v1alpha12.FlyteworkflowV1alpha1Interface {
	return &fakeFlyteWF
}

const namespace = "p-d"

const clusterID = "C1"

var execID = &core.WorkflowExecutionIdentifier{
	Project: "proj",
	Domain:  "domain",
	Name:    "name",
}

var flyteWf = &v1alpha1.FlyteWorkflow{
	ExecutionID: v1alpha1.ExecutionID{
		WorkflowExecutionIdentifier: execID,
	},
}

var testInputs = &core.LiteralMap{
	Literals: map[string]*core.Literal{
		"foo": {
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Primitive{
						Primitive: &core.Primitive{
							Value: &core.Primitive_Integer{
								Integer: 4,
							},
						},
					},
				},
			},
		},
	},
}

func getFakeExecutionCluster() execClusterIfaces.ClusterInterface {
	fakeCluster := clusterMock.MockCluster{}
	fakeCluster.SetGetTargetCallback(func(ctx context.Context, spec *executioncluster.ExecutionTargetSpec) (target *executioncluster.ExecutionTarget, e error) {
		return &executioncluster.ExecutionTarget{
			ID:          clusterID,
			FlyteClient: &FakeK8FlyteClient{},
		}, nil
	})
	return &fakeCluster
}

func TestGetID(t *testing.T) {
	executor := K8sWorkflowExecutor{}
	assert.Equal(t, defaultIdentifier, executor.ID())
}

func TestExecute(t *testing.T) {
	fakeFlyteWorkflow := FakeFlyteWorkflow{}
	fakeFlyteWorkflow.createCallback = func(flyteWorkflow *v1alpha1.FlyteWorkflow, opts v1.CreateOptions) (*v1alpha1.FlyteWorkflow, error) {
		assert.Equal(t, flyteWf, flyteWorkflow)
		assert.Empty(t, opts)
		return nil, nil
	}
	fakeFlyteWF.flyteWorkflowsCallback = func(ns string) v1alpha12.FlyteWorkflowInterface {
		assert.Equal(t, namespace, ns)
		return &fakeFlyteWorkflow
	}

	mockApplicationConfig := runtimeMocks.MockApplicationProvider{}
	mockApplicationConfig.SetTopLevelConfig(runtimeInterfaces.ApplicationConfig{
		UseOffloadedWorkflowClosure: false,
	})
	mockRuntime := runtimeMocks.NewMockConfigurationProvider(&mockApplicationConfig, nil, nil, nil, nil, nil)

	mockBuilder := mocks.FlyteWorkflowBuilder{}
	workflowClosure := core.CompiledWorkflowClosure{
		Primary: &core.CompiledWorkflow{
			Template: &core.WorkflowTemplate{
				Id: &core.Identifier{
					Project: "p",
					Domain:  "d",
					Name:    "n",
					Version: "version",
				},
			},
		},
	}
	mockBuilder.OnBuildMatch(mock.MatchedBy(func(wfClosure *core.CompiledWorkflowClosure) bool {
		return proto.Equal(wfClosure, &workflowClosure)
	}), mock.MatchedBy(func(inputs *core.LiteralMap) bool {
		return proto.Equal(inputs, testInputs)
	}), mock.MatchedBy(func(executionID *core.WorkflowExecutionIdentifier) bool {
		return proto.Equal(executionID, execID)
	}), namespace).Return(flyteWf, nil)
	executor := K8sWorkflowExecutor{
		config:           mockRuntime,
		workflowBuilder:  &mockBuilder,
		executionCluster: getFakeExecutionCluster(),
	}

	resp, err := executor.Execute(context.TODO(), interfaces.ExecutionData{
		Namespace:               namespace,
		ExecutionID:             execID,
		ReferenceWorkflowName:   "ref_workflow_name",
		ReferenceLaunchPlanName: "ref_lp_name",
		WorkflowClosure:         &workflowClosure,
		ExecutionParameters: interfaces.ExecutionParameters{
			Inputs: testInputs,
			ExecutionConfig: &admin.WorkflowExecutionConfig{
				SecurityContext: &core.SecurityContext{
					RunAs: &core.Identity{
						IamRole:           testRoleSc,
						K8SServiceAccount: testK8sServiceAccountSc,
					},
				},
			},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, resp.Cluster, clusterID)
}

func TestExecute_AlreadyExists(t *testing.T) {
	fakeFlyteWorkflow := FakeFlyteWorkflow{}
	fakeFlyteWorkflow.createCallback = func(flyteWorkflow *v1alpha1.FlyteWorkflow, opts v1.CreateOptions) (*v1alpha1.FlyteWorkflow, error) {
		return nil, k8_api_err.NewAlreadyExists(schema.GroupResource{}, "")
	}
	fakeFlyteWF.flyteWorkflowsCallback = func(ns string) v1alpha12.FlyteWorkflowInterface {
		assert.Equal(t, namespace, ns)
		return &fakeFlyteWorkflow
	}

	mockApplicationConfig := runtimeMocks.MockApplicationProvider{}
	mockApplicationConfig.SetTopLevelConfig(runtimeInterfaces.ApplicationConfig{
		UseOffloadedWorkflowClosure: false,
	})
	mockRuntime := runtimeMocks.NewMockConfigurationProvider(&mockApplicationConfig, nil, nil, nil, nil, nil)

	mockBuilder := mocks.FlyteWorkflowBuilder{}
	mockBuilder.OnBuildMatch(mock.Anything, mock.Anything, mock.Anything, namespace).Return(flyteWf, nil)
	executor := K8sWorkflowExecutor{
		config:           mockRuntime,
		workflowBuilder:  &mockBuilder,
		executionCluster: getFakeExecutionCluster(),
	}

	resp, err := executor.Execute(context.TODO(), interfaces.ExecutionData{
		Namespace:               namespace,
		ExecutionID:             execID,
		ReferenceWorkflowName:   "ref_workflow_name",
		ReferenceLaunchPlanName: "ref_lp_name",
		ExecutionParameters: interfaces.ExecutionParameters{
			ExecutionConfig: &admin.WorkflowExecutionConfig{
				SecurityContext: &core.SecurityContext{
					RunAs: &core.Identity{
						IamRole:           testRoleSc,
						K8SServiceAccount: testK8sServiceAccountSc,
					},
				},
			},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, resp.Cluster, clusterID)
}

func TestExecute_MiscError(t *testing.T) {
	fakeFlyteWorkflow := FakeFlyteWorkflow{}
	fakeFlyteWorkflow.createCallback = func(flyteWorkflow *v1alpha1.FlyteWorkflow, opts v1.CreateOptions) (*v1alpha1.FlyteWorkflow, error) {
		return nil, errors.New("call failed")
	}
	fakeFlyteWF.flyteWorkflowsCallback = func(ns string) v1alpha12.FlyteWorkflowInterface {
		assert.Equal(t, namespace, ns)
		return &fakeFlyteWorkflow
	}

	mockApplicationConfig := runtimeMocks.MockApplicationProvider{}
	mockApplicationConfig.SetTopLevelConfig(runtimeInterfaces.ApplicationConfig{
		UseOffloadedWorkflowClosure: false,
	})
	mockRuntime := runtimeMocks.NewMockConfigurationProvider(&mockApplicationConfig, nil, nil, nil, nil, nil)

	mockBuilder := mocks.FlyteWorkflowBuilder{}
	mockBuilder.OnBuildMatch(mock.Anything, mock.Anything, mock.Anything, namespace).Return(flyteWf, nil)
	executor := K8sWorkflowExecutor{
		config:           mockRuntime,
		workflowBuilder:  &mockBuilder,
		executionCluster: getFakeExecutionCluster(),
	}

	_, err := executor.Execute(context.TODO(), interfaces.ExecutionData{
		Namespace:               namespace,
		ExecutionID:             execID,
		ReferenceWorkflowName:   "ref_workflow_name",
		ReferenceLaunchPlanName: "ref_lp_name",
		ExecutionParameters: interfaces.ExecutionParameters{
			ExecutionConfig: &admin.WorkflowExecutionConfig{
				SecurityContext: &core.SecurityContext{
					RunAs: &core.Identity{
						IamRole:           testRoleSc,
						K8SServiceAccount: testK8sServiceAccountSc,
					},
				},
			},
		},
	})
	assert.EqualError(t, err, "failed to create workflow in propeller call failed")
}

func TestAbort(t *testing.T) {
	fakeFlyteWorkflow := FakeFlyteWorkflow{}
	fakeFlyteWorkflow.deleteCallback = func(name string, options *v1.DeleteOptions) error {
		assert.Equal(t, execID.Name, name)
		assert.Equal(t, options.PropagationPolicy, &deletePropagationBackground)
		return nil
	}
	fakeFlyteWF.flyteWorkflowsCallback = func(ns string) v1alpha12.FlyteWorkflowInterface {
		assert.Equal(t, namespace, ns)
		return &fakeFlyteWorkflow
	}
	executor := K8sWorkflowExecutor{
		executionCluster: getFakeExecutionCluster(),
	}
	err := executor.Abort(context.TODO(), interfaces.AbortData{
		Namespace:   namespace,
		ExecutionID: execID,
		Cluster:     clusterID,
	})
	assert.NoError(t, err)
}

func TestAbort_Notfound(t *testing.T) {
	fakeFlyteWorkflow := FakeFlyteWorkflow{}
	fakeFlyteWorkflow.deleteCallback = func(name string, options *v1.DeleteOptions) error {
		return k8_api_err.NewNotFound(schema.GroupResource{
			Group:    "foo",
			Resource: "bar",
		}, execID.Name)
	}
	fakeFlyteWF.flyteWorkflowsCallback = func(ns string) v1alpha12.FlyteWorkflowInterface {
		assert.Equal(t, namespace, ns)
		return &fakeFlyteWorkflow
	}
	executor := K8sWorkflowExecutor{
		executionCluster: getFakeExecutionCluster(),
	}
	err := executor.Abort(context.TODO(), interfaces.AbortData{
		Namespace:   namespace,
		ExecutionID: execID,
		Cluster:     clusterID,
	})
	assert.NoError(t, err)
}

func TestAbort_MiscError(t *testing.T) {
	fakeFlyteWorkflow := FakeFlyteWorkflow{}
	fakeFlyteWorkflow.deleteCallback = func(name string, options *v1.DeleteOptions) error {
		return errors.New("call failed")
	}
	fakeFlyteWF.flyteWorkflowsCallback = func(ns string) v1alpha12.FlyteWorkflowInterface {
		assert.Equal(t, namespace, ns)
		return &fakeFlyteWorkflow
	}
	executor := K8sWorkflowExecutor{
		executionCluster: getFakeExecutionCluster(),
	}
	err := executor.Abort(context.TODO(), interfaces.AbortData{
		Namespace:   namespace,
		ExecutionID: execID,
		Cluster:     clusterID,
	})
	assert.EqualError(t, err, "failed to terminate execution: project:\"proj\" domain:\"domain\" name:\"name\"  with err call failed")
}

func TestExecute_OffloadWorkflowClosure(t *testing.T) {
	offloadedFlyteWf := &v1alpha1.FlyteWorkflow{
		ExecutionID: v1alpha1.ExecutionID{
			WorkflowExecutionIdentifier: execID,
		},
		WorkflowSpec: &v1alpha1.WorkflowSpec{},
		Tasks:        make(map[v1alpha1.TaskID]*v1alpha1.TaskSpec),
		SubWorkflows: make(map[v1alpha1.WorkflowID]*v1alpha1.WorkflowSpec),
	}

	mockApplicationConfig := runtimeMocks.MockApplicationProvider{}
	mockApplicationConfig.SetTopLevelConfig(runtimeInterfaces.ApplicationConfig{
		UseOffloadedWorkflowClosure: true,
	})
	mockRuntime := runtimeMocks.NewMockConfigurationProvider(&mockApplicationConfig, nil, nil, nil, nil, nil)

	mockBuilder := mocks.FlyteWorkflowBuilder{}
	workflowClosure := core.CompiledWorkflowClosure{
		Primary: &core.CompiledWorkflow{
			Template: &core.WorkflowTemplate{
				Id: &core.Identifier{
					Project: "p",
					Domain:  "d",
					Name:    "n",
					Version: "version",
				},
			},
		},
		SubWorkflows: []*core.CompiledWorkflow{},
		Tasks:        []*core.CompiledTask{},
	}
	mockBuilder.OnBuildMatch(mock.MatchedBy(func(wfClosure *core.CompiledWorkflowClosure) bool {
		return proto.Equal(wfClosure, &workflowClosure)
	}), mock.MatchedBy(func(inputs *core.LiteralMap) bool {
		return proto.Equal(inputs, testInputs)
	}), mock.MatchedBy(func(executionID *core.WorkflowExecutionIdentifier) bool {
		return proto.Equal(executionID, execID)
	}), namespace).Return(offloadedFlyteWf, nil)
	executor := K8sWorkflowExecutor{
		config:           mockRuntime,
		workflowBuilder:  &mockBuilder,
		executionCluster: getFakeExecutionCluster(),
	}

	assert.NotNil(t, offloadedFlyteWf.WorkflowSpec)
	assert.NotNil(t, offloadedFlyteWf.Tasks)
	assert.NotNil(t, offloadedFlyteWf.SubWorkflows)

	resp, err := executor.Execute(context.TODO(), interfaces.ExecutionData{
		Namespace:               namespace,
		ExecutionID:             execID,
		ReferenceWorkflowName:   "ref_workflow_name",
		ReferenceLaunchPlanName: "ref_lp_name",
		WorkflowClosure:         &workflowClosure,
		ExecutionParameters: interfaces.ExecutionParameters{
			Inputs: testInputs,
			ExecutionConfig: &admin.WorkflowExecutionConfig{
				SecurityContext: &core.SecurityContext{
					RunAs: &core.Identity{
						IamRole:           testRoleSc,
						K8SServiceAccount: testK8sServiceAccountSc,
					},
				},
			},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, resp.Cluster, clusterID)

	assert.Nil(t, offloadedFlyteWf.WorkflowSpec)
	assert.Nil(t, offloadedFlyteWf.Tasks)
	assert.Nil(t, offloadedFlyteWf.SubWorkflows)
}
