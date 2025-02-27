package gormimpl

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	adminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
)

const Project = "project"
const Domain = "domain"
const Name = "name"
const Version = "version"
const Description = "description"
const ResourceType = "resource_type"
const State = "state"
const ID = "id"

const executionTableName = "executions"
const namedEntityMetadataTableName = "named_entity_metadata"
const nodeExecutionTableName = "node_executions"
const taskExecutionTableName = "task_executions"
const taskTableName = "tasks"
const workflowTableName = "workflows"
const descriptionEntityTableName = "description_entities"
const executionTagsTableName = "execution_tags"

const limit = "limit"
const filters = "filters"

var identifierGroupBy = fmt.Sprintf("%s, %s, %s", Project, Domain, Name)

var entityToTableName = map[common.Entity]string{
	common.Execution:           "executions",
	common.LaunchPlan:          "launch_plans",
	common.NodeExecution:       "node_executions",
	common.NodeExecutionEvent:  "node_execution_events",
	common.Task:                "tasks",
	common.TaskExecution:       "task_executions",
	common.Workflow:            "workflows",
	common.NamedEntity:         "entities",
	common.NamedEntityMetadata: "named_entity_metadata",
	common.Signal:              "signals",
	common.AdminTag:            "admin_tags",
	common.ExecutionAdminTag:   "execution_admin_tags",
	common.ExecutionTag:        "execution_tags",
}

var innerJoinExecToNodeExec = fmt.Sprintf(
	"INNER JOIN %[1]s ON %[2]s.execution_project = %[1]s.execution_project AND %[2]s.execution_domain = %[1]s.execution_domain AND %[2]s.execution_name = %[1]s.execution_name",
	executionTableName, nodeExecutionTableName)
var innerJoinExecToTaskExec = fmt.Sprintf(
	"INNER JOIN %[1]s ON %[2]s.execution_project = %[1]s.execution_project AND %[2]s.execution_domain = %[1]s.execution_domain AND %[2]s.execution_name = %[1]s.execution_name",
	executionTableName, taskExecutionTableName)

var innerJoinNodeExecToTaskExec = fmt.Sprintf(
	"INNER JOIN %[1]s ON %s.node_id = %[1]s.node_id AND %[2]s.execution_project = %[1]s.execution_project AND %[2]s.execution_domain = %[1]s.execution_domain AND %[2]s.execution_name = %[1]s.execution_name",
	nodeExecutionTableName, taskExecutionTableName)

// Because dynamic tasks do NOT necessarily register static task definitions, we use a left join to not exclude
// dynamic tasks from list queries.
var leftJoinTaskToTaskExec = fmt.Sprintf(
	"LEFT JOIN %[1]s ON %[2]s.project = %[1]s.project AND %[2]s.domain = %[1]s.domain AND %[2]s.name = %[1]s.name AND "+
		" %[2]s.version = %[1]s.version",
	taskTableName, taskExecutionTableName)

// Validates there are no missing but required parameters in ListResourceInput
func ValidateListInput(input interfaces.ListResourceInput) adminErrors.FlyteAdminError {
	if input.Limit == 0 {
		return errors.GetInvalidInputError(limit)
	}
	if len(input.InlineFilters) == 0 {
		return errors.GetInvalidInputError(filters)
	}
	return nil
}

func applyFilters(tx *gorm.DB, inlineFilters []common.InlineFilter, mapFilters []common.MapFilter, isolationFilter common.IsolationFilter) (*gorm.DB, error) {
	for _, filter := range inlineFilters {
		gormQueryExpr, err := filter.GetGormQueryExpr()
		if err != nil {
			return nil, errors.GetInvalidInputError(err.Error())
		}
		tx = tx.Where(gormQueryExpr.Query, gormQueryExpr.Args)
	}
	for _, mapFilter := range mapFilters {
		tx = tx.Where(mapFilter.GetFilter())
	}
	if isolationFilter != nil {
		tx = tx.Where(tx.Scopes(isolationFilter.GetScopes()...))
	}
	return tx, nil
}

func applyScopedFilters(tx *gorm.DB, inlineFilters []common.InlineFilter, mapFilters []common.MapFilter, isolationFilter common.IsolationFilter) (*gorm.DB, error) {
	for _, filter := range inlineFilters {
		tableName, ok := entityToTableName[filter.GetEntity()]
		if !ok {
			return nil, adminErrors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"unrecognized entity in filter expression: %v", filter.GetEntity())
		}
		gormQueryExpr, err := filter.GetGormJoinTableQueryExpr(tableName)
		if err != nil {
			return nil, err
		}
		tx = tx.Where(gormQueryExpr.Query, gormQueryExpr.Args)
	}
	for _, mapFilter := range mapFilters {
		tx = tx.Where(mapFilter.GetFilter())
	}
	if isolationFilter != nil {
		tx = tx.Where(tx.Scopes(isolationFilter.GetScopes()...))
	}
	return tx, nil
}

func getIDFilter(id uint) (query string, args interface{}) {
	return fmt.Sprintf("%s = ?", ID), id
}
