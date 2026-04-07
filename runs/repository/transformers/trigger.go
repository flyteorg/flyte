package transformers

import (
	"context"
	"database/sql"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	taskpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	triggerpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// ToTriggerKey creates an interfaces.TriggerNameKey from a common.TriggerName proto.
func ToTriggerKey(name *common.TriggerName) interfaces.TriggerNameKey {
	return interfaces.TriggerNameKey{
		Project:  name.GetProject(),
		Domain:   name.GetDomain(),
		TaskName: name.GetTaskName(),
		Name:     name.GetName(),
	}
}

// NewTriggerModel builds a models.Trigger from a deploy request.
// The caller supplies the TriggerIdentifier (including the expected revision for optimistic locking).
func NewTriggerModel(
	ctx context.Context,
	id *common.TriggerIdentifier,
	spec *triggerpb.TriggerSpec,
	automationSpec *taskpb.TriggerAutomationSpec,
) (*models.Trigger, error) {
	// TODO(nary): populate with real caller identity after adding auth
	deployedBy := sql.NullString{}

	specBytes, err := proto.Marshal(spec)
	if err != nil {
		return nil, err
	}

	if automationSpec == nil {
		automationSpec = &taskpb.TriggerAutomationSpec{}
	}
	if automationSpec.GetSchedule() != nil {
		automationSpec.Type = taskpb.TriggerAutomationSpecType_TYPE_SCHEDULE
	} else {
		automationSpec.Type = taskpb.TriggerAutomationSpecType_TYPE_NONE
	}

	automationSpecBytes, err := proto.Marshal(automationSpec)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	name := id.GetName()
	m := &models.Trigger{
		Project:        name.GetProject(),
		Domain:         name.GetDomain(),
		TaskName:       name.GetTaskName(),
		Name:           name.GetName(),
		Spec:           specBytes,
		AutomationSpec: automationSpecBytes,
		TaskVersion:    spec.GetTaskVersion(),
		Active:         spec.GetActive(),
		AutomationType: automationSpec.GetType().String(),
		DeployedBy:     deployedBy,
		UpdatedBy:      deployedBy,
		DeployedAt:     now,
		UpdatedAt:      now,
		Description:    newNullString(spec.GetDescription()),
	}
	return m, nil
}

// TriggerModelToTriggerDetails converts a latest-state trigger row to the full proto response.
func TriggerModelToTriggerDetails(ctx context.Context, m *models.Trigger) (*triggerpb.TriggerDetails, error) {
	automationSpec, err := UnmarshalAutomationSpec(m.AutomationSpec, m.AutomationType)
	if err != nil {
		logger.Errorf(ctx, "failed to unmarshal automation spec: %v", err)
		return nil, err
	}

	var spec triggerpb.TriggerSpec
	if err := proto.Unmarshal(m.Spec, &spec); err != nil {
		logger.Errorf(ctx, "failed to unmarshal trigger spec: %v", err)
		return nil, err
	}
	// Overlay denormalized fields that may have changed since the spec was serialized.
	spec.Active = m.Active
	spec.TaskVersion = m.TaskVersion
	if m.Description.Valid {
		spec.Description = m.Description.String
	}

	details := &triggerpb.TriggerDetails{
		Id:   triggerIdentifierFromTrigger(m),
		Spec: &spec,
		Status: &triggerpb.TriggerStatus{
			DeployedAt: timestamppb.New(m.DeployedAt),
			UpdatedAt:  timestamppb.New(m.UpdatedAt),
		},
		AutomationSpec: automationSpec,
	}
	if m.TriggeredAt.Valid {
		details.Status.TriggeredAt = timestamppb.New(m.TriggeredAt.Time)
	}
	if m.DeletedAt.Valid {
		details.Status.DeletedAt = timestamppb.New(m.DeletedAt.Time)
	}
	return details, nil
}

// TriggerRevisionModelToTriggerDetails converts an immutable revision row to the full proto response.
func TriggerRevisionModelToTriggerDetails(ctx context.Context, m *models.TriggerRevision) (*triggerpb.TriggerDetails, error) {
	automationSpec, err := UnmarshalAutomationSpec(m.AutomationSpec, m.AutomationType)
	if err != nil {
		logger.Errorf(ctx, "failed to unmarshal automation spec: %v", err)
		return nil, err
	}

	var spec triggerpb.TriggerSpec
	if err := proto.Unmarshal(m.Spec, &spec); err != nil {
		logger.Errorf(ctx, "failed to unmarshal trigger spec: %v", err)
		return nil, err
	}
	spec.Active = m.Active
	spec.TaskVersion = m.TaskVersion

	details := &triggerpb.TriggerDetails{
		Id:   triggerIdentifierFromTriggerRevision(m),
		Spec: &spec,
		Status: &triggerpb.TriggerStatus{
			DeployedAt:  timestamppb.New(m.DeployedAt),
			UpdatedAt:   timestamppb.New(m.UpdatedAt),
			TriggeredAt: nullTimeToTimestamppb(m.TriggeredAt),
			DeletedAt:   nullTimeToTimestamppb(m.DeletedAt),
		},
		AutomationSpec: automationSpec,
	}
	return details, nil
}

// TriggerModelsToTriggers converts latest-state trigger rows to the list-view proto.
func TriggerModelsToTriggers(ctx context.Context, ms []*models.Trigger) ([]*triggerpb.Trigger, error) {
	out := make([]*triggerpb.Trigger, 0, len(ms))
	for _, m := range ms {
		automationSpec, err := UnmarshalAutomationSpec(m.AutomationSpec, m.AutomationType)
		if err != nil {
			logger.Errorf(ctx, "failed to unmarshal automation spec: %v", err)
			return nil, err
		}
		t := &triggerpb.Trigger{
			Id: triggerIdentifierFromTrigger(m),
			Status: &triggerpb.TriggerStatus{
				DeployedAt: timestamppb.New(m.DeployedAt),
				UpdatedAt:  timestamppb.New(m.UpdatedAt),
			},
			Active:         m.Active,
			AutomationSpec: automationSpec,
		}
		if m.TriggeredAt.Valid {
			t.Status.TriggeredAt = timestamppb.New(m.TriggeredAt.Time)
		}
		out = append(out, t)
	}
	return out, nil
}

// TriggerRevisionModelsToTriggerRevisions converts revision rows to the revision-history proto.
func TriggerRevisionModelsToTriggerRevisions(_ context.Context, ms []*models.TriggerRevision) ([]*triggerpb.TriggerRevision, error) {
	out := make([]*triggerpb.TriggerRevision, 0, len(ms))
	for _, m := range ms {
		// createdAt is the latest of deployed_at / updated_at / deleted_at —
		// effectively the wall-clock time this revision row was committed.
		createdAt := m.DeployedAt
		if m.UpdatedAt.After(createdAt) {
			createdAt = m.UpdatedAt
		}
		if m.DeletedAt.Valid && m.DeletedAt.Time.After(createdAt) {
			createdAt = m.DeletedAt.Time
		}

		rev := &triggerpb.TriggerRevision{
			Id: triggerIdentifierFromTriggerRevision(m),
			Status: &triggerpb.TriggerStatus{
				DeployedAt:  timestamppb.New(m.DeployedAt),
				UpdatedAt:   timestamppb.New(m.UpdatedAt),
				TriggeredAt: nullTimeToTimestamppb(m.TriggeredAt),
				DeletedAt:   nullTimeToTimestamppb(m.DeletedAt),
			},
			Action:    triggerpb.TriggerRevisionAction(triggerpb.TriggerRevisionAction_value[m.Action]),
			CreatedAt: timestamppb.New(createdAt),
		}
		out = append(out, rev)
	}
	return out, nil
}

// UnmarshalAutomationSpec deserializes stored automation spec bytes.
func UnmarshalAutomationSpec(specBytes []byte, automationType string) (*taskpb.TriggerAutomationSpec, error) {
	var spec taskpb.TriggerAutomationSpec
	if automationType != "" {
		if v, ok := taskpb.TriggerAutomationSpecType_value[automationType]; ok {
			spec.Type = taskpb.TriggerAutomationSpecType(v)
		}
	}
	if len(specBytes) == 0 {
		return &spec, nil
	}
	return &spec, proto.Unmarshal(specBytes, &spec)
}

func triggerIdentifierFromTrigger(m *models.Trigger) *common.TriggerIdentifier {
	return &common.TriggerIdentifier{
		Name: &common.TriggerName{
			Project:  m.Project,
			Domain:   m.Domain,
			TaskName: m.TaskName,
			Name:     m.Name,
		},
		Revision: m.LatestRevision,
	}
}

func triggerIdentifierFromTriggerRevision(m *models.TriggerRevision) *common.TriggerIdentifier {
	return &common.TriggerIdentifier{
		Name: &common.TriggerName{
			Project:  m.Project,
			Domain:   m.Domain,
			TaskName: m.TaskName,
			Name:     m.Name,
		},
		Revision: m.Revision,
	}
}

func nullTimeToTimestamppb(t sql.NullTime) *timestamppb.Timestamp {
	if t.Valid {
		return timestamppb.New(t.Time)
	}
	return nil
}

