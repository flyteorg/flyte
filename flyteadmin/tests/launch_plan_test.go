//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"

	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

var launchPlanIdentifier = core.Identifier{
	ResourceType: core.ResourceType_LAUNCH_PLAN,
	Project:      "admintests",
	Domain:       "development",
	Name:         "name",
	Version:      "version",
}

func getWorkflowCreateRequest() admin.WorkflowCreateRequest {
	return admin.WorkflowCreateRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_WORKFLOW,
			Project:      entityProjects[0],
			Domain:       entityDomains[0],
			Name:         entityNames[0],
			Version:      entityVersions[0],
		},
		Spec: &admin.WorkflowSpec{
			Template: &core.WorkflowTemplate{
				Id: &core.Identifier{
					ResourceType: core.ResourceType_WORKFLOW,
					Project:      entityProjects[0],
					Domain:       entityDomains[0],
					Name:         entityNames[0],
					Version:      entityVersions[0],
				},
				Interface: &core.TypedInterface{
					Inputs: &core.VariableMap{
						Variables: map[string]*core.Variable{
							"foo": {
								Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
							},
						},
					},
				},
				Nodes: []*core.Node{
					{
						Id: "I'm a node",
						Target: &core.Node_TaskNode{
							TaskNode: &core.TaskNode{
								Reference: &core.TaskNode_ReferenceId{
									ReferenceId: &core.Identifier{
										ResourceType: core.ResourceType_TASK,
										Project:      "admintests",
										Domain:       "development",
										Name:         "name_a",
										Version:      "123",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func getLaunchPlanCreateRequestWithCronSchedule(workflowIdentifier *core.Identifier, test_cron_expr string) admin.LaunchPlanCreateRequest {
	lpCreateReq := getLaunchPlanCreateRequest(workflowIdentifier)
	lpCreateReq.Spec.EntityMetadata = &admin.LaunchPlanMetadata{
		Schedule: &admin.Schedule{
			ScheduleExpression: &admin.Schedule_CronExpression{CronExpression: test_cron_expr},
		},
	}
	return lpCreateReq
}

func getLaunchPlanCreateRequest(workflowIdentifier *core.Identifier) admin.LaunchPlanCreateRequest {
	return admin.LaunchPlanCreateRequest{
		Id: &launchPlanIdentifier,
		Spec: &admin.LaunchPlanSpec{
			WorkflowId: workflowIdentifier,
			DefaultInputs: &core.ParameterMap{
				Parameters: map[string]*core.Parameter{
					"foo": {
						Var: &core.Variable{
							Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
						},
						Behavior: &core.Parameter_Default{
							Default: coreutils.MustMakeLiteral("foo-value"),
						},
					},
				},
			},
		},
	}
}

func getLaunchPlanCreateRequestWithFixedRateSchedule(workflowIdentifier *core.Identifier, test_rate_value uint32, test_rate_unit admin.FixedRateUnit) admin.LaunchPlanCreateRequest {
	lpCreateReq := getLaunchPlanCreateRequest(workflowIdentifier)
	lpCreateReq.Spec.EntityMetadata = &admin.LaunchPlanMetadata{
		Schedule: &admin.Schedule{
			ScheduleExpression: &admin.Schedule_Rate{
				Rate: &admin.FixedRate{
					Value: test_rate_value,
					Unit:  test_rate_unit,
				},
			},
		},
	}
	return lpCreateReq
}

func testOneListLaunchPlanHTTPGetRequest(t *testing.T, query string, expected_resp_num int) {
	url := fmt.Sprintf("%s/%s", GetTestHostEndpoint(), query)
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
	octetStreamedLaunchPlanList := admin.LaunchPlanList{}
	proto.Unmarshal(body, &octetStreamedLaunchPlanList)
	assert.Len(t, octetStreamedLaunchPlanList.LaunchPlans, expected_resp_num)
}

func getLaunchPlanHTTPQueryString(project string, domain string, name string) string {
	if name != "" {
		return fmt.Sprintf(
			"api/v1/launch_plans/%s/%s/%s?limit=20",
			project,
			domain,
			name,
		)
	}
	return fmt.Sprintf(
		"api/v1/launch_plans/%s/%s?limit=20",
		project,
		domain,
	)

}

func getLaunchPlanHTTPQueryStringWithFilter(project string, domain string, name string, filterStr string) string {
	if name != "" {
		return fmt.Sprintf(
			"api/v1/launch_plans/%s/%s/%s?limit=20&filters=%s",
			project,
			domain,
			name,
			url.QueryEscape(filterStr),
		)
	}
	return fmt.Sprintf(
		"api/v1/launch_plans/%s/%s?limit=20&filters=%s",
		project,
		domain,
		url.QueryEscape(filterStr),
	)
}

func TestCreateLaunchPlan(t *testing.T) {
	truncateAllTablesForTestingOnly()
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()
	insertTasksForTests(t, client)

	createWorkflowReq := getWorkflowCreateRequest()

	_, err := client.CreateWorkflow(ctx, &createWorkflowReq)
	assert.Nil(t, err)

	createLaunchPlanReq := getLaunchPlanCreateRequest(createWorkflowReq.Id)
	_, err = client.CreateLaunchPlan(ctx, &createLaunchPlanReq)
	assert.Nil(t, err)
}

func TestGetLaunchPlanHTTP(t *testing.T) {
	truncateAllTablesForTestingOnly()
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	insertTasksForTests(t, client)
	createWorkflowReq := getWorkflowCreateRequest()

	_, err := client.CreateWorkflow(ctx, &createWorkflowReq)
	assert.Nil(t, err)

	createLaunchPlanReq := getLaunchPlanCreateRequest(createWorkflowReq.Id)
	_, err = client.CreateLaunchPlan(ctx, &createLaunchPlanReq)

	url := fmt.Sprintf("%s/api/v1/launch_plans/admintests/development/name/version", GetTestHostEndpoint())
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
	octetStreamedLaunchPlan := admin.LaunchPlan{}
	proto.Unmarshal(body, &octetStreamedLaunchPlan)
	assert.True(t, proto.Equal(&launchPlanIdentifier, octetStreamedLaunchPlan.Id))
}

func TestEnableDisableLaunchPlan(t *testing.T) {
	// Insert test workflow data.
	truncateAllTablesForTestingOnly()
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	insertTasksForTests(t, client)
	createWorkflowReq := getWorkflowCreateRequest()

	_, err := client.CreateWorkflow(ctx, &createWorkflowReq)
	assert.Nil(t, err)

	createLaunchPlanReq := getLaunchPlanCreateRequest(createWorkflowReq.Id)
	_, err = client.CreateLaunchPlan(ctx, &createLaunchPlanReq)
	assert.Nil(t, err)

	// Test that enabling a launch plan succeeds.
	_, err = client.UpdateLaunchPlan(ctx, &admin.LaunchPlanUpdateRequest{
		Id:    &launchPlanIdentifier,
		State: admin.LaunchPlanState_ACTIVE,
	})
	assert.Nil(t, err)
	enabledLaunchPlan, err := client.GetLaunchPlan(ctx, &admin.ObjectGetRequest{
		Id: createLaunchPlanReq.Id,
	})
	assert.Nil(t, err)
	assert.Equal(t, admin.LaunchPlanState_ACTIVE, enabledLaunchPlan.Closure.State)

	// Make sure successive calls to set active are ... successful.
	_, err = client.UpdateLaunchPlan(ctx, &admin.LaunchPlanUpdateRequest{
		Id:    createLaunchPlanReq.Id,
		State: admin.LaunchPlanState_ACTIVE,
	})
	assert.Nil(t, err)
	enabledLaunchPlan, err = client.GetLaunchPlan(ctx, &admin.ObjectGetRequest{
		Id: createLaunchPlanReq.Id,
	})
	assert.Nil(t, err)
	assert.Equal(t, admin.LaunchPlanState_ACTIVE, enabledLaunchPlan.Closure.State)

	// Test that disabling a launch plan succeeds.
	_, err = client.UpdateLaunchPlan(ctx, &admin.LaunchPlanUpdateRequest{
		Id:    &launchPlanIdentifier,
		State: admin.LaunchPlanState_INACTIVE,
	})
	assert.Nil(t, err)

	disabledLaunchPlan, err := client.GetLaunchPlan(ctx, &admin.ObjectGetRequest{
		Id: createLaunchPlanReq.Id,
	})
	assert.Nil(t, err)
	assert.Equal(t, admin.LaunchPlanState_INACTIVE, disabledLaunchPlan.Closure.State)
}

// Tests that enabling one launch plan also disables the currently active version.
func TestUpdateActiveLaunchPlanVersion(t *testing.T) {
	// Insert test workflow data.
	truncateAllTablesForTestingOnly()
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	insertTasksForTests(t, client)
	createWorkflowReq := getWorkflowCreateRequest()

	_, err := client.CreateWorkflow(ctx, &createWorkflowReq)
	assert.Nil(t, err)

	// Create a test launch plan and set it to active.
	createLaunchPlanReq := getLaunchPlanCreateRequest(createWorkflowReq.Id)
	_, err = client.CreateLaunchPlan(ctx, &createLaunchPlanReq)
	assert.Nil(t, err)

	_, err = client.UpdateLaunchPlan(ctx, &admin.LaunchPlanUpdateRequest{
		Id:    &launchPlanIdentifier,
		State: admin.LaunchPlanState_ACTIVE,
	})
	assert.Nil(t, err)

	// Create a new version of the launch plan and set that to active.
	newIdentifier := core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      "admintests",
		Domain:       "development",
		Name:         "name",
		Version:      "new",
	}
	createNewLaunchPlanReq := getLaunchPlanCreateRequest(createWorkflowReq.Id)
	createNewLaunchPlanReq.Id = &newIdentifier
	_, err = client.CreateLaunchPlan(ctx, &createNewLaunchPlanReq)
	assert.Nil(t, err)

	_, err = client.UpdateLaunchPlan(ctx, &admin.LaunchPlanUpdateRequest{
		Id:    &newIdentifier,
		State: admin.LaunchPlanState_ACTIVE,
	})
	assert.Nil(t, err)

	enabledLaunchPlan, err := client.GetLaunchPlan(ctx, &admin.ObjectGetRequest{
		Id: &newIdentifier,
	})
	assert.Nil(t, err)
	assert.Equal(t, admin.LaunchPlanState_ACTIVE, enabledLaunchPlan.Closure.State)

	// Assert that by enabling a specific launch version admin also disables the previously active version.
	disabledLaunchPlan, err := client.GetLaunchPlan(ctx, &admin.ObjectGetRequest{
		Id: &launchPlanIdentifier,
	})
	assert.Nil(t, err)
	assert.Equal(t, admin.LaunchPlanState_INACTIVE, disabledLaunchPlan.Closure.State)
}

func TestListLaunchPlans(t *testing.T) {
	truncateAllTablesForTestingOnly()
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()

	insertTasksForTests(t, client)
	createWorkflowReq := getWorkflowCreateRequest()

	_, err := client.CreateWorkflow(ctx, &createWorkflowReq)
	assert.Nil(t, err)

	versionA := core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      "admintests",
		Domain:       "development",
		Name:         "name1",
		Version:      "a",
	}
	createLaunchPlanReq := getLaunchPlanCreateRequest(createWorkflowReq.Id)
	createLaunchPlanReq.Id = &versionA
	_, err = client.CreateLaunchPlan(ctx, &createLaunchPlanReq)
	assert.Nil(t, err)

	versionB := core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      "admintests",
		Domain:       "development",
		Name:         "name1",
		Version:      "b",
	}
	createLaunchPlanReq = getLaunchPlanCreateRequest(createWorkflowReq.Id)
	createLaunchPlanReq.Id = &versionB
	_, err = client.CreateLaunchPlan(ctx, &createLaunchPlanReq)
	assert.Nil(t, err)

	versionADifferentName := core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      "admintests",
		Domain:       "development",
		Name:         "name2",
		Version:      "a",
	}
	createLaunchPlanReq = getLaunchPlanCreateRequest(createWorkflowReq.Id)
	createLaunchPlanReq.Id = &versionADifferentName
	_, err = client.CreateLaunchPlan(ctx, &createLaunchPlanReq)
	assert.Nil(t, err)

	versionADifferentProj := core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      "flytekit",
		Domain:       "development",
		Name:         "name1",
		Version:      "a",
	}
	createLaunchPlanReq = getLaunchPlanCreateRequest(createWorkflowReq.Id)
	createLaunchPlanReq.Id = &versionADifferentProj
	_, err = client.CreateLaunchPlan(ctx, &createLaunchPlanReq)
	assert.Nil(t, err)
	conn.Close()

	t.Run("TestListLaunchPlansGrpc", testListLaunchPlansGrpc)
	t.Run("TestListLaunchPlansHTTP", testListLaunchPlansHTTP)
	t.Run("TestListLaunchPlansFiltersGrpc", testListLaunchPlansFiltersGrpc)
	t.Run("TestListLaunchPlansFiltersHTTP", testListLaunchPlansFiltersHTTP)
	t.Run("TestListLaunchPlansPaginationGrpc", testListLaunchPlansPaginationGrpc)
	t.Run("TestListLaunchPlansPaginationHTTP", testListLaunchPlansPaginationHTTP)
	t.Run("TestListLaunchPlanIdentifiersGrpc", testListLaunchPlanIdentifiersGrpc)
	t.Run("TestListLaunchPlanIdentifiersHTTP", testListLaunchPlanIdentifiersHTTP)
}

func TestListLaunchPlansFilterOnSchedule(t *testing.T) {
	truncateAllTablesForTestingOnly()
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()

	insertTasksForTests(t, client)
	createWorkflowReq := getWorkflowCreateRequest()

	_, err := client.CreateWorkflow(ctx, &createWorkflowReq)
	assert.Nil(t, err)

	versionACronSchedule := core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      "admintests",
		Domain:       "development",
		Name:         "name1",
		Version:      "a",
	}
	createLaunchPlanReq := getLaunchPlanCreateRequestWithCronSchedule(createWorkflowReq.Id, "* * * * *")
	createLaunchPlanReq.Id = &versionACronSchedule
	_, err = client.CreateLaunchPlan(ctx, &createLaunchPlanReq)
	assert.Nil(t, err)

	versionBRateSchedule := core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      "admintests",
		Domain:       "development",
		Name:         "name1",
		Version:      "b",
	}
	createLaunchPlanReq = getLaunchPlanCreateRequestWithFixedRateSchedule(createWorkflowReq.Id, 24, admin.FixedRateUnit_HOUR)
	createLaunchPlanReq.Id = &versionBRateSchedule
	_, err = client.CreateLaunchPlan(ctx, &createLaunchPlanReq)
	assert.Nil(t, err)

	_, err = client.UpdateLaunchPlan(ctx, &admin.LaunchPlanUpdateRequest{
		Id:    createLaunchPlanReq.Id,
		State: admin.LaunchPlanState_ACTIVE,
	})
	assert.Nil(t, err)
	enabledLaunchPlan, err := client.GetLaunchPlan(ctx, &admin.ObjectGetRequest{
		Id: createLaunchPlanReq.Id,
	})
	assert.Nil(t, err)
	assert.Equal(t, admin.LaunchPlanState_ACTIVE, enabledLaunchPlan.Closure.State)

	noScheduleLaunchPlan := core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      "admintests",
		Domain:       "development",
		Name:         "name2",
		Version:      "a",
	}
	createLaunchPlanReq = getLaunchPlanCreateRequest(createWorkflowReq.Id)
	createLaunchPlanReq.Id = &noScheduleLaunchPlan
	_, err = client.CreateLaunchPlan(ctx, &createLaunchPlanReq)
	assert.Nil(t, err)

	rateScheduleDifferentProj := core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      "flytekit",
		Domain:       "development",
		Name:         "name1",
		Version:      "a",
	}
	createLaunchPlanReq = getLaunchPlanCreateRequestWithFixedRateSchedule(createWorkflowReq.Id, 20, admin.FixedRateUnit_DAY)
	createLaunchPlanReq.Id = &rateScheduleDifferentProj
	_, err = client.CreateLaunchPlan(ctx, &createLaunchPlanReq)
	assert.Nil(t, err)

	_, err = client.UpdateLaunchPlan(ctx, &admin.LaunchPlanUpdateRequest{
		Id:    createLaunchPlanReq.Id,
		State: admin.LaunchPlanState_ACTIVE,
	})
	assert.Nil(t, err)
	enabledLaunchPlan, err = client.GetLaunchPlan(ctx, &admin.ObjectGetRequest{
		Id: createLaunchPlanReq.Id,
	})
	assert.Nil(t, err)
	assert.Equal(t, admin.LaunchPlanState_ACTIVE, enabledLaunchPlan.Closure.State)

	conn.Close()

	t.Run("TestListLaunchPlanFilterOnScheduleGrpc", testListLaunchPlansFilterOnScheduleGrpc)
	t.Run("TestListLaunchPlanFilterOnScheduleHTTP", testListLaunchPlansFilterOnScheduleHTTP)
	t.Run("TestListLaunchPlanWithActiveScheduleGrpc", testListLaunchPlansWithActiveScheduleGrpc)
	t.Run("TestListLaunchPlanWithActiveScheduleHTTP", testListLaunchPlansWithActiveScheduleHTTP)
}

