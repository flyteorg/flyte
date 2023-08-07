//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/repositories"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
)

var workflowExecutionID = &core.WorkflowExecutionIdentifier{
	Project: project,
	Domain:  domain,
	Name:    name,
}

func TestUpdateWorkflowExecution(t *testing.T) {
	truncateAllTablesForTestingOnly()
	populateWorkflowExecutionForTestingOnly(project, domain, name)

	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	beganAt := time.Now()
	beganAtProto, _ := ptypes.TimestampProto(beganAt)

	_, err := client.CreateWorkflowEvent(ctx, &admin.WorkflowExecutionEventRequest{
		RequestId: "request id",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: workflowExecutionID,
			Phase:       core.WorkflowExecution_RUNNING,
			OccurredAt:  beganAtProto,
		},
	})
	assert.Nil(t, err)
	resp, err := client.GetExecution(ctx, &admin.WorkflowExecutionGetRequest{
		Id: workflowExecutionID,
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(workflowExecutionID, resp.Id))
	assert.Equal(t, core.WorkflowExecution_RUNNING, resp.Closure.Phase)
	assert.True(t, proto.Equal(beganAtProto, resp.Closure.StartedAt))
	assert.True(t, proto.Equal(beganAtProto, resp.Closure.UpdatedAt))

	// Now mark the execution as successful and verify this updates the execution as we'd expect.
	duration := 5 * time.Minute
	succeededAt := beganAt.Add(duration)
	succeededAtProto, _ := ptypes.TimestampProto(succeededAt)
	outputURI := "output uri"
	_, err = client.CreateWorkflowEvent(ctx, &admin.WorkflowExecutionEventRequest{
		RequestId: "request id",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId:  workflowExecutionID,
			Phase:        core.WorkflowExecution_SUCCEEDED,
			OccurredAt:   succeededAtProto,
			OutputResult: &event.WorkflowExecutionEvent_OutputUri{OutputUri: outputURI},
		},
	})
	assert.Nil(t, err)
	resp, err = client.GetExecution(ctx, &admin.WorkflowExecutionGetRequest{
		Id: workflowExecutionID,
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(workflowExecutionID, resp.Id))
	assert.Equal(t, core.WorkflowExecution_SUCCEEDED, resp.Closure.Phase)
	assert.True(t, proto.Equal(beganAtProto, resp.Closure.StartedAt))
	assert.True(t, proto.Equal(succeededAtProto, resp.Closure.UpdatedAt))
	assert.NotEmpty(t, resp.Closure.Duration)
	assert.Equal(t, outputURI, resp.Closure.GetOutputs().GetUri())
}

func TestUpdateWorkflowExecution_InvalidPhaseTransitions(t *testing.T) {
	truncateAllTablesForTestingOnly()
	populateWorkflowExecutionForTestingOnly(project, domain, name)

	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	beganAt := time.Now()
	beganAtProto, _ := ptypes.TimestampProto(beganAt)

	_, err := client.CreateWorkflowEvent(ctx, &admin.WorkflowExecutionEventRequest{
		RequestId: "request id",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: workflowExecutionID,
			Phase:       core.WorkflowExecution_SUCCEEDED,
			OccurredAt:  beganAtProto,
		},
	})
	assert.Nil(t, err)
	resp, err := client.GetExecution(ctx, &admin.WorkflowExecutionGetRequest{
		Id: workflowExecutionID,
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(workflowExecutionID, resp.Id))
	assert.Equal(t, core.WorkflowExecution_SUCCEEDED, resp.Closure.Phase)
	assert.True(t, proto.Equal(beganAtProto, resp.Closure.UpdatedAt))

	// Now mark the execution as running and verify this fails to update as we'd expect.
	_, err = client.CreateWorkflowEvent(ctx, &admin.WorkflowExecutionEventRequest{
		RequestId: "request id",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: workflowExecutionID,
			Phase:       core.WorkflowExecution_RUNNING,
		},
	})
	assert.NotNil(t, err)

	// Try again with failed and verify this fails to update as we'd expect.
	_, err = client.CreateWorkflowEvent(ctx, &admin.WorkflowExecutionEventRequest{
		RequestId: "request id",
		Event: &event.WorkflowExecutionEvent{
			ExecutionId: workflowExecutionID,
			Phase:       core.WorkflowExecution_FAILED,
		},
	})
	assert.NotNil(t, err)

	// Assert the execution remains unchanged
	resp, err = client.GetExecution(ctx, &admin.WorkflowExecutionGetRequest{
		Id: workflowExecutionID,
	})
	assert.Nil(t, err)
	assert.True(t, proto.Equal(workflowExecutionID, resp.Id))
	assert.Equal(t, core.WorkflowExecution_SUCCEEDED, resp.Closure.Phase)
	assert.True(t, proto.Equal(beganAtProto, resp.Closure.UpdatedAt))
}

