package transformers

import (
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/golang/protobuf/ptypes"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

var taskDigest = []byte("task digest")

func TestCreateTask(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	task, err := CreateTaskModel(request, *testutils.GetTaskClosure(), taskDigest)
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
	createdAtProto, _ := ptypes.TimestampProto(createdAt)
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
	}, task.Id))
	expectedClosure := testutils.GetTaskClosure()
	expectedClosure.CreatedAt = createdAtProto
	assert.True(t, proto.Equal(expectedClosure, task.Closure))
}

func TestFromTaskModels(t *testing.T) {
	createdAtA := time.Now()
	createdAtAProto, _ := ptypes.TimestampProto(createdAtA)

	createdAtB := createdAtA.Add(time.Hour)
	createdAtBProto, _ := ptypes.TimestampProto(createdAtB)

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
	}, taskList[0].Id))
	expectedClosure := testutils.GetTaskClosure()
	expectedClosure.CreatedAt = createdAtAProto
	assert.True(t, proto.Equal(expectedClosure, taskList[0].Closure))

	assert.True(t, proto.Equal(&core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "project b",
		Domain:       "domain b",
		Name:         "name b",
		Version:      "version b",
	}, taskList[1].Id))
	expectedClosure = &admin.TaskClosure{
		CreatedAt: createdAtBProto,
	}
	assert.True(t, proto.Equal(expectedClosure, taskList[1].Closure))
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
	assert.Equal(t, "domain a", taskIds[0].Domain)
	assert.Equal(t, "project a", taskIds[0].Project)
	assert.Equal(t, "name a", taskIds[0].Name)
	assert.Equal(t, "domain b", taskIds[1].Domain)
	assert.Equal(t, "project b", taskIds[1].Project)
	assert.Equal(t, "name b", taskIds[1].Name)
}
