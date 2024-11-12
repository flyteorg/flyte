package errors

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wI2L/jsondiff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

var identifier = core.Identifier{
	ResourceType: core.ResourceType_TASK,
	Project:      "testProj",
	Domain:       "domain",
	Name:         "t1",
	Version:      "ver",
}

func TestGrpcStatusError(t *testing.T) {

	msg := "some error"
	curPhase := "some phase"
	statusErr := NewAlreadyInTerminalStateError(context.Background(), msg, curPhase)
	assert.NotNil(t, statusErr)
	s, ok := status.FromError(statusErr)
	assert.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, s.Code())
	assert.Equal(t, "some error", s.Message())

	details, ok := s.Details()[0].(*admin.EventFailureReason)
	assert.True(t, ok)
	_, ok = details.GetReason().(*admin.EventFailureReason_AlreadyInTerminalState)
	assert.True(t, ok)
}

func TestNewIncompatibleClusterError(t *testing.T) {
	errorMsg := "foo"
	cluster := "C1"
	statusErr := NewIncompatibleClusterError(context.Background(), errorMsg, cluster)
	assert.NotNil(t, statusErr)
	s, ok := status.FromError(statusErr)
	assert.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, s.Code())
	assert.Equal(t, errorMsg, s.Message())

	details, ok := s.Details()[0].(*admin.EventFailureReason)
	assert.True(t, ok)
	_, ok = details.GetReason().(*admin.EventFailureReason_IncompatibleCluster)
	assert.True(t, ok)
}

func TestJsonDifferHasDiffError(t *testing.T) {
	oldSpec := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
		"four":  4,
		"five":  5,
	}
	newSpec := map[string]int{
		"five":  5,
		"four":  0,
		"three": 3,
		"two":   2,
		"one":   1,
	}
	diff, _ := jsondiff.Compare(oldSpec, newSpec)
	rdiff, _ := jsondiff.Compare(newSpec, oldSpec)
	rs := compareJsons(diff, rdiff)
	assert.Equal(t, "/four:\n\t- 4 \n\t+ 0", strings.Join(rs, "\n"))
}

func TestJsonDifferNoDiffError(t *testing.T) {
	oldSpec := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
		"four":  4,
		"five":  5,
	}
	newSpec := map[string]int{
		"five":  5,
		"four":  4,
		"three": 3,
		"two":   2,
		"one":   1,
	}
	diff, _ := jsondiff.Compare(oldSpec, newSpec)
	rdiff, _ := jsondiff.Compare(newSpec, oldSpec)
	rs := compareJsons(diff, rdiff)
	assert.Equal(t, "", strings.Join(rs, "\n"))
}

func TestNewTaskExistsDifferentStructureError(t *testing.T) {
	req := &admin.TaskCreateRequest{
		Id: &identifier,
	}

	oldTask := &core.CompiledTask{
		Template: &core.TaskTemplate{
			Target: &core.TaskTemplate_Container{
				Container: &core.Container{
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{
								Name:  core.Resources_CPU,
								Value: "150m",
							},
						},
					},
				},
			},
			Id: &identifier,
		},
	}

	newTask := &core.CompiledTask{
		Template: &core.TaskTemplate{
			Target: &core.TaskTemplate_Container{
				Container: &core.Container{
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{
								Name:  core.Resources_CPU,
								Value: "250m",
							},
						},
					},
				},
			},
			Id: &identifier,
		},
	}

	statusErr := NewTaskExistsDifferentStructureError(context.Background(), req, oldTask, newTask)
	assert.NotNil(t, statusErr)
	s, ok := status.FromError(statusErr)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, s.Code())
	assert.Equal(t, "t1 task with different structure already exists. (Please register a new version of the task):\n/template/Target/Container/resources/requests/0/value:\n\t- 150m \n\t+ 250m", s.Message())
}

func TestNewTaskExistsIdenticalStructureError(t *testing.T) {
	req := &admin.TaskCreateRequest{
		Id: &identifier,
	}
	statusErr := NewTaskExistsIdenticalStructureError(context.Background(), req)
	assert.NotNil(t, statusErr)
	s, ok := status.FromError(statusErr)
	assert.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, s.Code())
	assert.Equal(t, "task with identical structure already exists", s.Message())
}

