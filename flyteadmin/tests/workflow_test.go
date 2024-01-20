//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
)

func TestCreateWorkflow(t *testing.T) {
	truncateAllTablesForTestingOnly()
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	taskCreateReq := testutils.GetValidTaskRequest()
	taskCreateReq.Id.Project = "admintests"
	taskCreateReq.Id.Domain = "development"
	taskCreateReq.Id.Name = "simple task"
	_, err := client.CreateTask(ctx, &taskCreateReq)
	assert.NoError(t, err)

	identifier := core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "admintests",
		Domain:       "development",
		Name:         "name",
		Version:      "version",
	}
	req := admin.WorkflowCreateRequest{
		Id: &identifier,
		Spec: &admin.WorkflowSpec{
			Template: &core.WorkflowTemplate{
				Id:        &identifier,
				Interface: &core.TypedInterface{},
				Nodes: []*core.Node{
					{
						Id: "I'm a node",
						Target: &core.Node_TaskNode{
							TaskNode: &core.TaskNode{
								Reference: &core.TaskNode_ReferenceId{
									ReferenceId: taskCreateReq.Id,
								},
							},
						},
					},
				},
			},
		},
	}

	_, err = client.CreateWorkflow(ctx, &req)
	assert.Nil(t, err)
}

func TestGetWorkflows(t *testing.T) {
	truncateAllTablesForTestingOnly()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	insertTasksForTests(t, client)
	insertWorkflowsForTests(t, client)

	t.Run("TestGetWorkflowGrpc", testGetWorkflowGrpc)
	t.Run("TestGetWorkflowHTTP", testGetWorkflowHTTP)
	t.Run("TestListWorkflowGrpc", testListWorkflowGrpc)
	t.Run("TestListWorkflowHTTP", testListWorkflowHTTP)
	t.Run("TestListWorkflow_PaginationGrpc", testListWorkflow_PaginationGrpc)
	t.Run("TestListWorkflow_PaginationHTTP", testListWorkflow_PaginationHTTP)
	t.Run("TestListWorkflow_FiltersGrpc", testListWorkflow_FiltersGrpc)
	t.Run("TestListWorkflow_FiltersHTTP", testListWorkflow_FiltersHTTP)
}

func TestGetDynamicNodeWorkflow(t *testing.T) {
	ctx := context.Background()
	truncateAllTablesForTestingOnly()
	populateWorkflowExecutionForTestingOnly(project, domain, name)
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	occurredAt := time.Now()
	occurredAtProto := timestamppb.New(occurredAt)

	createTaskAndNodeExecution(ctx, t, client, conn, occurredAtProto)

	_, err := client.CreateTaskEvent(ctx, &admin.TaskExecutionEventRequest{
		RequestId: "request id",
		Event: &event.TaskExecutionEvent{
			TaskId:                taskIdentifier,
			ParentNodeExecutionId: nodeExecutionId,
			Phase:                 core.TaskExecution_RUNNING,
			RetryAttempt:          1,
			OccurredAt:            occurredAtProto,
		},
	})
	require.NoError(t, err)

	dynamicWfId := core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "admintests",
		Domain:       "development",
		Name:         "name",
		Version:      "version",
	}
	dynamicWf := &core.CompiledWorkflowClosure{
		Primary: &core.CompiledWorkflow{
			Template: &core.WorkflowTemplate{
				Id:        &dynamicWfId,
				Interface: &core.TypedInterface{},
				Nodes: []*core.Node{
					{
						Id: "I'm a node",
						Target: &core.Node_TaskNode{
							TaskNode: &core.TaskNode{
								Reference: &core.TaskNode_ReferenceId{
									ReferenceId: taskIdentifier,
								},
							},
						},
					},
				},
			},
		},
	}
	childOccurredAt := occurredAt.Add(time.Minute)
	childOccurredAtProto := timestamppb.New(childOccurredAt)
	childNodeExecutionID := &core.NodeExecutionIdentifier{
		NodeId: "child_node",
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: project,
			Domain:  domain,
			Name:    name,
		},
	}
	_, err = client.CreateNodeEvent(ctx, &admin.NodeExecutionEventRequest{
		RequestId: "request id",
		Event: &event.NodeExecutionEvent{
			Id:    childNodeExecutionID,
			Phase: core.NodeExecution_RUNNING,
			InputValue: &event.NodeExecutionEvent_InputUri{
				InputUri: inputURI,
			},
			OccurredAt: childOccurredAtProto,
			ParentTaskMetadata: &event.ParentTaskExecutionMetadata{
				Id: taskExecutionIdentifier,
			},
			IsDynamic: true,
			IsParent:  true,
			TargetMetadata: &event.NodeExecutionEvent_TaskNodeMetadata{
				TaskNodeMetadata: &event.TaskNodeMetadata{
					DynamicWorkflow: &event.DynamicWorkflowNodeMetadata{
						Id:                &dynamicWfId,
						CompiledWorkflow:  dynamicWf,
						DynamicJobSpecUri: "s3://bla-bla",
					},
				},
			},
		},
	})
	require.NoError(t, err)

	t.Run("TestGetDynamicNodeWorkflowGrpc", func(t *testing.T) {
		resp, err := client.GetDynamicNodeWorkflow(ctx, &admin.GetDynamicNodeWorkflowRequest{
			Id: childNodeExecutionID,
		})

		assert.NoError(t, err)
		assert.True(t, proto.Equal(dynamicWf, resp.GetCompiledWorkflow()))
	})

	t.Run("TestGetDynamicNodeWorkflowHttp", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/v1/dynamic_node_workflow/project/domain/execution%%20name/child_node", GetTestHostEndpoint())
		getRequest, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)
		addHTTPRequestHeaders(getRequest)

		httpClient := &http.Client{Timeout: 2 * time.Second}
		resp, err := httpClient.Do(getRequest)
		require.NoError(t, err)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode, "unexpected resp: %s", string(body))
		wfResp := &admin.DynamicNodeWorkflowResponse{}
		require.NoError(t, proto.Unmarshal(body, wfResp))
		assert.True(t, proto.Equal(dynamicWf, wfResp.GetCompiledWorkflow()))
	})
}

