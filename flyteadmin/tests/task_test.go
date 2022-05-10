//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

func TestCreateBasic(t *testing.T) {
	ctx := context.Background()
	truncateAllTablesForTestingOnly()

	// Get client and gRPC connection
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	var req admin.TaskCreateRequest

	// Empty spec
	taskIdentifier := core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "admintests",
		Domain:       "development",
		Name:         "mytask.name",
		Version:      "jfdklsa",
	}
	req = admin.TaskCreateRequest{
		Id: &taskIdentifier,
	}
	_, err := client.CreateTask(ctx, &req)
	assert.Error(t, err, "should error when spec is empty")

	// Test that task creation and retrieval works
	valid := testutils.GetValidTaskRequest()
	response, err := client.CreateTask(ctx, &valid)
	assert.NoError(t, err)
	task, err := client.GetTask(ctx, &admin.ObjectGetRequest{
		Id: valid.Id,
	})
	assert.NoError(t, err)
	assert.True(t, proto.Equal(valid.Id, task.Id))
	// Test to make sure URN matches in the future
	// Currently the returned object does not have the URN

	// Test creating twice in a row fails on uniqueness
	response, err = client.CreateTask(ctx, &valid)
	assert.Nil(t, response)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "already exists"))
}

func TestCreateTaskWithoutContainer(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	req := testutils.GetValidTaskRequest()
	req.Id.Version = "TestCreateTaskWithoutContainer"
	req.Spec.Template.Target = nil

	_, err := client.CreateTask(ctx, &req)
	assert.Nil(t, err)
}

func TestGetTasks(t *testing.T) {
	truncateAllTablesForTestingOnly()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	insertTasksForTests(t, client)

	t.Run("TestGetTaskGrpc", testGetTaskGrpc)
	t.Run("TestGetTaskHTTP", testGetTaskHTTP)
	t.Run("TestListTaskGrpc", testListTaskGrpc)
	t.Run("TestListTaskHTTP", testListTaskHTTP)
	t.Run("TestListTask_PaginationGrpc", testListTask_PaginationGrpc)
	t.Run("TestListTask_PaginationHTTP", testListTask_PaginationHTTP)
	t.Run("TestListTask_FiltersGrpc", testListTask_FiltersGrpc)
	t.Run("TestListTask_FiltersHTTP", testListTask_FiltersHTTP)
}

func testGetTaskGrpc(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	identifier := core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "admintests",
		Domain:       "development",
		Name:         "name_a",
		Version:      "123",
	}
	task, err := client.GetTask(ctx, &admin.ObjectGetRequest{
		Id: &identifier,
	})
	assert.NoError(t, err)
	assert.True(t, proto.Equal(&identifier, task.Id))
}

func testGetTaskHTTP(t *testing.T) {
	url := fmt.Sprintf("%s/api/v1/tasks/admintests/development/name_a/123", GetTestHostEndpoint())
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
	octetStreamedTask := admin.Task{}
	proto.Unmarshal(body, &octetStreamedTask)
	assert.Equal(t, "admintests", octetStreamedTask.Id.GetProject())
	assert.Equal(t, "development", octetStreamedTask.Id.GetDomain())
	assert.Equal(t, "name_a", octetStreamedTask.Id.GetName())
	assert.Equal(t, "123", octetStreamedTask.Id.Version)
}

func testListTaskGrpc(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	tasks, err := client.ListTasks(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
			Name:    "name_a",
		},
		Limit: 20,
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       "version",
		},
	})
	assert.NoError(t, err)
	assert.Len(t, tasks.Tasks, 3)

	for _, task := range tasks.Tasks {
		assert.Equal(t, "admintests", task.Id.Project)
		assert.Equal(t, "development", task.Id.Domain)
		assert.Equal(t, "name_a", task.Id.Name)
		assert.Contains(t, entityVersions, task.Id.Version)
	}
	assert.True(t, tasks.Tasks[0].Id.Version < tasks.Tasks[1].Id.Version)
	assert.True(t, tasks.Tasks[1].Id.Version < tasks.Tasks[2].Id.Version)
}