func populateWorkflowExecutionsForTestingOnly() {
	insertExecutionStatements := []string{
		// Insert a couple of executions with the same project + domain for the same launch plan & workflow
		fmt.Sprintf(insertExecutionQueryStr, "project1", "domain1", "name1", "RUNNING", 1, 2),
		fmt.Sprintf(insertExecutionQueryStr, "project1", "domain1", "name2", "SUCCEEDED", 1, 2),

		// And one with a different launch plan
		fmt.Sprintf(insertExecutionQueryStr, "project1", "domain1", "name3", "RUNNING", 3, 2),

		// And another with a different workflow
		fmt.Sprintf(insertExecutionQueryStr, "project1", "domain1", "name4", "FAILED", 1, 4),

		// And a few in a whole different project + domain scope
		// (still referencing the same launch plan and workflow just to avoid inserting additional values in the db).
		fmt.Sprintf(insertExecutionQueryStr, "project1", "domain2", "name1", "RUNNING", 1, 2),
		fmt.Sprintf(insertExecutionQueryStr, "project2", "domain2", "name1", "SUCCEEDED", 1, 2),
	}
	db, err := repositories.GetDB(context.Background(), getDbConfig(), getLoggerConfig())
	ctx := context.Background()
	if err != nil {
		logger.Fatal(ctx, "Failed to open DB connection due to %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal(ctx, err)
	}

	defer func(deferCtx context.Context) {
		if err = sqlDB.Close(); err != nil {
			logger.Fatal(deferCtx, err)
		}
	}(ctx)

	// Insert dummy launch plans;
	db.Exec(`INSERT INTO launch_plans ("id", "project", "domain", "name", "version", "spec", "closure") ` +
		`VALUES (1, 'project1', 'domain1', 'name1', 'version1', E'\\000', E'\\000')`)
	db.Exec(`INSERT INTO launch_plans ("id", "project", "domain", "name", "version", "spec", "closure") ` +
		`VALUES (3, 'project2', 'domain2', 'name2', 'version1', E'\\000', E'\\000')`)
	// And dummy workflows:
	db.Exec(`INSERT INTO workflows ("id", "project", "domain", "name", "version", "remote_closure_identifier") ` +
		`VALUES (2, 'project1', 'domain1', 'name1', 'version1', 's3://foo')`)
	db.Exec(`INSERT INTO workflows ("id", "project", "domain", "name", "version", "remote_closure_identifier") ` +
		`VALUES (4, 'project2', 'domain2', 'name2', 'version1', 's3://foo')`)

	// Insert dummy tags
	db.Exec(`INSERT INTO admin_tags ("id", "name") ` + `VALUES (1, 'hello')`)
	db.Exec(`INSERT INTO admin_tags ("id", "name") ` + `VALUES (2, 'flyte')`)
	db.Exec(`INSERT INTO execution_admin_tags ("execution_project", "execution_domain", "execution_name", "admin_tag_id") ` + `VALUES ('project1', 'domain1', 'name1', 1)`)
	db.Exec(`INSERT INTO execution_admin_tags ("execution_project", "execution_domain", "execution_name", "admin_tag_id") ` + `VALUES ('project1', 'domain1', 'name1', 2)`)
	db.Exec(`INSERT INTO execution_admin_tags ("execution_project", "execution_domain", "execution_name", "admin_tag_id") ` + `VALUES ('project1', 'domain1', 'name3', 2)`)
	db.Exec(`INSERT INTO execution_admin_tags ("execution_project", "execution_domain", "execution_name", "admin_tag_id") ` + `VALUES ('project1', 'domain1', 'name4', 1)`)

	for _, statement := range insertExecutionStatements {
		db.Exec(statement)
	}
}

func TestListWorkflowExecutions(t *testing.T) {
	truncateAllTablesForTestingOnly()
	populateWorkflowExecutionsForTestingOnly()

	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	resp, err := client.ListExecutions(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project1",
			Domain:  "domain1",
		},
		Limit: 5,
	})
	assert.Nil(t, err)
	assert.Equal(t, len(resp.Executions), 4)
}

func TestListWorkflowExecutionsWithTags(t *testing.T) {
	truncateAllTablesForTestingOnly()
	populateWorkflowExecutionsForTestingOnly()

	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	resp, err := client.ListExecutions(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project1",
			Domain:  "domain1",
		},
		Limit:   5,
		Filters: "value_in(admin_tag.name, hello)",
	})
	assert.Nil(t, err)
	assert.Equal(t, len(resp.Executions), 2)
}

func TestListWorkflowExecutions_Filters(t *testing.T) {
	truncateAllTablesForTestingOnly()
	populateWorkflowExecutionsForTestingOnly()

	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	resp, err := client.ListExecutions(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project1",
			Domain:  "domain1",
		},
		Limit:   5,
		Filters: "eq(phase,RUNNING)", // 1 corresponds to running.
	})
	assert.Nil(t, err)
	assert.Len(t, resp.Executions, 2)

	resp, err = client.ListExecutions(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project1",
			Domain:  "domain1",
		},
		Limit:   5,
		Filters: "eq(phase,RUNNING)+ne(launch_plan.project, project2)",
	})
	assert.Nil(t, err)
	assert.Len(t, resp.Executions, 1)

	resp, err = client.ListExecutions(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project1",
			Domain:  "domain1",
		},
		Limit:   5,
		Filters: "eq(phase,FAILED)+eq(workflow.domain, domain2)",
	})
	assert.Nil(t, err)
	assert.Len(t, resp.Executions, 1)

	resp, err = client.ListExecutions(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project1",
			Domain:  "domain1",
		},
		Limit:   5,
		Filters: "contains(workflow.name, 1)",
	})
	assert.Nil(t, err)
	assert.Len(t, resp.Executions, 3)
}

func TestListWorkflowExecutions_Pagination(t *testing.T) {
	truncateAllTablesForTestingOnly()
	populateWorkflowExecutionsForTestingOnly()

	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	resp, err := client.ListExecutions(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project1",
			Domain:  "domain1",
		},
		Limit: 3,
	})
	assert.Nil(t, err)
	assert.Equal(t, len(resp.Executions), 3)
	assert.NotEmpty(t, resp.Token)

	resp, err = client.ListExecutions(ctx, &admin.ResourceListRequest{
		Id: &admin.NamedEntityIdentifier{
			Project: "project1",
			Domain:  "domain1",
		},
		Limit: 3,
		Token: resp.Token,
	})
	assert.Nil(t, err)
	assert.Equal(t, len(resp.Executions), 1)
	assert.Empty(t, resp.Token)
}
