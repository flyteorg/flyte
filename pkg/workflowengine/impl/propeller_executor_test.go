package impl

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lyft/flyteadmin/pkg/workflowengine/interfaces"

	"github.com/lyft/flytestdlib/promutils"

	"errors"

	flyte_admin_error "github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	flyteclient "github.com/lyft/flytepropeller/pkg/client/clientset/versioned"
	v1alpha12 "github.com/lyft/flytepropeller/pkg/client/clientset/versioned/typed/flyteworkflow/v1alpha1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	k8_api_err "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var fakeFlyteWF = FakeFlyteWorkflowV1alpha1{}
var scope = promutils.NewTestScope()
var propellerTestMetrics = newPropellerMetrics(scope)

var roleNameKey = "iam.amazonaws.com/role"

var acceptedAt = time.Now()

func getFlytePropellerForTest(clientSet flyteclient.Interface, builder *FlyteWorkflowBuilderTest) *FlytePropeller {
	return &FlytePropeller{
		client:      clientSet,
		builder:     builder,
		roleNameKey: roleNameKey,
		metrics:     propellerTestMetrics,
	}
}

type buildFlyteWorkflowFunc func(
	wfClosure *core.CompiledWorkflowClosure, inputs *core.LiteralMap, executionID *core.WorkflowExecutionIdentifier,
	namespace string) (*v1alpha1.FlyteWorkflow, error)
type FlyteWorkflowBuilderTest struct {
	buildCallback buildFlyteWorkflowFunc
}

func (b *FlyteWorkflowBuilderTest) BuildFlyteWorkflow(
	wfClosure *core.CompiledWorkflowClosure, inputs *core.LiteralMap, executionID *core.WorkflowExecutionIdentifier,
	namespace string) (*v1alpha1.FlyteWorkflow, error) {
	if b.buildCallback != nil {
		return b.buildCallback(wfClosure, inputs, executionID, namespace)
	}
	return &v1alpha1.FlyteWorkflow{}, nil
}

type createCallback func(*v1alpha1.FlyteWorkflow) (*v1alpha1.FlyteWorkflow, error)
type deleteCallback func(name string, options *v1.DeleteOptions) error
type FakeFlyteWorkflow struct {
	v1alpha12.FlyteWorkflowInterface
	createCallback createCallback
	deleteCallback deleteCallback
}

func (b *FakeFlyteWorkflow) Create(wf *v1alpha1.FlyteWorkflow) (*v1alpha1.FlyteWorkflow, error) {
	if b.createCallback != nil {
		return b.createCallback(wf)
	}
	return nil, nil
}

func (b *FakeFlyteWorkflow) Delete(name string, options *v1.DeleteOptions) error {
	if b.deleteCallback != nil {
		return b.deleteCallback(name, options)
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
}

func (b *FakeK8FlyteClient) FlyteworkflowV1alpha1() v1alpha12.FlyteworkflowV1alpha1Interface {
	return &fakeFlyteWF
}

func TestExecuteWorkflowHappyCase(t *testing.T) {
	client := FakeK8FlyteClient{}
	fakeFlyteWorkflow := FakeFlyteWorkflow{
		createCallback: func(workflow *v1alpha1.FlyteWorkflow) (*v1alpha1.FlyteWorkflow, error) {
			assert.EqualValues(t, map[string]string{
				"customlabel": "labelval",
			}, workflow.Labels)
			expectedAnnotations := map[string]string{
				"iam.amazonaws.com/role": "role-1",
				"customannotation":       "annotationval",
			}
			assert.EqualValues(t, expectedAnnotations, workflow.Annotations)
			return nil, nil
		},
	}
	fakeFlyteWF.flyteWorkflowsCallback = func(namespace string) v1alpha12.FlyteWorkflowInterface {
		assert.Equal(t, "p-d", namespace)
		return &fakeFlyteWorkflow
	}
	propeller := getFlytePropellerForTest(&client, &FlyteWorkflowBuilderTest{})

	err := propeller.ExecuteWorkflow(
		context.Background(),
		interfaces.ExecuteWorkflowInputs{
			ExecutionID: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "n",
			},
			WfClosure: core.CompiledWorkflowClosure{
				Primary: &core.CompiledWorkflow{
					Template: &core.WorkflowTemplate{},
				},
			},
			Reference: admin.LaunchPlan{
				Id: &core.Identifier{
					Project: "lp-p",
					Domain:  "lp-d",
				},
				Spec: &admin.LaunchPlanSpec{
					Role: "role-1",
				},
			},
			AcceptedAt: acceptedAt,
			Labels: map[string]string{
				"customlabel": "labelval",
			},
			Annotations: map[string]string{
				"customannotation": "annotationval",
			},
		})
	assert.Nil(t, err)
}