func testGetWorkflowGrpc(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	workflow, err := client.GetWorkflow(ctx, &admin.ObjectGetRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_WORKFLOW,
			Project:      "admintests",
			Domain:       "development",
			Name:         "name_a",
			Version:      "123",
		},
	})
	assert.NoError(t, err)
	assert.True(t, proto.Equal(&core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "admintests",
		Domain:       "development",
		Name:         "name_a",
		Version:      "123",
	}, workflow.Id))
}

func testGetWorkflowHTTP(t *testing.T) {
	url := fmt.Sprintf("%s/api/v1/workflows/admintests/development/name_a/123", GetTestHostEndpoint())
	getRequest, err := http.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	addHTTPRequestHeaders(getRequest)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(getRequest)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	octetStreamedWorkflow := admin.Workflow{}
	proto.Unmarshal(body, &octetStreamedWorkflow)
	assert.Equal(t, "admintests", octetStreamedWorkflow.Id.GetProject())
	assert.Equal(t, "development", octetStreamedWorkflow.Id.GetDomain())
	assert.Equal(t, "name_a", octetStreamedWorkflow.Id.GetName())
	assert.Equal(t, "123", octetStreamedWorkflow.Id.Version)
}

func testListWorkflowGrpc(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	workflows, err := client.ListWorkflows(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
			Name:    "name_a",
		},
		Limit: 20,
	})
	assert.NoError(t, err)
	assert.Len(t, workflows.Workflows, 3)

	for _, workflow := range workflows.Workflows {
		assert.Equal(t, "admintests", workflow.Id.Project)
		assert.Equal(t, "development", workflow.Id.Domain)
		assert.Equal(t, "name_a", workflow.Id.Name)
		assert.Contains(t, entityVersions, workflow.Id.Version)
	}
}

func testListWorkflowHTTP(t *testing.T) {
	url := fmt.Sprintf("%s/api/v1/workflows/admintests/development/name_a?limit=20", GetTestHostEndpoint())
	getRequest, err := http.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	addHTTPRequestHeaders(getRequest)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(getRequest)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	octetStreamedWorkflowList := admin.WorkflowList{}
	proto.Unmarshal(body, &octetStreamedWorkflowList)
	assert.Len(t, octetStreamedWorkflowList.Workflows, 3)

	for _, workflow := range octetStreamedWorkflowList.Workflows {
		assert.Equal(t, "admintests", workflow.Id.Project)
		assert.Equal(t, "development", workflow.Id.Domain)
		assert.Equal(t, "name_a", workflow.Id.Name)
		assert.Contains(t, entityVersions, workflow.Id.Version)
	}
}

