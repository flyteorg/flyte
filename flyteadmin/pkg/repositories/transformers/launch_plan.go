package transformers

import (
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
)

func CreateLaunchPlan(
	request admin.LaunchPlanCreateRequest,
	expectedOutputs *core.VariableMap) admin.LaunchPlan {

	return admin.LaunchPlan{
		Id:   request.Id,
		Spec: request.Spec,
		Closure: &admin.LaunchPlanClosure{
			ExpectedInputs:  request.Spec.DefaultInputs,
			ExpectedOutputs: expectedOutputs,
		},
	}
}

// Transforms a admin.LaunchPlan object to a LaunchPlan model
func CreateLaunchPlanModel(
	launchPlan admin.LaunchPlan,
	workflowRepoID uint,
	digest []byte,
	initState admin.LaunchPlanState) (models.LaunchPlan, error) {
	spec, err := proto.Marshal(launchPlan.Spec)
	if err != nil {
		return models.LaunchPlan{}, errors.NewFlyteAdminError(codes.Internal, "Failed to serialize launch plan spec")
	}
	closure, err := proto.Marshal(launchPlan.Closure)
	if err != nil {
		return models.LaunchPlan{}, errors.NewFlyteAdminError(codes.Internal, "Failed to serialize launch plan closure")
	}

	scheduleType := models.LaunchPlanScheduleTypeNONE
	if launchPlan.Spec.EntityMetadata != nil && launchPlan.Spec.EntityMetadata.Schedule != nil {
		if launchPlan.Spec.EntityMetadata.Schedule.GetCronExpression() != "" {
			scheduleType = models.LaunchPlanScheduleTypeCRON
		} else if launchPlan.Spec.EntityMetadata.Schedule.GetRate() != nil {
			scheduleType = models.LaunchPlanScheduleTypeRATE
		}
	}

	state := int32(initState)

	return models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: launchPlan.Id.Project,
			Domain:  launchPlan.Id.Domain,
			Name:    launchPlan.Id.Name,
			Version: launchPlan.Id.Version,
		},
		Spec:         spec,
		State:        &state,
		Closure:      closure,
		WorkflowID:   workflowRepoID,
		Digest:       digest,
		ScheduleType: scheduleType,
	}, nil
}

// Transforms a LaunchPlanModel to a LaunchPlan
func FromLaunchPlanModel(model models.LaunchPlan) (*admin.LaunchPlan, error) {
	spec := &admin.LaunchPlanSpec{}
	err := proto.Unmarshal(model.Spec, spec)
	if err != nil {
		return nil, errors.NewFlyteAdminError(codes.Internal, "failed to unmarshal spec")
	}

	var closure admin.LaunchPlanClosure
	err = proto.Unmarshal(model.Closure, &closure)
	if err != nil {
		return nil, errors.NewFlyteAdminError(codes.Internal, "failed to unmarshal closure")
	}
	createdAt, err := ptypes.TimestampProto(model.CreatedAt)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to read created at")
	}
	updatedAt, err := ptypes.TimestampProto(model.UpdatedAt)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "failed to read updated at")
	}
	if model.State != nil {
		closure.State = admin.LaunchPlanState(*model.State)
	}
	closure.CreatedAt = createdAt
	closure.UpdatedAt = updatedAt

	id := core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      model.Project,
		Domain:       model.Domain,
		Name:         model.Name,
		Version:      model.Version,
	}

	return &admin.LaunchPlan{
		Id:      &id,
		Spec:    spec,
		Closure: &closure,
	}, nil
}

func FromLaunchPlanModels(launchPlanModels []models.LaunchPlan) ([]*admin.LaunchPlan, error) {
	launchPlans := make([]*admin.LaunchPlan, len(launchPlanModels))
	for idx, launchPlanModel := range launchPlanModels {
		launchPlan, err := FromLaunchPlanModel(launchPlanModel)
		if err != nil {
			return nil, err
		}
		launchPlans[idx] = launchPlan
	}
	return launchPlans, nil
}

func FromLaunchPlanModelsToIdentifiers(launchPlanModels []models.LaunchPlan) []*admin.NamedEntityIdentifier {
	ids := make([]*admin.NamedEntityIdentifier, len(launchPlanModels))
	for i, launchPlanModel := range launchPlanModels {
		ids[i] = &admin.NamedEntityIdentifier{
			Project: launchPlanModel.Project,
			Domain:  launchPlanModel.Domain,
			Name:    launchPlanModel.Name,
		}
	}
	return ids
}
