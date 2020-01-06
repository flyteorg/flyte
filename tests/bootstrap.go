// +build integration

package tests

import (
	"fmt"

	"github.com/jinzhu/gorm"

	database_config "github.com/lyft/flyteadmin/pkg/repositories/config"
	"github.com/lyft/flytestdlib/promutils"
)

const insertExecutionQueryStr = `INSERT INTO "executions" ` +
	`("execution_project","execution_domain","execution_name","phase","launch_plan_id","workflow_id") ` +
	`VALUES ('%s', '%s', '%s', '%s', '%d', '%d')`

var adminScope = promutils.NewScope("flyteadmin")

func getDbConfig() database_config.DbConfig {
	return database_config.DbConfig{
		Host:   "postgres",
		Port:   5432,
		DbName: "postgres",
		User:   "postgres",
	}
}

// Use this for running integration tests in tandem with flyteadmin_config.yaml
func getLocalDbConfig() database_config.DbConfig {
	return database_config.DbConfig{
		Host:   "localhost",
		Port:   5432,
		DbName: "postgres",
		User:   "postgres",
	}
}

func truncateTableForTesting(db *gorm.DB, tableName string) {
	db.Exec(fmt.Sprintf("TRUNCATE TABLE %s;", tableName))
}

func truncateAllTablesForTestingOnly() {
	// Load the running configuration in order to talk to the running flyteadmin instance
	fmt.Println("Truncating tables")
	TruncateTasks := fmt.Sprintf("TRUNCATE TABLE tasks;")
	TruncateWorkflows := fmt.Sprintf("TRUNCATE TABLE workflows;")
	TruncateLaunchPlans := fmt.Sprintf("TRUNCATE TABLE launch_plans;")
	// HACK: alter executions table so that spec is not required for testing.
	TruncateExecutions := fmt.Sprintf("TRUNCATE TABLE executions; alter table executions alter column spec drop not null;")
	TruncateExecutionEvents := fmt.Sprintf("TRUNCATE TABLE execution_events;")
	TruncateNamedEntityMetadata := fmt.Sprintf("TRUNCATE TABLE named_entity_metadata;")
	TruncateNodeExecutions := fmt.Sprintf("TRUNCATE TABLE node_executions;")
	TruncateNodeExecutionEvents := fmt.Sprintf("TRUNCATE TABLE node_execution_events;")
	TruncateTaskExecutions := fmt.Sprintf("TRUNCATE TABLE task_executions;")
	db := database_config.OpenDbConnection(database_config.NewPostgresConfigProvider(getDbConfig(), adminScope))
	defer db.Close()
	db.Exec(TruncateTasks)
	db.Exec(TruncateWorkflows)
	db.Exec(TruncateLaunchPlans)
	db.Exec(TruncateExecutions)
	db.Exec(TruncateExecutionEvents)
	db.Exec(TruncateNamedEntityMetadata)
	db.Exec(TruncateNodeExecutions)
	db.Exec(TruncateNodeExecutionEvents)
	db.Exec(TruncateTaskExecutions)
}

func populateWorkflowExecutionForTestingOnly(project, domain, name string) {
	InsertExecution := fmt.Sprintf(insertExecutionQueryStr, project, domain, name, "UNDEFINED", 1, 2)
	db := database_config.OpenDbConnection(database_config.NewPostgresConfigProvider(getDbConfig(), adminScope))
	defer db.Close()
	db.Exec(InsertExecution)
}
