package aws

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"

	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/NYTimes/gizmo/pubsub/pubsubtest"
	flyteAdminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/mocks"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/ptypes"
)

const testKickoffTime = "kickoff time arg"

var testKickoffTimestamp = time.Date(2017, 12, 22, 18, 43, 48, 0, time.UTC)
var testIdentifier = admin.NamedEntityIdentifier{
	Name:    "name",
	Project: "project",
	Domain:  "domain",
}

var protoTestTimestamp, _ = ptypes.TimestampProto(testKickoffTimestamp)
var testKickoffTimeProtoLiteral = core.Literal{
	Value: &core.Literal_Scalar{
		Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{
				Primitive: &core.Primitive{
					Value: &core.Primitive_Datetime{
						Datetime: protoTestTimestamp,
					},
				},
			},
		},
	},
}

var executorScope promutils.Scope
var executorMetrics workflowExecutorMetrics

func init() {
	labeled.SetMetricKeys(contextutils.TaskIDKey)
	executorScope = promutils.NewScope("test_wflow_exec")
	executorMetrics = newWorkflowExecutorMetrics(executorScope)
}

func newWorkflowExecutorForTest(
	subscriber pubsub.Subscriber, executionManager interfaces.ExecutionInterface,
	launchPlanManager interfaces.LaunchPlanInterface) workflowExecutor {
	return workflowExecutor{
		subscriber:        subscriber,
		executionManager:  executionManager,
		launchPlanManager: launchPlanManager,
		metrics:           executorMetrics,
	}
}

func TestResolveKickoffTimeArg(t *testing.T) {
	scheduleRequest := ScheduledWorkflowExecutionRequest{
		KickoffTimeArg: testKickoffTime,
		KickoffTime:    testKickoffTimestamp,
	}
	launchPlan := admin.LaunchPlan{
		Closure: &admin.LaunchPlanClosure{
			ExpectedInputs: &core.ParameterMap{
				Parameters: map[string]*core.Parameter{
					testKickoffTime: {},
				},
			},
		},
	}
	executionRequest := admin.ExecutionCreateRequest{
		Project: testIdentifier.Project,
		Domain:  testIdentifier.Domain,
		Name:    testIdentifier.Name,
		Inputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{},
		},
	}
	testExecutor := newWorkflowExecutorForTest(nil, nil, nil)
	err := testExecutor.resolveKickoffTimeArg(scheduleRequest, launchPlan, &executionRequest)
	assert.Nil(t, err)
	assert.Contains(t, executionRequest.Inputs.Literals, testKickoffTime)
	assert.Equal(t, testKickoffTimeProtoLiteral,
		*executionRequest.Inputs.Literals[testKickoffTime])
}

func TestResolveKickoffTimeArg_NoKickoffTimeArg(t *testing.T) {
	scheduleRequest := ScheduledWorkflowExecutionRequest{
		KickoffTimeArg: testKickoffTime,
		KickoffTime:    testKickoffTimestamp,
	}
	launchPlan := admin.LaunchPlan{
		Closure: &admin.LaunchPlanClosure{
			ExpectedInputs: &core.ParameterMap{
				Parameters: map[string]*core.Parameter{
					"foo": {},
				},
			},
		},
	}
	executionRequest := admin.ExecutionCreateRequest{
		Project: testIdentifier.Project,
		Domain:  testIdentifier.Domain,
		Name:    testIdentifier.Name,
		Inputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{},
		},
	}
	testExecutor := newWorkflowExecutorForTest(nil, nil, nil)
	err := testExecutor.resolveKickoffTimeArg(scheduleRequest, launchPlan, &executionRequest)
	assert.Nil(t, err)
	assert.NotContains(t, executionRequest.Inputs.Literals, testKickoffTime)
}

func TestGetActiveLaunchPlanVersion(t *testing.T) {
	launchPlanNamedIdentifier := &admin.NamedEntityIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}
	launchPlanIdentifier := core.Identifier{
		Project: launchPlanNamedIdentifier.Project,
		Domain:  launchPlanNamedIdentifier.Domain,
		Name:    launchPlanNamedIdentifier.Name,
		Version: "foo",
	}

	launchPlanManager := mocks.NewMockLaunchPlanManager()
	launchPlanManager.(*mocks.MockLaunchPlanManager).SetListLaunchPlansCallback(
		func(ctx context.Context, request admin.ResourceListRequest) (
			*admin.LaunchPlanList, error) {
			assert.True(t, proto.Equal(launchPlanNamedIdentifier, request.Id))
			assert.Equal(t, "eq(state,1)", request.Filters)
			assert.Equal(t, uint32(1), request.Limit)
			return &admin.LaunchPlanList{
				LaunchPlans: []*admin.LaunchPlan{
					{
						Id: &launchPlanIdentifier,
					},
				},
			}, nil
		})
	testExecutor := newWorkflowExecutorForTest(nil, nil, launchPlanManager)
	launchPlan, err := testExecutor.getActiveLaunchPlanVersion(launchPlanNamedIdentifier)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&launchPlanIdentifier, launchPlan.Id))
}

