package service

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"

	common "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	workflow "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

func TestListActionsRejectsUnsupportedAndInvalidQueueScopes(t *testing.T) {
	queue := &common.QueueIdentifier{Name: "queue"}
	for _, tt := range []struct {
		name    string
		request *workflow.ListActionsRequest
		code    connect.Code
	}{
		{"queue scope", &workflow.ListActionsRequest{ScopeBy: &workflow.ListActionsRequest_QueueId{QueueId: queue}}, connect.CodeUnimplemented},
		{"missing scope", &workflow.ListActionsRequest{}, connect.CodeInvalidArgument},
		{"nil run scope", &workflow.ListActionsRequest{ScopeBy: &workflow.ListActionsRequest_RunId{}}, connect.CodeInvalidArgument},
		{"nil queue scope", &workflow.ListActionsRequest{ScopeBy: &workflow.ListActionsRequest_QueueId{}}, connect.CodeInvalidArgument},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := (&RunService{}).ListActions(context.Background(), connect.NewRequest(tt.request))
			assert.Equal(t, tt.code, connect.CodeOf(err))
		})
	}
}

func TestListActionsRunScopePreservesExistingWireFormat(t *testing.T) {
	run := &common.RunIdentifier{Org: "org", Project: "project", Domain: "development", Name: "run"}
	payload, err := proto.Marshal(run)
	require.NoError(t, err)
	// Before scope_by existed, run_id was the singular message at field 2.
	oldWire := protowire.AppendTag(nil, 2, protowire.BytesType)
	oldWire = protowire.AppendBytes(oldWire, payload)
	request := &workflow.ListActionsRequest{}
	require.NoError(t, proto.Unmarshal(oldWire, request))
	require.IsType(t, &workflow.ListActionsRequest_RunId{}, request.GetScopeBy())
	assert.True(t, proto.Equal(run, request.GetRunId()))
	newWire, err := proto.Marshal(request)
	require.NoError(t, err)
	assert.Equal(t, oldWire, newWire)

	// JSON stays flat: the oneof does not add a scopeBy envelope.
	jsonWire, err := protojson.Marshal(request)
	require.NoError(t, err)
	assert.Contains(t, string(jsonWire), `"runId"`)
	assert.NotContains(t, string(jsonWire), `"scopeBy"`)
}

func TestListActionsJSONRejectsConflictingScopes(t *testing.T) {
	request := &workflow.ListActionsRequest{}
	err := protojson.Unmarshal([]byte(`{"runId":{"name":"run"},"queueId":{"name":"queue"}}`), request)
	require.Error(t, err)
}