func testListLaunchPlansGrpc(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	resp, err := client.ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
		},
		Limit: 10,
	})
	assert.Nil(t, err)
	assert.Len(t, resp.LaunchPlans, 3)

	resp, err = client.ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
			Name:    "name1",
		},
		Limit: 10,
	})
	assert.Nil(t, err)
	assert.Len(t, resp.LaunchPlans, 2)
}

func testListLaunchPlansHTTP(t *testing.T) {
	var queryStr string

	queryStr = getLaunchPlanHTTPQueryString("admintests", "development", "")
	testOneListLaunchPlanHTTPGetRequest(t, queryStr, 3)

	queryStr = getLaunchPlanHTTPQueryString("admintests", "development", "name1")
	testOneListLaunchPlanHTTPGetRequest(t, queryStr, 2)
}

func testListLaunchPlansFiltersGrpc(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	resp, err := client.ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
		},
		Filters: "eq(version, b)",
		Limit:   10,
	})
	assert.Nil(t, err)
	assert.Len(t, resp.LaunchPlans, 1)

	resp, err = client.ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
		},
		Filters: "eq(workflow.name, i-don't-exist)",
		Limit:   10,
	})
	assert.Nil(t, err)
	assert.Len(t, resp.LaunchPlans, 0)
}

func testListLaunchPlansFiltersHTTP(t *testing.T) {
	var queryStr string

	queryStr = getLaunchPlanHTTPQueryStringWithFilter(
		"admintests", "development", "", "eq(version,b)")
	testOneListLaunchPlanHTTPGetRequest(t, queryStr, 1)

	queryStr = getLaunchPlanHTTPQueryStringWithFilter(
		"admintests", "development", "", "eq(workflow.name,i-don't-exist)")
	testOneListLaunchPlanHTTPGetRequest(t, queryStr, 0)

}