func TestGetActiveLaunchPlanVersion_ManagerError(t *testing.T) {
	launchPlanIdentifier := &admin.NamedEntityIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}

	expectedErr := errors.New("expected error")
	launchPlanManager := mocks.NewMockLaunchPlanManager()
	launchPlanManager.(*mocks.MockLaunchPlanManager).SetListLaunchPlansCallback(
		func(ctx context.Context, request admin.ResourceListRequest) (
			*admin.LaunchPlanList, error) {
			return nil, expectedErr
		})
	testExecutor := newWorkflowExecutorForTest(nil, nil, launchPlanManager)
	_, err := testExecutor.getActiveLaunchPlanVersion(launchPlanIdentifier)
	assert.EqualError(t, err, expectedErr.Error())
}

func TestFormulateExecutionCreateRequest(t *testing.T) {
	launchPlanIdentifier := core.Identifier{
		Project: "foo",
		Domain:  "bar",
		Name:    "baz",
		Version: "12345",
	}
	launchPlan := admin.LaunchPlan{
		Spec: &admin.LaunchPlanSpec{
			WorkflowId: &core.Identifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
				Version: "version",
			},
		},
		Id: &launchPlanIdentifier,
	}
	testExecutor := newWorkflowExecutorForTest(nil, nil, nil)
	executionRequest := testExecutor.formulateExecutionCreateRequest(launchPlan, time.Unix(1543607788, 0))
	assert.Equal(t, "foo", executionRequest.Project)
	assert.Equal(t, "bar", executionRequest.Domain)
	assert.Equal(t, "a2k4s9v5j246kwmdmh4t", executionRequest.Name)

	assert.True(t, proto.Equal(&launchPlanIdentifier, executionRequest.Spec.LaunchPlan))
	assert.Equal(t, admin.ExecutionMetadata_SCHEDULED, executionRequest.Spec.Metadata.Mode)
	assert.Equal(t, int64(1543607788), executionRequest.Spec.Metadata.ScheduledAt.Seconds)
}

func TestRun(t *testing.T) {
	launchPlanIdentifier := &admin.NamedEntityIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}
	payload, _ := proto.Marshal(launchPlanIdentifier)
	firstScheduleWorkflowPayload := ScheduleWorkflowPayload{
		Time:           "2017-12-22T18:43:48Z",
		KickoffTimeArg: testKickoffTime,
		Payload:        payload,
	}
	secondScheduleWorkflowPayload := ScheduleWorkflowPayload{
		Time:    "2017-12-22T18:43:48Z",
		Payload: payload,
	}
	messages := make([]interface{}, 2)
	messages[0] = firstScheduleWorkflowPayload
	messages[1] = secondScheduleWorkflowPayload
	testSubscriber := pubsubtest.TestSubscriber{
		JSONMessages: messages,
	}
	testExecutionManager := mocks.MockExecutionManager{}
	var messagesSeen int
	testExecutionManager.SetCreateCallback(func(
		ctx context.Context, request admin.ExecutionCreateRequest, requestedAt time.Time) (
		*admin.ExecutionCreateResponse, error) {
		assert.Equal(t, "project", request.Project)
		assert.Equal(t, "domain", request.Domain)
		assert.Equal(t, "ar8fphnlc5wh9dksjncj", request.Name)
		if messagesSeen == 0 {
			assert.Contains(t, request.Inputs.Literals, testKickoffTime)
			assert.Equal(t, testKickoffTimeProtoLiteral, *request.Inputs.Literals[testKickoffTime])
		}
		messagesSeen++
		return &admin.ExecutionCreateResponse{}, nil
	})
	launchPlanManager := mocks.NewMockLaunchPlanManager()
	launchPlanManager.(*mocks.MockLaunchPlanManager).SetListLaunchPlansCallback(
		func(ctx context.Context, request admin.ResourceListRequest) (
			*admin.LaunchPlanList, error) {
			assert.Equal(t, "project", request.Id.Project)
			assert.Equal(t, "domain", request.Id.Domain)
			assert.Equal(t, "eq(state,1)", request.Filters)
			assert.Equal(t, uint32(1), request.Limit)
			return &admin.LaunchPlanList{
				LaunchPlans: []*admin.LaunchPlan{
					{
						Id: &core.Identifier{
							Project: "project",
							Domain:  "domain",
							Name:    "name",
							Version: "foo",
						},
						Spec: &admin.LaunchPlanSpec{
							WorkflowId: &core.Identifier{
								Project: "project",
								Domain:  "domain",
								Name:    "name",
								Version: "version",
							},
						},
						Closure: &admin.LaunchPlanClosure{
							ExpectedInputs: &core.ParameterMap{
								Parameters: map[string]*core.Parameter{
									testKickoffTime: {},
								},
							},
						},
					},
				},
			}, nil
		})
	testExecutor := newWorkflowExecutorForTest(&testSubscriber, &testExecutionManager, launchPlanManager)
	err := testExecutor.run()
	assert.Len(t, messages, messagesSeen)
	assert.Nil(t, err)
}

func TestStop(t *testing.T) {
	testSubscriber := pubsubtest.TestSubscriber{}
	testExecutor := newWorkflowExecutorForTest(&testSubscriber, nil, nil)
	assert.Nil(t, testExecutor.Stop())
}

func TestStop_Error(t *testing.T) {
	testSubscriber := pubsubtest.TestSubscriber{
		GivenStopError: errors.New("foo"),
	}
	testExecutor := newWorkflowExecutorForTest(&testSubscriber, nil, nil)
	err := testExecutor.Stop()
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
}
