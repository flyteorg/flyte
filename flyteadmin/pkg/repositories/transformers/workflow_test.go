package transformers

import (
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

const remoteClosureIdentifier = "remote closure id"

var workflowDigest = []byte("workflow digest")

func TestCreateWorkflow(t *testing.T) {
	request := testutils.GetWorkflowRequest()
	workflow, err := CreateWorkflowModel(request, remoteClosureIdentifier, workflowDigest)
	assert.NoError(t, err)
	assert.Equal(t, "project", workflow.Project)
	assert.Equal(t, "domain", workflow.Domain)
	assert.Equal(t, "name", workflow.Name)
	assert.Equal(t, "version", workflow.Version)
	expectedTypedInterface := testutils.GetWorkflowRequestInterfaceBytes()
	assert.Equal(t, expectedTypedInterface, workflow.TypedInterface)
	assert.Equal(t, remoteClosureIdentifier, workflow.RemoteClosureIdentifier)
	assert.Equal(t, workflowDigest, workflow.Digest)
}

func TestCreateWorkflowEmptyInterface(t *testing.T) {
	request := testutils.GetWorkflowRequest()
	request.Spec.Template.Interface = nil
	workflow, err := CreateWorkflowModel(request, remoteClosureIdentifier, workflowDigest)
	assert.NoError(t, err)
	assert.Equal(t, "project", workflow.Project)
	assert.Equal(t, "domain", workflow.Domain)
	assert.Equal(t, "name", workflow.Name)
	assert.Equal(t, "version", workflow.Version)
	assert.Empty(t, workflow.TypedInterface)
	assert.Equal(t, remoteClosureIdentifier, workflow.RemoteClosureIdentifier)
	assert.Equal(t, workflowDigest, workflow.Digest)
}

func TestFromWorkflowModel(t *testing.T) {
	createdAt := time.Now()
	createdAtProto, _ := ptypes.TimestampProto(createdAt)
	workflowModel := models.Workflow{
		BaseModel: models.BaseModel{
			CreatedAt: createdAt,
		},
		WorkflowKey: models.WorkflowKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Version: "version",
		},
		TypedInterface:          testutils.GetWorkflowRequestInterfaceBytes(),
		RemoteClosureIdentifier: remoteClosureIdentifier,
	}
	workflow, err := FromWorkflowModel(workflowModel)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(&core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "project",
		Domain:       "domain",
		Name:         "name",
		Version:      "version",
	}, workflow.GetId()))

	var workflowInterface core.TypedInterface
	err = proto.Unmarshal(workflowModel.TypedInterface, &workflowInterface)
	assert.NoError(t, err)

	assert.True(t, proto.Equal(&admin.WorkflowClosure{
		CreatedAt: createdAtProto,
		CompiledWorkflow: &core.CompiledWorkflowClosure{
			Primary: &core.CompiledWorkflow{
				Template: &core.WorkflowTemplate{
					Interface: &workflowInterface,
				},
			},
		},
	}, workflow.GetClosure()))
}

func TestFromWorkflowModels(t *testing.T) {
	createdAtA := time.Now()
	createdAtAProto, _ := ptypes.TimestampProto(createdAtA)

	createdAtB := createdAtA.Add(time.Hour)
	createdAtBProto, _ := ptypes.TimestampProto(createdAtB)

	workflowModels := []models.Workflow{
		{
			BaseModel: models.BaseModel{
				CreatedAt: createdAtA,
			},
			WorkflowKey: models.WorkflowKey{
				Project: "project a",
				Domain:  "domain a",
				Name:    "name a",
				Version: "version a",
			},
			TypedInterface:          testutils.GetWorkflowRequestInterfaceBytes(),
			RemoteClosureIdentifier: remoteClosureIdentifier,
		},
		{
			BaseModel: models.BaseModel{
				CreatedAt: createdAtB,
			},
			WorkflowKey: models.WorkflowKey{
				Project: "project b",
				Domain:  "domain b",
				Name:    "name b",
				Version: "version b",
			},
		},
	}

	workflowList, err := FromWorkflowModels(workflowModels)
	assert.NoError(t, err)
	assert.Len(t, workflowList, len(workflowModels))
	assert.True(t, proto.Equal(&core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "project a",
		Domain:       "domain a",
		Name:         "name a",
		Version:      "version a",
	}, workflowList[0].GetId()))

	workflowInterface := testutils.GetWorkflowRequestInterface()
	assert.NoError(t, err)

	assert.True(t, proto.Equal(&admin.WorkflowClosure{
		CreatedAt: createdAtAProto,
		CompiledWorkflow: &core.CompiledWorkflowClosure{
			Primary: &core.CompiledWorkflow{
				Template: &core.WorkflowTemplate{
					Interface: workflowInterface,
				},
			},
		},
	}, workflowList[0].GetClosure()))

	assert.True(t, proto.Equal(&core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "project b",
		Domain:       "domain b",
		Name:         "name b",
		Version:      "version b",
	}, workflowList[1].GetId()))

	assert.True(t, proto.Equal(&admin.WorkflowClosure{
		CreatedAt: createdAtBProto,
		CompiledWorkflow: &core.CompiledWorkflowClosure{
			Primary: &core.CompiledWorkflow{
				Template: &core.WorkflowTemplate{
					Interface: &core.TypedInterface{},
				},
			},
		},
	}, workflowList[1].GetClosure()))
}
