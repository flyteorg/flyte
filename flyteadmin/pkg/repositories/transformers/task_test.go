package transformers

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

var taskDigest = []byte("task digest")

func TestCreateTask(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	task, err := CreateTaskModel(request, testutils.GetTaskClosure(), taskDigest)
	assert.NoError(t, err)
	assert.Equal(t, "project", task.Project)
	assert.Equal(t, "domain", task.Domain)
	assert.Equal(t, "name", task.Name)
	assert.Equal(t, "version", task.Version)
	assert.Equal(t, testutils.GetTaskClosureBytes(), task.Closure)
	assert.Equal(t, taskDigest, task.Digest)
	assert.Equal(t, "type", task.Type)
}

func TestFromTaskModel(t *testing.T) {
	createdAt := time.Now()
	createdAtProto := timestamppb.New(createdAt)
	taskModel := models.Task{
		BaseModel: models.BaseModel{
			CreatedAt: createdAt,
		},
		TaskKey: models.TaskKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Version: "version",
		},
		Closure: testutils.GetTaskClosureBytes(),
	}
	task, err := FromTaskModel(taskModel)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(&core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "project",
		Domain:       "domain",
		Name:         "name",
		Version:      "version",
	}, task.GetId()))
	expectedClosure := testutils.GetTaskClosure()
	expectedClosure.CreatedAt = createdAtProto
	assert.True(t, proto.Equal(expectedClosure, task.GetClosure()))
}

func TestFromTaskModels(t *testing.T) {
	createdAtA := time.Now()
	createdAtAProto := timestamppb.New(createdAtA)

	createdAtB := createdAtA.Add(time.Hour)
	createdAtBProto := timestamppb.New(createdAtB)

	taskModels := []models.Task{
		{
			BaseModel: models.BaseModel{
				CreatedAt: createdAtA,
			},
			TaskKey: models.TaskKey{
				Project: "project a",
				Domain:  "domain a",
				Name:    "name a",
				Version: "version a",
			},
			Closure: testutils.GetTaskClosureBytes(),
		},
		{
			BaseModel: models.BaseModel{
				CreatedAt: createdAtB,
			},
			TaskKey: models.TaskKey{
				Project: "project b",
				Domain:  "domain b",
				Name:    "name b",
				Version: "version b",
			},
		},
	}

	taskList, err := FromTaskModels(taskModels)
	assert.NoError(t, err)
	assert.Len(t, taskList, len(taskModels))
	assert.True(t, proto.Equal(&core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "project a",
		Domain:       "domain a",
		Name:         "name a",
		Version:      "version a",
	}, taskList[0].GetId()))
	expectedClosure := testutils.GetTaskClosure()
	expectedClosure.CreatedAt = createdAtAProto
	assert.True(t, proto.Equal(expectedClosure, taskList[0].GetClosure()))

	assert.True(t, proto.Equal(&core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "project b",
		Domain:       "domain b",
		Name:         "name b",
		Version:      "version b",
	}, taskList[1].GetId()))
	expectedClosure = &admin.TaskClosure{
		CreatedAt: createdAtBProto,
	}
	assert.True(t, proto.Equal(expectedClosure, taskList[1].GetClosure()))
}

func TestFromTaskModelsToIdentifiers(t *testing.T) {
	taskModels := []models.Task{
		{
			TaskKey: models.TaskKey{
				Project: "project a",
				Domain:  "domain a",
				Name:    "name a",
				Version: "version a",
			},
		},
		{
			TaskKey: models.TaskKey{
				Project: "project b",
				Domain:  "domain b",
				Name:    "name b",
				Version: "version b",
			},
		},
	}

	taskIds := FromTaskModelsToIdentifiers(taskModels)
	assert.Equal(t, "domain a", taskIds[0].GetDomain())
	assert.Equal(t, "project a", taskIds[0].GetProject())
	assert.Equal(t, "name a", taskIds[0].GetName())
	assert.Equal(t, "domain b", taskIds[1].GetDomain())
	assert.Equal(t, "project b", taskIds[1].GetProject())
	assert.Equal(t, "name b", taskIds[1].GetName())
}