func testListLaunchPlansFilterOnScheduleGrpc(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	resp, err := client.ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
		},
		Filters: "eq(schedule_type, NONE)",
		Limit:   10,
	})
	assert.Nil(t, err)
	assert.Len(t, resp.LaunchPlans, 1)

	resp, err = client.ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
		},
		Filters: "eq(schedule_type, CRON)",
		Limit:   10,
	})
	assert.Nil(t, err)
	assert.Len(t, resp.LaunchPlans, 1)

	resp, err = client.ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
		},
		Filters: "eq(schedule_type, RATE)",
		Limit:   10,
	})
	assert.Nil(t, err)
	assert.Len(t, resp.LaunchPlans, 1)

	resp, err = client.ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "flytekit",
			Domain:  "development",
		},
		Filters: "eq(schedule_type, NONE)",
		Limit:   10,
	})
	assert.Nil(t, err)
	assert.Len(t, resp.LaunchPlans, 0)

	resp, err = client.ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "flytekit",
			Domain:  "development",
		},
		Filters: "eq(schedule_type, CRON)",
		Limit:   10,
	})
	assert.Nil(t, err)
	assert.Len(t, resp.LaunchPlans, 0)

	resp, err = client.ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "flytekit",
			Domain:  "development",
		},
		Filters: "eq(schedule_type, RATE)",
		Limit:   10,
	})
	assert.Nil(t, err)
	assert.Len(t, resp.LaunchPlans, 1)
}

