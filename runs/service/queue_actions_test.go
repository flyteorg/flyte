package service

import (
	"connectrpc.com/connect"
	"context"
	common "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	workflow "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestListActionsRejectsUnsupportedAndInvalidQueueScopes(t *testing.T) {
	queue := &common.QueueIdentifier{Name: "queue"}
	run := &common.RunIdentifier{Project: "project", Domain: "development", Name: "run"}
	for _, tt := range []struct {
		name    string
		request *workflow.ListActionsRequest
		code    connect.Code
	}{
		{"queue scope", &workflow.ListActionsRequest{QueueId: queue}, connect.CodeUnimplemented},
		{"missing scope", &workflow.ListActionsRequest{}, connect.CodeInvalidArgument},
		{"both scopes", &workflow.ListActionsRequest{RunId: run, QueueId: queue}, connect.CodeInvalidArgument},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := (&RunService{}).ListActions(context.Background(), connect.NewRequest(tt.request))
			assert.Equal(t, tt.code, connect.CodeOf(err))
		})
	}
}
