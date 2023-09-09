package gormimpl

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"gorm.io/gorm"

	"github.com/flyteorg/flyteadmin/pkg/common"
	adminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
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
const AdminTagsTableName = "admin_tags"
const executionAdminTagsTableName = "execution_admin_tags"

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
}

var innerJoinExecToNodeExec = fmt.Sprintf(
	"INNER JOIN %s ON %s.execution_project = %s.execution_project AND "+
		"%s.execution_domain = %s.execution_domain AND %s.execution_name = %s.execution_name",
	executionTableName, nodeExecutionTableName, executionTableName, nodeExecutionTableName, executionTableName,
	nodeExecutionTableName, executionTableName)

var innerJoinNodeExecToTaskExec = fmt.Sprintf(
	"INNER JOIN %s ON %s.node_id = %s.node_id AND %s.execution_project = %s.execution_project AND "+
		"%s.execution_domain = %s.execution_domain AND %s.execution_name = %s.execution_name",
	nodeExecutionTableName, taskExecutionTableName, nodeExecutionTableName, taskExecutionTableName,
	nodeExecutionTableName, taskExecutionTableName, nodeExecutionTableName, taskExecutionTableName,
	nodeExecutionTableName)

// Because dynamic tasks do NOT necessarily register static task definitions, we use a left join to not exclude
// dynamic tasks from list queries.
var leftJoinTaskToTaskExec = fmt.Sprintf(
	"LEFT JOIN %s ON %s.project = %s.project AND %s.domain = %s.domain AND %s.name = %s.name AND "+
		"%s.version = %s.version",
	taskTableName, taskExecutionTableName, taskTableName, taskExecutionTableName, taskTableName,
	taskExecutionTableName, taskTableName, taskExecutionTableName, taskTableName)

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

func applyFilters(tx *gorm.DB, inlineFilters []common.InlineFilter, mapFilters []common.MapFilter) (*gorm.DB, error) {
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
	return tx, nil
}

func applyScopedFilters(tx *gorm.DB, inlineFilters []common.InlineFilter, mapFilters []common.MapFilter) (*gorm.DB, error) {
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
	return tx, nil
}
