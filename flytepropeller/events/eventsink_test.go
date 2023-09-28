package events

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
)

func TestFileEvent(t *testing.T) {
	now := ptypes.TimestampNow()
	dir, err := ioutil.TempDir("", "eventstest")
	if err != nil {
		assert.FailNow(t, "test dir creation failed")
	}
	defer func() { assert.NoError(t, os.RemoveAll(dir)) }()

	file := path.Join(dir, "events.test")
	sink, err := NewFileSink(file)
	if err != nil {
		assert.FailNow(t, "failed to create file sync "+err.Error())
	}

	executionID := &core.WorkflowExecutionIdentifier{
		Project: "FlyteTest",
		Domain:  "FlyteStaging",
		Name:    "Name",
	}

	workflowEvent := &event.WorkflowExecutionEvent{
		ExecutionId: executionID,
		Phase:       core.WorkflowExecution_SUCCEEDED,
		OccurredAt:  now,
	}
	err = sink.Sink(context.Background(), workflowEvent)
	assert.NoError(t, err)

	nodeEvent := &event.NodeExecutionEvent{
		Id: &core.NodeExecutionIdentifier{
			NodeId: "node1",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "FlyteTest",
				Domain:  "FlyteStaging",
				Name:    "Name",
			},
		},
		Phase:      core.NodeExecution_RUNNING,
		OccurredAt: now,
	}
	err = sink.Sink(context.Background(), nodeEvent)
	assert.NoError(t, err)

	taskEvent := &event.TaskExecutionEvent{
		TaskId: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      executionID.Project,
			Domain:       executionID.Domain,
			Name:         executionID.Name,
		},
		ParentNodeExecutionId: nodeEvent.Id,
		Phase:                 core.TaskExecution_FAILED,
		OccurredAt:            now,
	}
	assert.NoError(t, err)
	err = sink.Sink(context.Background(), taskEvent)
	assert.NoError(t, err)

	expected := []string{
		"[--WF EVENT--] project:\"FlyteTest\" domain:\"FlyteStaging\" name:\"Name\" , " +
			"Phase: SUCCEEDED, OccuredAt: " + ptypes.TimestampString(now),
		"[--NODE EVENT--] node_id:\"node1\" execution_id:<project:\"FlyteTest\" " +
			"domain:\"FlyteStaging\" name:\"Name\" > , Phase: RUNNING, OccuredAt: " + ptypes.TimestampString(now),
		"[--TASK EVENT--] resource_type:TASK project:\"FlyteTest\" domain:\"FlyteStaging\" " +
			"name:\"Name\" ,node_id:\"node1\" execution_id:<project:\"FlyteTest\" domain:\"FlyteStaging\" " +
			"name:\"Name\" > , Phase: FAILED, OccuredAt: " + ptypes.TimestampString(now),
	}

	actual, err := readLinesFromFile(file)
	if err != nil {
		assert.FailNow(t, "failed to read file "+err.Error())
	}

	assert.True(t, reflect.DeepEqual(expected, actual), "Expected %v\nActual %v", expected, actual)
}

func readLinesFromFile(name string) ([]string, error) {
	/* #nosec */
	raw, err := ioutil.ReadFile(name)

	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(raw), "\n")

	// remove the last element because is an empty element from split
	return lines[0 : len(lines)-1], nil
}
