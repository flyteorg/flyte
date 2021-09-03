package mocks

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/async/schedule/aws"
	"github.com/flyteorg/flyteadmin/pkg/async/schedule/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

type AddScheduleFunc func(ctx context.Context, input interfaces.AddScheduleInput) error
type RemoveScheduleFunc func(ctx context.Context, input interfaces.RemoveScheduleInput) error
type MockEventScheduler struct {
	addScheduleFunc    AddScheduleFunc
	removeScheduleFunc RemoveScheduleFunc
}

func (s *MockEventScheduler) CreateScheduleInput(ctx context.Context, appConfig *runtimeInterfaces.SchedulerConfig,
	identifier core.Identifier, schedule *admin.Schedule) (interfaces.AddScheduleInput, error) {
	payload, _ := aws.SerializeScheduleWorkflowPayload(
		schedule.GetKickoffTimeInputArg(),
		admin.NamedEntityIdentifier{
			Project: identifier.Project,
			Domain:  identifier.Domain,
			Name:    identifier.Name,
		})
	return interfaces.AddScheduleInput{Identifier: identifier, ScheduleExpression: *schedule, Payload: payload}, nil
}

func (s *MockEventScheduler) AddSchedule(ctx context.Context, input interfaces.AddScheduleInput) error {
	if s.addScheduleFunc != nil {
		return s.addScheduleFunc(ctx, input)
	}
	return nil
}

func (s *MockEventScheduler) SetAddScheduleFunc(addScheduleFunc AddScheduleFunc) {
	s.addScheduleFunc = addScheduleFunc
}

func (s *MockEventScheduler) RemoveSchedule(ctx context.Context, input interfaces.RemoveScheduleInput) error {
	if s.removeScheduleFunc != nil {
		return s.removeScheduleFunc(ctx, input)
	}
	return nil
}

func (s *MockEventScheduler) SetRemoveScheduleFunc(removeScheduleFunc RemoveScheduleFunc) {
	s.removeScheduleFunc = removeScheduleFunc
}

func NewMockEventScheduler() interfaces.EventScheduler {
	return &MockEventScheduler{}
}