func TestExecuteWorkflowCallFailed(t *testing.T) {
	client := FakeK8FlyteClient{}
	fakeFlyteWorkflow := FakeFlyteWorkflow{
		createCallback: func(workflow *v1alpha1.FlyteWorkflow) (*v1alpha1.FlyteWorkflow, error) {
			expectedAnnotations := map[string]string{
				"iam.amazonaws.com/role": "role-1",
			}
			assert.EqualValues(t, expectedAnnotations, workflow.Annotations)
			return nil, errors.New("call failed")
		},
	}
	fakeFlyteWF.flyteWorkflowsCallback = func(namespace string) v1alpha12.FlyteWorkflowInterface {
		assert.Equal(t, "p-d", namespace)
		return &fakeFlyteWorkflow
	}
	propeller := getFlytePropellerForTest(&client, &FlyteWorkflowBuilderTest{})

	err := propeller.ExecuteWorkflow(
		context.Background(),
		interfaces.ExecuteWorkflowInputs{
			ExecutionID: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "n",
			},
			WfClosure: core.CompiledWorkflowClosure{
				Primary: &core.CompiledWorkflow{
					Template: &core.WorkflowTemplate{},
				},
			},
			Reference: admin.LaunchPlan{
				Id: &core.Identifier{
					Project: "p",
					Domain:  "d",
				},
				Spec: &admin.LaunchPlanSpec{
					Role: "role-1",
				},
			},
			AcceptedAt: acceptedAt,
		})

	assert.NotNil(t, err)
	assert.EqualError(t, err, "failed to create workflow in propeller call failed")
	assert.Equal(t, codes.Internal, err.(flyte_admin_error.FlyteAdminError).Code())

}

func TestExecuteWorkflowAlreadyExistsNoError(t *testing.T) {
	client := FakeK8FlyteClient{}
	fakeFlyteWorkflow := FakeFlyteWorkflow{
		createCallback: func(workflow *v1alpha1.FlyteWorkflow) (*v1alpha1.FlyteWorkflow, error) {
			expectedAnnotations := map[string]string{
				"iam.amazonaws.com/role": "role-1",
			}
			assert.EqualValues(t, expectedAnnotations, workflow.Annotations)
			return nil, k8_api_err.NewAlreadyExists(schema.GroupResource{}, "")
		},
	}
	fakeFlyteWF.flyteWorkflowsCallback = func(namespace string) v1alpha12.FlyteWorkflowInterface {
		assert.Equal(t, "p-d", namespace)
		return &fakeFlyteWorkflow
	}
	propeller := getFlytePropellerForTest(&client, &FlyteWorkflowBuilderTest{})

	err := propeller.ExecuteWorkflow(
		context.Background(),
		interfaces.ExecuteWorkflowInputs{
			ExecutionID: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "n",
			},
			WfClosure: core.CompiledWorkflowClosure{
				Primary: &core.CompiledWorkflow{
					Template: &core.WorkflowTemplate{},
				},
			},
			Reference: admin.LaunchPlan{
				Id: &core.Identifier{
					Project: "p",
					Domain:  "d",
				},
				Spec: &admin.LaunchPlanSpec{
					Role: "role-1",
				},
			},
			AcceptedAt: acceptedAt,
		})

	assert.Nil(t, err)
}

func TestExecuteWorkflowBuildFailed(t *testing.T) {
	client := FakeK8FlyteClient{}
	builder := FlyteWorkflowBuilderTest{}
	builder.buildCallback = func(wfClosure *core.CompiledWorkflowClosure, inputs *core.LiteralMap,
		executionID *core.WorkflowExecutionIdentifier, namespace string) (*v1alpha1.FlyteWorkflow, error) {
		return nil, flyte_admin_error.NewFlyteAdminError(codes.Aborted, "abort")
	}
	propeller := getFlytePropellerForTest(&client, &builder)
	identifier := core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "project",
		Domain:       "domain",
		Name:         "name",
		Version:      "version",
	}
	err := propeller.ExecuteWorkflow(
		context.Background(),
		interfaces.ExecuteWorkflowInputs{
			ExecutionID: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "n",
			},
			WfClosure: core.CompiledWorkflowClosure{
				Primary: &core.CompiledWorkflow{
					Template: &core.WorkflowTemplate{
						Id: &identifier,
					},
				},
			},
			Reference: admin.LaunchPlan{
				Id: &core.Identifier{
					Project: "p",
					Domain:  "d",
				},
			},
			AcceptedAt: acceptedAt,
		})
	assert.NotNil(t, err)
	assert.EqualError(t, err, fmt.Sprintf("failed to build the workflow [%+v] abort", &identifier))
	assert.Equal(t, codes.Internal, err.(flyte_admin_error.FlyteAdminError).Code())
}