func testListLaunchPlansFilterOnScheduleHTTP(t *testing.T) {

	var queryStr string

	queryStr = getLaunchPlanHTTPQueryStringWithFilter(
		"admintests", "development", "", "eq(schedule_type,NONE)")
	testOneListLaunchPlanHTTPGetRequest(t, queryStr, 1)

	queryStr = getLaunchPlanHTTPQueryStringWithFilter(
		"admintests", "development", "", "eq(schedule_type,CRON)")
	testOneListLaunchPlanHTTPGetRequest(t, queryStr, 1)

	queryStr = getLaunchPlanHTTPQueryStringWithFilter(
		"admintests", "development", "", "eq(schedule_type,RATE)")
	testOneListLaunchPlanHTTPGetRequest(t, queryStr, 1)

	queryStr = getLaunchPlanHTTPQueryStringWithFilter(
		"flytekit", "development", "", "eq(schedule_type,NONE)")
	testOneListLaunchPlanHTTPGetRequest(t, queryStr, 0)

	queryStr = getLaunchPlanHTTPQueryStringWithFilter(
		"flytekit", "development", "", "eq(schedule_type,CRON)")
	testOneListLaunchPlanHTTPGetRequest(t, queryStr, 0)

	queryStr = getLaunchPlanHTTPQueryStringWithFilter(
		"flytekit", "development", "", "eq(schedule_type,RATE)")
	testOneListLaunchPlanHTTPGetRequest(t, queryStr, 1)
}