func testListTaskHTTP(t *testing.T) {

	url := fmt.Sprintf(
		"%s/api/v1/tasks/admintests/development/name_a?sort_by.direction=ASCENDING&sort_by.key=version&limit=20",
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
	octetStreamedTaskList := admin.TaskList{}
	proto.Unmarshal(body, &octetStreamedTaskList)
	assert.Len(t, octetStreamedTaskList.Tasks, 3)

	for _, task := range octetStreamedTaskList.Tasks {
		assert.Equal(t, "admintests", task.Id.Project)
		assert.Equal(t, "development", task.Id.Domain)
		assert.Equal(t, "name_a", task.Id.Name)
		assert.Contains(t, entityVersions, task.Id.Version)
	}
	assert.True(t, octetStreamedTaskList.Tasks[0].Id.Version < octetStreamedTaskList.Tasks[1].Id.Version)
	assert.True(t, octetStreamedTaskList.Tasks[1].Id.Version < octetStreamedTaskList.Tasks[2].Id.Version)
}

func testListTask_PaginationGrpc(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	tasks, err := client.ListTasks(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
			Name:    "name_a",
		},
		Limit: 2,
	})
	assert.NoError(t, err)
	assert.Len(t, tasks.Tasks, 2)

	firstResponseVersions := make([]string, 2)
	for idx, task := range tasks.Tasks {
		assert.Equal(t, "admintests", task.Id.Project)
		assert.Equal(t, "development", task.Id.Domain)
		assert.Equal(t, "name_a", task.Id.Name)
		assert.Contains(t, entityVersions, task.Id.Version)

		firstResponseVersions[idx] = task.Id.Version
	}

	tasks, err = client.ListTasks(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
			Name:    "name_a",
		},
		Limit: 2,
		Token: "2",
	})
	assert.NoError(t, err)
	assert.Len(t, tasks.Tasks, 1)

	for _, task := range tasks.Tasks {
		assert.Equal(t, "admintests", task.Id.Project)
		assert.Equal(t, "development", task.Id.Domain)
		assert.Equal(t, "name_a", task.Id.Name)
		assert.Contains(t, entityVersions, task.Id.Version)
		assert.NotContains(t, firstResponseVersions, task.Id.Version)
	}
}

func testListTask_PaginationHTTP(t *testing.T) {
	url := fmt.Sprintf(
		"%s/api/v1/tasks/admintests/development/name_a?sort_by.direction=ASCENDING&sort_by.key=version&limit=2",
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
	octetStreamedTaskList := admin.TaskList{}
	proto.Unmarshal(body, &octetStreamedTaskList)
	assert.Len(t, octetStreamedTaskList.Tasks, 2)

	firstResponseVersions := make([]string, 2)
	for idx, task := range octetStreamedTaskList.Tasks {
		assert.Equal(t, "admintests", task.Id.Project)
		assert.Equal(t, "development", task.Id.Domain)
		assert.Equal(t, "name_a", task.Id.Name)
		assert.Contains(t, entityVersions, task.Id.Version)

		firstResponseVersions[idx] = task.Id.Version
	}

	url = fmt.Sprintf(
		"%s/api/v1/tasks/admintests/development/name_a?sort_by.direction=ASCENDING&sort_by.key=version&limit=2&token=%s",
		GetTestHostEndpoint(), octetStreamedTaskList.Token)
	getRequest, err = http.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	addHTTPRequestHeaders(getRequest)
	resp, err = httpClient.Do(getRequest)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ = ioutil.ReadAll(resp.Body)
	octetStreamedTaskList = admin.TaskList{}
	proto.Unmarshal(body, &octetStreamedTaskList)
	assert.Len(t, octetStreamedTaskList.Tasks, 1)

	for _, task := range octetStreamedTaskList.Tasks {
		assert.Equal(t, "admintests", task.Id.Project)
		assert.Equal(t, "development", task.Id.Domain)
		assert.Equal(t, "name_a", task.Id.Name)
		assert.Contains(t, entityVersions, task.Id.Version)
		assert.NotContains(t, firstResponseVersions, task.Id.Version)
	}
}

func testListTask_FiltersGrpc(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	tasks, err := client.ListTasks(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
		},
		Limit:   20,
		Filters: "eq(version,123)",
	})
	assert.NoError(t, err)
	assert.Len(t, tasks.Tasks, 3)

	for _, task := range tasks.Tasks {
		assert.Equal(t, "admintests", task.Id.Project)
		assert.Equal(t, "development", task.Id.Domain)
		assert.Contains(t, []string{"name_a", "name_b", "name_c"}, task.Id.Name)
		assert.Equal(t, "123", task.Id.Version)
	}
}

