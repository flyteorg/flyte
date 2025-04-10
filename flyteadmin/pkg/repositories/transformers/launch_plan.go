package transformers

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func CreateLaunchPlan(
	request *admin.LaunchPlanCreateRequest,
	expectedOutputs *core.VariableMap) *admin.LaunchPlan {

	return &admin.LaunchPlan{
		Id:   request.GetId(),
		Spec: request.GetSpec(),
		Closure: &admin.LaunchPlanClosure{
			ExpectedInputs:  request.GetSpec().GetDefaultInputs(),
			ExpectedOutputs: expectedOutputs,
		},
	}
}

// Transforms a admin.LaunchPlan object to a LaunchPlan model
func CreateLaunchPlanModel(
	launchPlan *admin.LaunchPlan,
	workflowRepoID uint,
	digest []byte,
	initState admin.LaunchPlanState) (models.LaunchPlan, error) {
	spec, err := proto.Marshal(launchPlan.GetSpec())
	if err != nil {
		return models.LaunchPlan{}, errors.NewFlyteAdminError(codes.Internal, "Failed to serialize launch plan spec")
	}
	closure, err := proto.Marshal(launchPlan.GetClosure())
	if err != nil {
		return models.LaunchPlan{}, errors.NewFlyteAdminError(codes.Internal, "Failed to serialize launch plan closure")
	}

	var launchConditionType models.LaunchConditionType
	scheduleType := models.LaunchPlanScheduleTypeNONE
	if launchPlan.GetSpec().GetEntityMetadata() != nil && launchPlan.GetSpec().GetEntityMetadata().GetSchedule() != nil {
		if launchPlan.GetSpec().GetEntityMetadata().GetSchedule().GetCronExpression() != "" || launchPlan.GetSpec().GetEntityMetadata().GetSchedule().GetCronSchedule() != nil {
			scheduleType = models.LaunchPlanScheduleTypeCRON
			launchConditionType = models.LaunchConditionTypeSCHED
		} else if launchPlan.GetSpec().GetEntityMetadata().GetSchedule().GetRate() != nil {
			scheduleType = models.LaunchPlanScheduleTypeRATE
			launchConditionType = models.LaunchConditionTypeSCHED
		}
	}

	state := int32(initState)

	lpModel := models.LaunchPlan{
		LaunchPlanKey: models.LaunchPlanKey{
			Project: launchPlan.GetId().GetProject(),
			Domain:  launchPlan.GetId().GetDomain(),
			Name:    launchPlan.GetId().GetName(),
			Version: launchPlan.GetId().GetVersion(),
		},
		Spec:         spec,
		State:        &state,
		Closure:      closure,
		WorkflowID:   workflowRepoID,
		Digest:       digest,
		ScheduleType: scheduleType,
	}
	if launchConditionType != "" {
		lpModel.LaunchConditionType = &launchConditionType
	}
	return lpModel, nil
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
	createdAt := timestamppb.New(model.CreatedAt)
	updatedAt := timestamppb.New(model.UpdatedAt)
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
