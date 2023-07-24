package transformers

import (
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
)

var active = int32(admin.LaunchPlanState_ACTIVE)

var expectedInputs = &core.ParameterMap{
	Parameters: map[string]*core.Parameter{
		"foo": {},
	},
}
var expectedOutputs = &core.VariableMap{
	Variables: map[string]*core.Variable{
		"baz": {},
	},
}

func getRequest() admin.LaunchPlanCreateRequest {
	request := testutils.GetLaunchPlanRequest()
	request.Spec.DefaultInputs = expectedInputs
	return request
}

func TestCreateLaunchPlan(t *testing.T) {
	request := getRequest()

	launchPlan := CreateLaunchPlan(request, expectedOutputs)
	assert.True(t, proto.Equal(
		&admin.LaunchPlan{
			Id:   request.Id,
			Spec: request.Spec,
			Closure: &admin.LaunchPlanClosure{
				ExpectedInputs:  expectedInputs,
				ExpectedOutputs: expectedOutputs,
			},
		}, &launchPlan))
}

func TestToLaunchPlanModel(t *testing.T) {
	lpRequest := getRequest()
	workflowID := uint(11)
	launchPlanDigest := []byte("launch plan")

	launchPlan := admin.LaunchPlan{
		Id:   lpRequest.Id,
		Spec: lpRequest.Spec,
		Closure: &admin.LaunchPlanClosure{
			ExpectedInputs:  expectedInputs,
			ExpectedOutputs: expectedOutputs,
		},
	}

	launchPlanModel, err := CreateLaunchPlanModel(launchPlan, workflowID, launchPlanDigest, admin.LaunchPlanState_INACTIVE)
	assert.NoError(t, err)
	assert.Equal(t, "project", launchPlanModel.Project)
	assert.Equal(t, "domain", launchPlanModel.Domain)
	assert.Equal(t, "name", launchPlanModel.Name)
	assert.Equal(t, "version", launchPlanModel.Version)
	assert.Equal(t, workflowID, launchPlanModel.WorkflowID)

	expectedSpec, _ := proto.Marshal(lpRequest.Spec)
	assert.Equal(t, expectedSpec, launchPlanModel.Spec)
	assert.Equal(t, models.LaunchPlanScheduleTypeNONE, launchPlanModel.ScheduleType)

	expectedClosure := launchPlan.Closure

	var actualClosure admin.LaunchPlanClosure
	err = proto.Unmarshal(launchPlanModel.Closure, &actualClosure)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(expectedClosure, &actualClosure))
	assert.Equal(t, launchPlanDigest, launchPlanModel.Digest)
}

func TestToLaunchPlanModelWithCronSchedule(t *testing.T) {
	lpRequest := testutils.GetLaunchPlanRequestWithCronSchedule("* * * * *")
	lpRequest.Spec.DefaultInputs = expectedInputs
	workflowID := uint(11)
	launchPlanDigest := []byte("launch plan")

	launchPlan := admin.LaunchPlan{
		Id:   lpRequest.Id,
		Spec: lpRequest.Spec,
		Closure: &admin.LaunchPlanClosure{
			ExpectedInputs:  expectedInputs,
			ExpectedOutputs: expectedOutputs,
		},
	}
	launchPlanModel, err := CreateLaunchPlanModel(launchPlan, workflowID, launchPlanDigest, admin.LaunchPlanState_INACTIVE)

	assert.NoError(t, err)
	assert.Equal(t, "project", launchPlanModel.Project)
	assert.Equal(t, "domain", launchPlanModel.Domain)
	assert.Equal(t, "name", launchPlanModel.Name)
	assert.Equal(t, "version", launchPlanModel.Version)
	assert.Equal(t, workflowID, launchPlanModel.WorkflowID)

	expectedSpec, _ := proto.Marshal(lpRequest.Spec)
	assert.Equal(t, expectedSpec, launchPlanModel.Spec)
	assert.Equal(t, models.LaunchPlanScheduleTypeCRON, launchPlanModel.ScheduleType)

	expectedClosure := launchPlan.Closure

	var actualClosure admin.LaunchPlanClosure
	err = proto.Unmarshal(launchPlanModel.Closure, &actualClosure)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(expectedClosure, &actualClosure))
	assert.Equal(t, launchPlanDigest, launchPlanModel.Digest)
}