func TestExecuteWorkflowRoleKeyNotRequired(t *testing.T) {
	client := FakeK8FlyteClient{}
	builder := FlyteWorkflowBuilderTest{}

	fakeFlyteWorkflow := FakeFlyteWorkflow{
		createCallback: func(workflow *v1alpha1.FlyteWorkflow) (*v1alpha1.FlyteWorkflow, error) {
			return nil, nil
		},
	}
	fakeFlyteWF.flyteWorkflowsCallback = func(namespace string) v1alpha12.FlyteWorkflowInterface {
		assert.Equal(t, "p-d", namespace)
		return &fakeFlyteWorkflow
	}
	propeller := getFlytePropellerForTest(&client, &builder)
	err := propeller.ExecuteWorkflow(
		context.Background(),
		interfaces.ExecuteWorkflowInputs{
			ExecutionID: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "n",
			},
			WfClosure: core.CompiledWorkflowClosure{
				Primary: &core.CompiledWorkflow{
					Template: &core.WorkflowTemplate{},
				},
			},
			Reference: admin.LaunchPlan{
				Id: &core.Identifier{
					Project: "p",
					Domain:  "d",
				},
			},
			AcceptedAt: acceptedAt,
		})
	assert.Nil(t, err)
}

func TestTerminateExecution(t *testing.T) {
	client := FakeK8FlyteClient{}
	builder := FlyteWorkflowBuilderTest{}

	fakeFlyteWorkflow := FakeFlyteWorkflow{
		deleteCallback: func(name string, options *v1.DeleteOptions) error {
			assert.Equal(t, "n", name)
			return errors.New("expected error")
		},
	}
	fakeFlyteWF.flyteWorkflowsCallback = func(namespace string) v1alpha12.FlyteWorkflowInterface {
		assert.Equal(t, "p-d", namespace)
		return &fakeFlyteWorkflow
	}
	propeller := getFlytePropellerForTest(&client, &builder)

	err := propeller.TerminateWorkflowExecution(context.Background(), &core.WorkflowExecutionIdentifier{
		Project: "p",
		Domain:  "d",
		Name:    "n",
	})
	assert.EqualError(t, err,
		"failed to terminate execution: project:\"p\" domain:\"d\" name:\"n\"  with err expected error")
}

func TestAddPermissions(t *testing.T) {
	propeller := getFlytePropellerForTest(&FakeK8FlyteClient{}, &FlyteWorkflowBuilderTest{})
	flyteWf := v1alpha1.FlyteWorkflow{}
	propeller.addPermissions(admin.LaunchPlan{
		Spec: &admin.LaunchPlanSpec{
			Auth: &admin.Auth{
				Method: &admin.Auth_AssumableIamRole{
					AssumableIamRole: "rollie-pollie",
				},
			},
			Role: "ignore-me",
		},
	}, &flyteWf)
	assert.EqualValues(t, flyteWf.Annotations, map[string]string{
		roleNameKey: "rollie-pollie",
	})

	flyteWf = v1alpha1.FlyteWorkflow{}
	propeller.addPermissions(admin.LaunchPlan{
		Spec: &admin.LaunchPlanSpec{
			Role: "rollie-pollie",
		},
	}, &flyteWf)
	assert.EqualValues(t, flyteWf.Annotations, map[string]string{
		roleNameKey: "rollie-pollie",
	})

	flyteWf = v1alpha1.FlyteWorkflow{}
	propeller.addPermissions(admin.LaunchPlan{
		Spec: &admin.LaunchPlanSpec{
			Auth: &admin.Auth{
				Method: &admin.Auth_KubernetesServiceAccount{
					KubernetesServiceAccount: "service-account",
				},
			},
		},
	}, &flyteWf)
	assert.Equal(t, "service-account", flyteWf.ServiceAccountName)
	assert.Empty(t, flyteWf.Annotations)
}