func testListLaunchPlansWithActiveScheduleGrpc(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	resp, err := client.ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
		},
		Filters: "eq(schedule_type, NONE)+ne(state, 1)",
		Limit:   10,
	})
	assert.Nil(t, err)
	assert.Len(t, resp.LaunchPlans, 1)

	resp, err = client.ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
		},
		Filters: "eq(schedule_type, CRON)+ne(state, 1)",
		Limit:   10,
	})
	assert.Nil(t, err)
	assert.Len(t, resp.LaunchPlans, 1)

	resp, err = client.ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
		},
		Filters: "eq(schedule_type, RATE)+eq(state, 1)",
		Limit:   10,
	})
	assert.Nil(t, err)
	assert.Len(t, resp.LaunchPlans, 1)

	resp, err = client.ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "flytekit",
			Domain:  "development",
		},
		Filters: "eq(schedule_type, NONE)+eq(state, 1)",
		Limit:   10,
	})
	assert.Nil(t, err)
	assert.Len(t, resp.LaunchPlans, 0)

	resp, err = client.ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "flytekit",
			Domain:  "development",
		},
		Filters: "eq(schedule_type, CRON)+eq(state, 1)",
		Limit:   10,
	})
	assert.Nil(t, err)
	assert.Len(t, resp.LaunchPlans, 0)

	resp, err = client.ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "flytekit",
			Domain:  "development",
		},
		Filters: "eq(schedule_type, RATE)+eq(state, 1)",
		Limit:   10,
	})
	assert.Nil(t, err)
	assert.Len(t, resp.LaunchPlans, 1)
}