func TestToLaunchPlanModelWithFixedRateSchedule(t *testing.T) {
	lpRequest := testutils.GetLaunchPlanRequestWithFixedRateSchedule(24, admin.FixedRateUnit_HOUR)
	lpRequest.Spec.DefaultInputs = expectedInputs
	workflowID := uint(11)
	launchPlanDigest := []byte("launch plan")

	launchPlan := admin.LaunchPlan{
		Id:   lpRequest.Id,
		Spec: lpRequest.Spec,
		Closure: &admin.LaunchPlanClosure{
			ExpectedInputs:  expectedInputs,
			ExpectedOutputs: expectedOutputs,
		},
	}
	launchPlanModel, err := CreateLaunchPlanModel(launchPlan, workflowID, launchPlanDigest, admin.LaunchPlanState_INACTIVE)

	assert.NoError(t, err)
	assert.Equal(t, "project", launchPlanModel.Project)
	assert.Equal(t, "domain", launchPlanModel.Domain)
	assert.Equal(t, "name", launchPlanModel.Name)
	assert.Equal(t, "version", launchPlanModel.Version)
	assert.Equal(t, workflowID, launchPlanModel.WorkflowID)

	expectedSpec, _ := proto.Marshal(lpRequest.Spec)
	assert.Equal(t, expectedSpec, launchPlanModel.Spec)
	assert.Equal(t, models.LaunchPlanScheduleTypeRATE, launchPlanModel.ScheduleType)

	expectedClosure := launchPlan.Closure

	var actualClosure admin.LaunchPlanClosure
	err = proto.Unmarshal(launchPlanModel.Closure, &actualClosure)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(expectedClosure, &actualClosure))
	assert.Equal(t, launchPlanDigest, launchPlanModel.Digest)
}

func TestFromLaunchPlanModel(t *testing.T) {
	lpRequest := getRequest()
	workflowRequest := testutils.GetWorkflowRequest()
	createdAt := time.Now()
	createdAtProto, _ := ptypes.TimestampProto(createdAt)
	updatedAt := createdAt.Add(time.Minute)
	updatedAtProto, _ := ptypes.TimestampProto(updatedAt)
	closure := admin.LaunchPlanClosure{
		ExpectedInputs:  lpRequest.Spec.DefaultInputs,
		ExpectedOutputs: workflowRequest.Spec.Template.Interface.Outputs,
		CreatedAt:       createdAtProto,
		UpdatedAt:       updatedAtProto,
		State:           admin.LaunchPlanState_ACTIVE,
	}
	specBytes, _ := proto.Marshal(lpRequest.Spec)
	closureBytes, _ := proto.Marshal(&closure)

	model := models.LaunchPlan{
		BaseModel: models.BaseModel{
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
		LaunchPlanKey: models.LaunchPlanKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Version: "version",
		},
		Spec:    specBytes,
		Closure: closureBytes,
		State:   &active,
	}
	lp, err := FromLaunchPlanModel(model)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(&core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      "project",
		Domain:       "domain",
		Name:         "name",
		Version:      "version",
	}, lp.Id))
	assert.True(t, proto.Equal(&closure, lp.Closure))
	assert.True(t, proto.Equal(lpRequest.Spec, lp.Spec))
}

func TestFromLaunchPlanModels(t *testing.T) {
	lpRequest := getRequest()
	workflowRequest := testutils.GetWorkflowRequest()
	state := int32(1)
	createdAt := time.Now()
	createdAtProto, _ := ptypes.TimestampProto(createdAt)
	updatedAt := createdAt.Add(time.Minute)
	updatedAtProto, _ := ptypes.TimestampProto(updatedAt)
	closure := admin.LaunchPlanClosure{
		ExpectedInputs:  lpRequest.Spec.DefaultInputs,
		ExpectedOutputs: workflowRequest.Spec.Template.Interface.Outputs,
		CreatedAt:       createdAtProto,
		UpdatedAt:       updatedAtProto,
		State:           admin.LaunchPlanState_ACTIVE,
	}
	specBytes, _ := proto.Marshal(lpRequest.Spec)
	closureBytes, _ := proto.Marshal(&closure)

	m1 := models.LaunchPlan{
		BaseModel: models.BaseModel{
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
		LaunchPlanKey: models.LaunchPlanKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Version: "version",
		},
		Spec:    specBytes,
		Closure: closureBytes,
		State:   &state,
	}
	m2 := models.LaunchPlan{
		BaseModel: models.BaseModel{
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
		LaunchPlanKey: models.LaunchPlanKey{
			Project: "project",
			Domain:  "staging",
			Name:    "othername",
			Version: "versionsecond",
		},
		Spec:    specBytes,
		Closure: closureBytes,
		State:   &state,
	}
	launchPlanModels := []models.LaunchPlan{
		m1, m2,
	}
	lp, err := FromLaunchPlanModels(launchPlanModels)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(lp))

	assert.True(t, proto.Equal(&core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      "project",
		Domain:       "staging",
		Name:         "othername",
		Version:      "versionsecond",
	}, lp[1].Id))
	assert.True(t, proto.Equal(&closure, lp[1].Closure))
	assert.True(t, proto.Equal(lpRequest.Spec, lp[1].Spec))
}