func TestNewWorkflowExistsDifferentStructureError(t *testing.T) {
	identifier := core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "testProj",
		Domain:       "domain",
		Name:         "hello",
		Version:      "ver",
	}

	req := &admin.WorkflowCreateRequest{
		Id: &identifier,
	}

	oldWorkflow := &core.CompiledWorkflowClosure{
		Primary: &core.CompiledWorkflow{
			Connections: &core.ConnectionSet{
				Upstream: map[string]*core.ConnectionSet_IdList{
					"foo": &core.ConnectionSet_IdList{
						Ids: []string{"start-node"},
					},
					"end-node": &core.ConnectionSet_IdList{
						Ids: []string{"foo"},
					},
				},
			},
			Template: &core.WorkflowTemplate{
				Nodes: []*core.Node{
					&core.Node{
						Id:     "foo",
						Target: &core.Node_TaskNode{},
					},
				},
				Id: &identifier,
			},
		},
	}

	newWorkflow := &core.CompiledWorkflowClosure{
		Primary: &core.CompiledWorkflow{
			Connections: &core.ConnectionSet{
				Upstream: map[string]*core.ConnectionSet_IdList{
					"bar": &core.ConnectionSet_IdList{
						Ids: []string{"start-node"},
					},
					"end-node": &core.ConnectionSet_IdList{
						Ids: []string{"bar"},
					},
				},
			},
			Template: &core.WorkflowTemplate{
				Nodes: []*core.Node{
					&core.Node{
						Id:     "bar",
						Target: &core.Node_TaskNode{},
					},
				},
				Id: &identifier,
			},
		},
	}

	statusErr := NewWorkflowExistsDifferentStructureError(context.Background(), req, oldWorkflow, newWorkflow)
	assert.NotNil(t, statusErr)
	s, ok := status.FromError(statusErr)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, s.Code())
	assert.Equal(t, "hello workflow with different structure already exists. (Please register a new version of the workflow):\n/primary/connections/upstream/bar:\n\t- <nil> \n\t+ map[ids:[start-node]]\n/primary/connections/upstream/end-node/ids/0:\n\t- foo \n\t+ bar\n/primary/connections/upstream/foo:\n\t- map[ids:[start-node]] \n\t+ <nil>\n/primary/template/nodes/0/id:\n\t- foo \n\t+ bar", s.Message())

	details, ok := s.Details()[0].(*admin.CreateWorkflowFailureReason)
	assert.True(t, ok)
	_, ok = details.GetReason().(*admin.CreateWorkflowFailureReason_ExistsDifferentStructure)
	assert.True(t, ok)
}

func TestNewWorkflowExistsIdenticalStructureError(t *testing.T) {
	req := &admin.WorkflowCreateRequest{
		Id: &identifier,
	}
	statusErr := NewWorkflowExistsIdenticalStructureError(context.Background(), req)
	assert.NotNil(t, statusErr)
	s, ok := status.FromError(statusErr)
	assert.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, s.Code())
	assert.Equal(t, "workflow with identical structure already exists", s.Message())

	details, ok := s.Details()[0].(*admin.CreateWorkflowFailureReason)
	assert.True(t, ok)
	_, ok = details.GetReason().(*admin.CreateWorkflowFailureReason_ExistsIdenticalStructure)
	assert.True(t, ok)
}

func TestNewLaunchPlanExistsDifferentStructureError(t *testing.T) {
	identifier := core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Name:         "lp_name",
	}

	req := &admin.LaunchPlanCreateRequest{
		Id: &identifier,
	}

	oldLaunchPlan := &admin.LaunchPlan{
		Spec: &admin.LaunchPlanSpec{
			WorkflowId: &core.Identifier{
				Project: "testProj",
				Domain:  "domain",
				Name:    "lp_name",
				Version: "ver1",
			},
		},
		Id: &identifier,
	}

	newLaunchPlan := &admin.LaunchPlan{
		Spec: &admin.LaunchPlanSpec{
			WorkflowId: &core.Identifier{
				Project: "testProj",
				Domain:  "domain",
				Name:    "lp_name",
				Version: "ver2",
			},
		},
		Id: &identifier,
	}

	statusErr := NewLaunchPlanExistsDifferentStructureError(context.Background(), req, oldLaunchPlan.Spec, newLaunchPlan.Spec)
	assert.NotNil(t, statusErr)
	s, ok := status.FromError(statusErr)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, s.Code())
	assert.Equal(t, "lp_name launch plan with different structure already exists. (Please register a new version of the launch plan):\n/workflow_id/version:\n\t- ver1 \n\t+ ver2", s.Message())
}

func TestNewLaunchPlanExistsIdenticalStructureError(t *testing.T) {
	req := &admin.LaunchPlanCreateRequest{
		Id: &identifier,
	}
	statusErr := NewLaunchPlanExistsIdenticalStructureError(context.Background(), req)
	assert.NotNil(t, statusErr)
	s, ok := status.FromError(statusErr)
	assert.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, s.Code())
	assert.Equal(t, "launch plan with identical structure already exists", s.Message())
}

func TestIsDoesNotExistError(t *testing.T) {
	assert.True(t, IsDoesNotExistError(NewFlyteAdminError(codes.NotFound, "foo")))
}

func TestIsNotDoesNotExistError(t *testing.T) {
	assert.False(t, IsDoesNotExistError(NewFlyteAdminError(codes.Canceled, "foo")))
}

func TestIsNotDoesNotExistErrorBecauseOfNoneAdminError(t *testing.T) {
	assert.False(t, IsDoesNotExistError(errors.New("foo")))
}

func TestNewInactiveProjectError(t *testing.T) {
	err := NewInactiveProjectError(context.TODO(), identifier.GetProject())
	statusErr, ok := status.FromError(err)

	assert.True(t, ok)

	details, ok := statusErr.Details()[0].(*admin.InactiveProject)

	assert.True(t, ok)
	assert.Equal(t, identifier.GetProject(), details.Id)
}