func testListWorkflow_PaginationGrpc(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	workflows, err := client.ListWorkflows(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
			Name:    "name_a",
		},
		Limit: 2,
	})
	assert.NoError(t, err)
	assert.Len(t, workflows.Workflows, 2)

	firstResponseVersions := make([]string, 2)
	for idx, workflow := range workflows.Workflows {
		assert.Equal(t, "admintests", workflow.Id.Project)
		assert.Equal(t, "development", workflow.Id.Domain)
		assert.Equal(t, "name_a", workflow.Id.Name)
		assert.Contains(t, entityVersions, workflow.Id.Version)

		firstResponseVersions[idx] = workflow.Id.Version
	}

	workflows, err = client.ListWorkflows(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
			Name:    "name_a",
		},
		Limit: 2,
		Token: "2",
	})
	assert.NoError(t, err)
	assert.Len(t, workflows.Workflows, 1)

	for _, workflow := range workflows.Workflows {
		assert.Equal(t, "admintests", workflow.Id.Project)
		assert.Equal(t, "development", workflow.Id.Domain)
		assert.Equal(t, "name_a", workflow.Id.Name)
		assert.Contains(t, entityVersions, workflow.Id.Version)
		assert.NotContains(t, firstResponseVersions, workflow.Id.Version)
	}
	assert.Empty(t, workflows.Token)
}

func testListWorkflow_PaginationHTTP(t *testing.T) {
	url := fmt.Sprintf("%s/api/v1/workflows/admintests/development/name_a?limit=2", GetTestHostEndpoint())
	getRequest, err := http.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	addHTTPRequestHeaders(getRequest)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(getRequest)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	octetStreamedWorkflowList := admin.WorkflowList{}
	proto.Unmarshal(body, &octetStreamedWorkflowList)
	assert.Len(t, octetStreamedWorkflowList.Workflows, 2)

	firstResponseVersions := make([]string, 2)
	for idx, workflow := range octetStreamedWorkflowList.Workflows {
		assert.Equal(t, "admintests", workflow.Id.Project)
		assert.Equal(t, "development", workflow.Id.Domain)
		assert.Equal(t, "name_a", workflow.Id.Name)
		assert.Contains(t, entityVersions, workflow.Id.Version)

		firstResponseVersions[idx] = workflow.Id.Version
	}

	url = fmt.Sprintf("%s/api/v1/workflows/admintests/development/name_a?limit=2&token=%s",
		GetTestHostEndpoint(), octetStreamedWorkflowList.Token)
	getRequest, err = http.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	addHTTPRequestHeaders(getRequest)

	resp, err = httpClient.Do(getRequest)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ = ioutil.ReadAll(resp.Body)
	octetStreamedWorkflowList = admin.WorkflowList{}
	proto.Unmarshal(body, &octetStreamedWorkflowList)
	assert.Len(t, octetStreamedWorkflowList.Workflows, 1)

	for _, workflow := range octetStreamedWorkflowList.Workflows {
		assert.Equal(t, "admintests", workflow.Id.Project)
		assert.Equal(t, "development", workflow.Id.Domain)
		assert.Equal(t, "name_a", workflow.Id.Name)
		assert.Contains(t, entityVersions, workflow.Id.Version)
		assert.NotContains(t, firstResponseVersions, workflow.Id.Version)
	}
	assert.Empty(t, octetStreamedWorkflowList.Token)
}

func testListWorkflow_FiltersGrpc(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	workflows, err := client.ListWorkflows(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
			Name:    "name_a",
		},
		Limit:   20,
		Filters: "eq(version,123)",
	})
	assert.NoError(t, err)
	assert.Len(t, workflows.Workflows, 1)

	workflow := workflows.Workflows[0]
	assert.True(t, proto.Equal(&core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "admintests",
		Domain:       "development",
		Name:         "name_a",
		Version:      "123",
	}, workflow.Id))
}

func testListWorkflow_FiltersHTTP(t *testing.T) {
	url := fmt.Sprintf("%s/api/v1/workflows/admintests/development/name_a?limit=20&filters=eq(version,123)",
		GetTestHostEndpoint())
	getRequest, err := http.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	addHTTPRequestHeaders(getRequest)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(getRequest)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	octetStreamedWorkflowList := admin.WorkflowList{}
	proto.Unmarshal(body, &octetStreamedWorkflowList)
	assert.Len(t, octetStreamedWorkflowList.Workflows, 1)

	workflow := octetStreamedWorkflowList.Workflows[0]
	assert.True(t, proto.Equal(&core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "admintests",
		Domain:       "development",
		Name:         "name_a",
		Version:      "123",
	}, workflow.Id))
}