func testListLaunchPlansWithActiveScheduleHTTP(t *testing.T) {

	var queryStr string

	queryStr = getLaunchPlanHTTPQueryStringWithFilter(
		"admintests", "development", "", "eq(schedule_type,NONE)+ne(state,1)")
	testOneListLaunchPlanHTTPGetRequest(t, queryStr, 1)

	queryStr = getLaunchPlanHTTPQueryStringWithFilter(
		"admintests", "development", "", "eq(schedule_type,CRON)+ne(state,1)")
	testOneListLaunchPlanHTTPGetRequest(t, queryStr, 1)

	queryStr = getLaunchPlanHTTPQueryStringWithFilter(
		"admintests", "development", "", "eq(schedule_type,RATE)+eq(state,1)")
	testOneListLaunchPlanHTTPGetRequest(t, queryStr, 1)

	queryStr = getLaunchPlanHTTPQueryStringWithFilter(
		"flytekit", "development", "", "eq(schedule_type,NONE)+ne(state,1)")
	testOneListLaunchPlanHTTPGetRequest(t, queryStr, 0)

	queryStr = getLaunchPlanHTTPQueryStringWithFilter(
		"flytekit", "development", "", "eq(schedule_type,CRON)+eq(state,1)")
	testOneListLaunchPlanHTTPGetRequest(t, queryStr, 0)

	queryStr = getLaunchPlanHTTPQueryStringWithFilter(
		"flytekit", "development", "", "eq(schedule_type,RATE)+eq(state,1)")
	testOneListLaunchPlanHTTPGetRequest(t, queryStr, 1)

}

