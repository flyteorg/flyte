package qubole

import (
	"encoding/json"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	tasksMocks "github.com/lyft/flyteplugins/go/tasks/v1/types/mocks"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func getMockTaskContext() *tasksMocks.TaskContext {
	taskCtx := &tasksMocks.TaskContext{}

	id := &tasksMocks.TaskExecutionID{}
	id.On("GetGeneratedName").Return("flyteplugins_integration")
	id.On("GetID").Return(core.TaskExecutionIdentifier{
		RetryAttempt: 0,
	})
	taskCtx.On("GetTaskExecutionID").Return(id)

	return taskCtx
}

func TestConstructEventInfoFromQuboleWorkItems(t *testing.T) {
	workItems := map[string]interface{}{
		"key_1": QuboleWorkItem{
			CommandId:          "12345",
			UniqueWorkCacheKey: "key_1",
			Retries:            0,
			Status:             QuboleWorkSucceeded,
			ClusterLabel:       "default",
			Tags:               []string{},
		},
	}

	out, err := constructEventInfoFromQuboleWorkItems(getMockTaskContext(), workItems)
	assert.NoError(t, err)
	assert.Equal(t, "Retry: 0 Status: Succeeded [12345]", out.Logs[0].Name)
	assert.Equal(t, "12345", out.CustomInfo.Fields["key_1"].GetStructValue().Fields["command_id"].GetStringValue())
	status := out.CustomInfo.Fields["key_1"].GetStructValue().Fields["status"]
	assert.Equal(t, float64(7), status.GetNumberValue())
	assert.True(t, strings.Contains(out.Logs[0].Uri, "api.qubole"))
}

func TestPrinting(t *testing.T) {
	assert.Equal(t, "NotStarted", QuboleWorkNotStarted.String())
	assert.Equal(t, "Unknown", QuboleWorkUnknown.String())
	assert.Equal(t, "Runnable", QuboleWorkRunnable.String())
	assert.Equal(t, "Running", QuboleWorkRunning.String())
	assert.Equal(t, "ExecutionFailed", QuboleWorkExecutionFailed.String())
	assert.Equal(t, "ExecutionSucceeded", QuboleWorkExecutionSucceeded.String())
	assert.Equal(t, "Failed", QuboleWorkFailed.String())
	assert.Equal(t, "Succeeded", QuboleWorkSucceeded.String())
}

func TestEquality(t *testing.T) {
	first := QuboleWorkItem{
		Tags:               []string{"hello"},
		Retries:            0,
		CommandId:          "123456789",
		Status:             QuboleWorkRunnable,
		UniqueWorkCacheKey: "fdsfa",
		ClusterLabel:       "default",
	}
	second := QuboleWorkItem{
		Tags:               []string{"hello"},
		Retries:            0,
		CommandId:          "123456789",
		Status:             QuboleWorkRunnable,
		UniqueWorkCacheKey: "fdsfa",
		ClusterLabel:       "default",
	}
	third := QuboleWorkItem{
		Tags:               []string{"hello"},
		Retries:            0,
		CommandId:          "123456789",
		Status:             QuboleWorkFailed,
		UniqueWorkCacheKey: "fdsfa",
		ClusterLabel:       "default",
	}
	fourth := QuboleWorkItem{
		Tags:               []string{"hello"},
		Retries:            0,
		CommandId:          "1234789",
		Status:             QuboleWorkRunnable,
		UniqueWorkCacheKey: "fdsfad",
		ClusterLabel:       "default",
	}

	assert.True(t, first.EqualTo(second))
	assert.False(t, first.EqualTo(third))
	assert.False(t, first.EqualTo(fourth))
}

func TestMapEquality(t *testing.T) {
	first := QuboleWorkItem{
		Tags:               []string{"hello"},
		Retries:            0,
		CommandId:          "123456789",
		Status:             QuboleWorkRunnable,
		UniqueWorkCacheKey: "fdsfa",
		ClusterLabel:       "default",
	}
	second := QuboleWorkItem{
		Tags:               []string{"hello"},
		Retries:            0,
		CommandId:          "123456789",
		Status:             QuboleWorkRunnable,
		UniqueWorkCacheKey: "fdsfa",
		ClusterLabel:       "default",
	}
	third := QuboleWorkItem{
		Tags:               []string{"hello"},
		Retries:            0,
		CommandId:          "123456789",
		Status:             QuboleWorkFailed,
		UniqueWorkCacheKey: "fdsfa",
		ClusterLabel:       "default",
	}

	old := map[string]QuboleWorkItem{
		"first":  first,
		"second": second,
	}

	old2 := map[string]QuboleWorkItem{
		"first":  first,
		"second": second,
		"third":  third,
	}

	new1 := map[string]interface{}{
		"first":  first,
		"second": second,
		"third":  third,
	}

	new2 := map[string]interface{}{
		"first":  first,
		"second": second,
	}

	new3 := map[string]interface{}{
		"first":  first,
		"second": third,
	}

	assert.False(t, workItemMapsAreEqual(old, new1))
	assert.False(t, workItemMapsAreEqual(old2, new2))
	assert.True(t, workItemMapsAreEqual(old, new2))
	assert.False(t, workItemMapsAreEqual(old, new3))
}

func TestInterfaceConverter(t *testing.T) {
	// This is a complicated step to reproduce what will ultimately be given to the function at runtime, the values
	// inside the CustomState
	item := QuboleWorkItem{
		Status: QuboleWorkRunning,
		CommandId: "123456",
		Query: "",
		UniqueWorkCacheKey: "fjdsakfjd",
	}
	raw, err := json.Marshal(map[string]interface{}{"":item})
	assert.NoError(t, err)

	// We can't unmarshal into a interface{} but we can unmarhsal into a interface{} if it's the value of a map.
	interfaceItem := map[string]interface{}{}
	err = json.Unmarshal(raw, &interfaceItem)
	assert.NoError(t, err)

	convertedItem, err := InterfaceConverter(interfaceItem[""])
	assert.NoError(t, err)
	assert.Equal(t, "123456", convertedItem.CommandId)
	assert.Equal(t, QuboleWorkRunning, convertedItem.Status)
	assert.Equal(t, "fjdsakfjd", convertedItem.UniqueWorkCacheKey)
}
