//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"

	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
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