func testListTask_FiltersHTTP(t *testing.T) {
	url := fmt.Sprintf(
		"%s/api/v1/tasks/admintests/development?filters=eq(version,123)&limit=20",
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
	octetStreamedTaskList := admin.TaskList{}
	proto.Unmarshal(body, &octetStreamedTaskList)
	assert.Len(t, octetStreamedTaskList.Tasks, 3)

	for _, task := range octetStreamedTaskList.Tasks {
		assert.Equal(t, "admintests", task.Id.Project)
		assert.Equal(t, "development", task.Id.Domain)
		assert.Contains(t, []string{"name_a", "name_b", "name_c"}, task.Id.Name)
		assert.Equal(t, "123", task.Id.Version)
	}
}

func TestListTaskIdsWithMultipleTaskVersions(t *testing.T) {
	ctx := context.Background()
	truncateAllTablesForTestingOnly()

	// Get client and gRPC connection
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	// Create two versions of one task and one version of another task, and one unrelated task
	valid := testutils.GetValidTaskRequestWithOverrides("project", "domain", "name", "v1")
	_, err := client.CreateTask(ctx, &valid)
	assert.NoError(t, err)

	valid = testutils.GetValidTaskRequestWithOverrides("project", "domain", "name", "v2")
	_, err = client.CreateTask(ctx, &valid)
	assert.NoError(t, err)

	valid = testutils.GetValidTaskRequestWithOverrides("project", "domain", "secondtask", "v1")
	_, err = client.CreateTask(ctx, &valid)
	assert.NoError(t, err)

	// This last one should not show up.
	valid = testutils.GetValidTaskRequestWithOverrides("project", "development", "hidden", "v1")
	_, err = client.CreateTask(ctx, &valid)
	assert.NoError(t, err)

	// We should only get back two entities
	res, err := client.ListTaskIds(ctx, &admin.NamedEntityIdentifierListRequest{
		Project: "project",
		Domain:  "domain",
		Limit:   100,
	})
	entities := res.GetEntities()
	assert.Equal(t, 2, len(entities))
	assert.Equal(t, "project", entities[0].GetProject())
	assert.Equal(t, "domain", entities[0].GetDomain())
	assert.Equal(t, "name", entities[0].GetName())
	assert.Equal(t, "project", entities[1].GetProject())
	assert.Equal(t, "domain", entities[1].GetDomain())
	assert.Equal(t, "secondtask", entities[1].GetName())

	// Also test the parallel HTTP endpoint.
	url := fmt.Sprintf(
		"%s/api/v1/task_ids/project/domain?limit=20",
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
	octetStreamedTaskIdList := admin.NamedEntityIdentifierList{}
	proto.Unmarshal(body, &octetStreamedTaskIdList)

	entities = octetStreamedTaskIdList.Entities
	assert.Len(t, entities, 2)
	assert.Equal(t, "project", entities[0].GetProject())
	assert.Equal(t, "domain", entities[0].GetDomain())
	assert.Equal(t, "name", entities[0].GetName())
	assert.Equal(t, "project", entities[1].GetProject())
	assert.Equal(t, "domain", entities[1].GetDomain())
	assert.Equal(t, "secondtask", entities[1].GetName())
}

func TestListTaskIdsWithPagination(t *testing.T) {
	ctx := context.Background()
	truncateAllTablesForTestingOnly()

	// Get client and gRPC connection
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	// Create two versions of one task and one version of another task, and one unrelated task
	valid := testutils.GetValidTaskRequestWithOverrides("project", "domain", "name", "v1")
	_, err := client.CreateTask(ctx, &valid)
	assert.NoError(t, err)

	valid = testutils.GetValidTaskRequestWithOverrides("project", "domain", "name", "v2")
	_, err = client.CreateTask(ctx, &valid)
	assert.NoError(t, err)

	valid = testutils.GetValidTaskRequestWithOverrides("project", "domain", "secondtask", "v1")
	_, err = client.CreateTask(ctx, &valid)
	assert.NoError(t, err)

	// This last one should not show up.
	valid = testutils.GetValidTaskRequestWithOverrides("project", "development", "hidden", "v1")
	_, err = client.CreateTask(ctx, &valid)
	assert.NoError(t, err)

	// We should only get back one entity
	res, err := client.ListTaskIds(ctx, &admin.NamedEntityIdentifierListRequest{
		Project: "project",
		Domain:  "domain",
		Limit:   1,
	})
	entities := res.GetEntities()
	assert.Equal(t, 1, len(entities))
	assert.Equal(t, "project", entities[0].GetProject())
	assert.Equal(t, "domain", entities[0].GetDomain())
	assert.Equal(t, "name", entities[0].GetName())
	assert.NotEmpty(t, res.Token)

	res, err = client.ListTaskIds(ctx, &admin.NamedEntityIdentifierListRequest{
		Project: "project",
		Domain:  "domain",
		Limit:   10,
		Token:   "1",
	})
	entities = res.GetEntities()
	assert.Equal(t, 1, len(entities))
	assert.Equal(t, "project", entities[0].GetProject())
	assert.Equal(t, "domain", entities[0].GetDomain())
	assert.Equal(t, "secondtask", entities[0].GetName())

	// Also test the parallel HTTP endpoint.
	url := fmt.Sprintf(
		"%s/api/v1/task_ids/project/domain?limit=1",
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
	octetStreamedTaskIdList := admin.NamedEntityIdentifierList{}
	proto.Unmarshal(body, &octetStreamedTaskIdList)

	entities = octetStreamedTaskIdList.Entities
	assert.Len(t, entities, 1)
	assert.Equal(t, "project", entities[0].GetProject())
	assert.Equal(t, "domain", entities[0].GetDomain())
	assert.Equal(t, "name", entities[0].GetName())
	assert.NotEmpty(t, octetStreamedTaskIdList.Token)
}

func TestCreateInvalidTaskType(t *testing.T) {
	ctx := context.Background()
	truncateAllTablesForTestingOnly()

	// Get client and gRPC connection
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	// Test that task creation fails based on task type
	valid := testutils.GetValidTaskRequest()
	valid.Spec.Template.Type = "sparkonk8s"
	_, err := client.CreateTask(ctx, &valid)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "task type must be whitelisted before use")
}