func testListLaunchPlansPaginationGrpc(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	resp, err := client.ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
		},
		Limit: 2,
	})
	assert.Nil(t, err)
	assert.Len(t, resp.LaunchPlans, 2)
	assert.NotEmpty(t, resp.Token)

	resp, err = client.ListLaunchPlans(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "admintests",
			Domain:  "development",
		},
		Limit: 2,
		Token: resp.Token,
	})
	assert.Nil(t, err)
	assert.Len(t, resp.LaunchPlans, 1)
	assert.Empty(t, resp.Token)
}

func testListLaunchPlansPaginationHTTP(t *testing.T) {
	url := fmt.Sprintf("%s/api/v1/launch_plans/admintests/development?limit=2", GetTestHostEndpoint())
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
	octetStreamedLaunchPlanList := admin.LaunchPlanList{}
	proto.Unmarshal(body, &octetStreamedLaunchPlanList)
	assert.Len(t, octetStreamedLaunchPlanList.LaunchPlans, 2)
	assert.NotEmpty(t, octetStreamedLaunchPlanList.Token)

	url = fmt.Sprintf("%s/api/v1/launch_plans/admintests/development?limit=2&token=%s", GetTestHostEndpoint(),
		octetStreamedLaunchPlanList.Token)
	getRequest, err = http.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	addHTTPRequestHeaders(getRequest)

	resp, err = httpClient.Do(getRequest)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ = ioutil.ReadAll(resp.Body)
	octetStreamedLaunchPlanList = admin.LaunchPlanList{}
	proto.Unmarshal(body, &octetStreamedLaunchPlanList)
	assert.Len(t, octetStreamedLaunchPlanList.LaunchPlans, 1)
	assert.Empty(t, octetStreamedLaunchPlanList.Token)
}

func testListLaunchPlanIdentifiersGrpc(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	descendingResponse, err := client.ListLaunchPlanIds(ctx, &admin.NamedEntityIdentifierListRequest{
		Project: "admintests",
		Domain:  "development",
		Limit:   10,
		SortBy: &admin.Sort{
			Direction: admin.Sort_DESCENDING,
			Key:       "name",
		},
	})
	assert.Nil(t, err)
	assert.Len(t, descendingResponse.Entities, 2)

	ascendingResponse, err := client.ListLaunchPlanIds(ctx, &admin.NamedEntityIdentifierListRequest{
		Project: "admintests",
		Domain:  "development",
		Limit:   10,
		SortBy: &admin.Sort{
			Direction: admin.Sort_ASCENDING,
			Key:       "name",
		},
	})
	assert.Nil(t, err)
	assert.Len(t, ascendingResponse.Entities, 2)

	assert.Equal(t, ascendingResponse.Entities[0].Name, descendingResponse.Entities[1].Name)
	assert.Equal(t, ascendingResponse.Entities[1].Name, descendingResponse.Entities[0].Name)
}

func testListLaunchPlanIdentifiersHTTP(t *testing.T) {
	url := fmt.Sprintf(
		"%s/api/v1/launch_plan_ids/admintests/development?limit=10&sort_by.key=name&sort_by.direction=ASCENDING",
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
	ascendingResponse := admin.NamedEntityIdentifierList{}
	proto.Unmarshal(body, &ascendingResponse)
	assert.Len(t, ascendingResponse.Entities, 2)

	url = fmt.Sprintf(
		"%s/api/v1/launch_plan_ids/admintests/development?limit=10&sort_by.key=name&sort_by.direction=DESCENDING",
		GetTestHostEndpoint())
	getRequest, err = http.NewRequest("GET", url, nil)
	assert.Nil(t, err)
	addHTTPRequestHeaders(getRequest)

	resp, err = httpClient.Do(getRequest)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ = ioutil.ReadAll(resp.Body)
	descendingResponse := admin.NamedEntityIdentifierList{}
	proto.Unmarshal(body, &descendingResponse)
	assert.Len(t, descendingResponse.Entities, 2)

	assert.Equal(t, ascendingResponse.Entities[0].Name, descendingResponse.Entities[1].Name)
	assert.Equal(t, ascendingResponse.Entities[1].Name, descendingResponse.Entities[0].Name)
}
