package ext

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flyte/flytectl/pkg/filters"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	workflowListResponse    *admin.WorkflowList
	namedEntityListResponse *admin.NamedEntityList
	workflowFilter          = filters.Filters{}
	workflowResponse        *admin.Workflow
)

func getWorkflowFetcherSetup() {
	ctx = context.Background()
	adminClient = new(mocks.AdminServiceClient)
	adminFetcherExt = AdminFetcherExtClient{AdminClient: adminClient}

	sortedListLiteralType := core.Variable{
		Type: &core.LiteralType{
			Type: &core.LiteralType_CollectionType{
				CollectionType: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_INTEGER,
					},
				},
			},
		},
	}
	variableMap := map[string]*core.Variable{
		"sorted_list1": &sortedListLiteralType,
		"sorted_list2": &sortedListLiteralType,
	}

	var compiledTasks []*core.CompiledTask
	compiledTasks = append(compiledTasks, &core.CompiledTask{
		Template: &core.TaskTemplate{
			Interface: &core.TypedInterface{
				Inputs: &core.VariableMap{
					Variables: variableMap,
				},
			},
		},
	})

	workflow1 := &admin.Workflow{
		Id: &core.Identifier{
			Name:    "task1",
			Version: "v1",
		},
		Closure: &admin.WorkflowClosure{
			CreatedAt: &timestamppb.Timestamp{Seconds: 1, Nanos: 0},
			CompiledWorkflow: &core.CompiledWorkflowClosure{
				Tasks: compiledTasks,
			},
		},
	}
	workflow2 := &admin.Workflow{
		Id: &core.Identifier{
			Name:    "workflow",
			Version: "v2",
		},
		Closure: &admin.WorkflowClosure{
			CreatedAt: &timestamppb.Timestamp{Seconds: 1, Nanos: 0},
			CompiledWorkflow: &core.CompiledWorkflowClosure{
				Tasks: compiledTasks,
			},
		},
	}

	namedEntity := &admin.NamedEntity{
		Id: &admin.NamedEntityIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "workflow",
		},
		ResourceType: core.ResourceType_WORKFLOW,
	}

	workflows := []*admin.Workflow{workflow2, workflow1}

	namedEntityListResponse = &admin.NamedEntityList{
		Entities: []*admin.NamedEntity{namedEntity},
	}
	workflowListResponse = &admin.WorkflowList{
		Workflows: workflows,
	}
	workflowResponse = workflows[0]
}

func TestFetchAllWorkflows(t *testing.T) {
	t.Run("non empty response", func(t *testing.T) {
		getWorkflowFetcherSetup()
		adminClient.OnListNamedEntitiesMatch(mock.Anything, mock.Anything).Return(namedEntityListResponse, nil)
		_, err := adminFetcherExt.FetchAllWorkflows(ctx, "project", "domain", workflowFilter)
		assert.Nil(t, err)
	})
	t.Run("empty response", func(t *testing.T) {
		getWorkflowFetcherSetup()
		namedEntityListResponse := &admin.NamedEntityList{}
		adminClient.OnListNamedEntitiesMatch(mock.Anything, mock.Anything).Return(namedEntityListResponse, nil)
		_, err := adminFetcherExt.FetchAllWorkflows(ctx, "project", "domain", workflowFilter)
		assert.Equal(t, fmt.Errorf("no workflow retrieved for project project domain domain"), err)
	})
}

func TestFetchAllWorkflowsError(t *testing.T) {
	getWorkflowFetcherSetup()
	adminClient.OnListNamedEntitiesMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	_, err := adminFetcherExt.FetchAllWorkflows(ctx, "project", "domain", workflowFilter)
	assert.Equal(t, fmt.Errorf("failed"), err)
}

func TestFetchAllVerOfWorkflow(t *testing.T) {
	getWorkflowFetcherSetup()
	adminClient.OnListWorkflowsMatch(mock.Anything, mock.Anything).Return(workflowListResponse, nil)
	_, err := adminFetcherExt.FetchAllVerOfWorkflow(ctx, "workflowName", "project", "domain", workflowFilter)
	assert.Nil(t, err)
}

func TestFetchAllVerOfWorkflowError(t *testing.T) {
	getWorkflowFetcherSetup()
	adminClient.OnListWorkflowsMatch(mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed"))
	_, err := adminFetcherExt.FetchAllVerOfWorkflow(ctx, "workflowName", "project", "domain", workflowFilter)
	assert.Equal(t, fmt.Errorf("failed"), err)
}

func TestFetchAllVerOfWorkflowEmptyResponse(t *testing.T) {
	workflowListResponse := &admin.WorkflowList{}
	getWorkflowFetcherSetup()
	adminClient.OnListWorkflowsMatch(mock.Anything, mock.Anything).Return(workflowListResponse, nil)
	_, err := adminFetcherExt.FetchAllVerOfWorkflow(ctx, "workflowName", "project", "domain", workflowFilter)
	assert.Equal(t, fmt.Errorf("no workflow retrieved for workflowName"), err)
}

func TestFetchWorkflowLatestVersion(t *testing.T) {
	getWorkflowFetcherSetup()
	adminClient.OnGetWorkflowMatch(mock.Anything, mock.Anything).Return(workflowResponse, nil)
	adminClient.OnListWorkflowsMatch(mock.Anything, mock.Anything).Return(workflowListResponse, nil)
	_, err := adminFetcherExt.FetchWorkflowLatestVersion(ctx, "workflowName", "project", "domain", workflowFilter)
	assert.Nil(t, err)
}

func TestFetchWorkflowLatestVersionError(t *testing.T) {
	workflowListResponse := &admin.WorkflowList{}
	getWorkflowFetcherSetup()
	adminClient.OnListWorkflowsMatch(mock.Anything, mock.Anything).Return(workflowListResponse, nil)
	_, err := adminFetcherExt.FetchWorkflowLatestVersion(ctx, "workflowName", "project", "domain", workflowFilter)
	assert.Equal(t, fmt.Errorf("no workflow retrieved for workflowName"), err)
}
