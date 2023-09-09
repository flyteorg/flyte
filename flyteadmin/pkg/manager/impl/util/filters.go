// Util around parsing request filters
package util

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"google.golang.org/grpc/codes"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

const (
	filterExpressionSeperator = "+"
	listValueSeparator        = ";"
)

// Matches filters of the form `func(field,value)` or `func(field, value)`
var filterRegex = regexp.MustCompile(`(.+)\((.+),\s?(.+)\)`)

// InlineFilter parsing consts. For example, matching on the filter string "contains(Name, foo)"
// will return a slice consisting of: ["contains(Name, foo)", "contains", "Name", "foo"]
const (
	funcMatchIndex           = 1
	fieldMatchIndex          = 2
	valueMatchIndex          = 3
	expectedMatchGroupLength = 4
)

var timestampFields = map[string]bool{
	"CreatedAt": true,
	"UpdatedAt": true,
	"DeletedAt": true,
	"StartedAt": true,
}

var durationFields = map[string]bool{
	"duration": true,
}

const filterFieldEntityPrefixFmt = "%s."
const secondsFormat = "%vs"

var filterFieldEntityPrefix = map[string]common.Entity{
	"task":                  common.Task,
	"workflow":              common.Workflow,
	"launch_plan":           common.LaunchPlan,
	"execution":             common.Execution,
	"node_execution":        common.NodeExecution,
	"task_execution":        common.TaskExecution,
	"entities":              common.NamedEntity,
	"named_entity_metadata": common.NamedEntityMetadata,
	"project":               common.Project,
	"signal":                common.Signal,
	"admin_tag":             common.AdminTag,
	"execution_admin_tag":   common.ExecutionAdminTag,
}

func parseField(field string, primaryEntity common.Entity) (common.Entity, string) {
	for prefix, entity := range filterFieldEntityPrefix {
		otherEntityPrefix := fmt.Sprintf(filterFieldEntityPrefixFmt, prefix)
		if strings.HasPrefix(field, otherEntityPrefix) {
			// Strip the referenced entity prefix from the field name.
			// e.g. workflow_name becomes simply "name"
			return entity, field[len(otherEntityPrefix):]
		}
	}

	return primaryEntity, field
}

func parseRepeatedValues(parsedValues string) []string {
	return strings.Split(parsedValues, listValueSeparator)
}

// Handles parsing repeated values and non-string values such as time fields.
func prepareValues(field string, values []string) (interface{}, error) {
	preparedValues := make([]interface{}, len(values))
	if isTimestampField := timestampFields[field]; isTimestampField {
		for idx, value := range values {
			timestamp, err := time.Parse(time.RFC3339Nano, value)
			if err != nil {
				return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"Timestamp %s must conform to RFC3339 Nano spec", value)
			}
			preparedValues[idx] = timestamp
		}
	} else if isDurationField := durationFields[strings.ToLower(field)]; isDurationField {
		for idx, value := range values {
			floatValue, err := strconv.ParseFloat(value, 64)
			if err == nil {
				// The value is an float. By default purely float values are assumed to represent durations in seconds.
				value = fmt.Sprintf(secondsFormat, floatValue)
			}
			duration, err := time.ParseDuration(value)
			if err != nil {
				return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"Failed to parse duration [%s]", value)
			}
			preparedValues[idx] = duration
		}
	} else {
		for idx, value := range values {
			preparedValues[idx] = value
		}
	}
	if len(preparedValues) == 1 {
		return preparedValues[0], nil
	}
	return preparedValues, nil
}

var allowedJoinEntities = map[common.Entity]sets.String{
	common.Execution:           sets.NewString(common.Execution, common.LaunchPlan, common.Workflow, common.Task, common.AdminTag),
	common.LaunchPlan:          sets.NewString(common.LaunchPlan, common.Workflow),
	common.NodeExecution:       sets.NewString(common.NodeExecution, common.Execution),
	common.NodeExecutionEvent:  sets.NewString(common.NodeExecutionEvent),
	common.Task:                sets.NewString(common.Task),
	common.TaskExecution:       sets.NewString(common.TaskExecution, common.Task, common.Execution, common.NodeExecution),
	common.Workflow:            sets.NewString(common.Workflow),
	common.NamedEntity:         sets.NewString(common.NamedEntity, common.NamedEntityMetadata),
	common.NamedEntityMetadata: sets.NewString(common.NamedEntityMetadata),
	common.Project:             sets.NewString(common.Project),
	common.Signal:              sets.NewString(common.Signal),
	common.AdminTag:            sets.NewString(common.AdminTag),
}

var entityColumns = map[common.Entity]sets.String{
	common.Execution:           models.ExecutionColumns,
	common.LaunchPlan:          models.LaunchPlanColumns,
	common.NodeExecution:       models.NodeExecutionColumns,
	common.NodeExecutionEvent:  models.NodeExecutionEventColumns,
	common.Task:                models.TaskColumns,
	common.TaskExecution:       models.TaskExecutionColumns,
	common.Workflow:            models.WorkflowColumns,
	common.NamedEntity:         models.NamedEntityColumns,
	common.NamedEntityMetadata: models.NamedEntityMetadataColumns,
	common.Project:             models.ProjectColumns,
	common.Signal:              models.SignalColumns,
	common.AdminTag:            models.AdminTagColumns,
}

func ParseFilters(filterParams string, primaryEntity common.Entity) ([]common.InlineFilter, error) {
	// Multiple filters can be appended as URI-escaped strings joined by filterExpressionSeperator
	filterExpressions := strings.Split(filterParams, filterExpressionSeperator)
	parsedFilters := make([]common.InlineFilter, 0)
	for _, filterExpression := range filterExpressions {
		// Parse string expression
		matches := filterRegex.FindStringSubmatch(filterExpression)
		if len(matches) != expectedMatchGroupLength {
			// Poorly formatted filter string doesn't match expected regex.
			return nil, shared.GetInvalidArgumentError(shared.Filters)
		}
		referencedEntity, field := parseField(matches[fieldMatchIndex], primaryEntity)

		joinEntities, ok := allowedJoinEntities[primaryEntity]
		if !ok {
			return nil, fmt.Errorf("unsupported entity '%s'", primaryEntity)
		}

		if !joinEntities.Has(referencedEntity) {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "'%s' entity is not allowed in filters", referencedEntity)
		}

		if !entityColumns[referencedEntity].Has(field) {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "'%s.%s' is invalid filter", referencedEntity, field)
		}

		// Parse and transform values
		parsedValues := parseRepeatedValues(matches[valueMatchIndex])
		preparedValues, err := prepareValues(field, parsedValues)
		if err != nil {
			return nil, err
		}
		// Create InlineFilter object.
		filter, err := common.NewInlineFilter(referencedEntity, matches[funcMatchIndex], field, preparedValues)
		if err != nil {
			return nil, err
		}
		parsedFilters = append(parsedFilters, filter)
	}
	return parsedFilters, nil
}

func GetSingleValueEqualityFilter(entity common.Entity, field, value string) (common.InlineFilter, error) {
	return common.NewSingleValueFilter(entity, common.Equal, field, value)
}

type FilterSpec struct {
	// All of these fields are optional (although they should not *all* be empty).
	Project        string
	Domain         string
	Name           string
	RequestFilters string
}

// Returns equality filters initialized for identifier attributes (project, domain & name)
// which can be optionally specified in requests.
func getIdentifierFilters(entity common.Entity, spec FilterSpec) ([]common.InlineFilter, error) {
	filters := make([]common.InlineFilter, 0)
	if spec.Project != "" {
		projectFilter, err := GetSingleValueEqualityFilter(entity, shared.Project, spec.Project)
		if err != nil {
			return nil, err
		}
		filters = append(filters, projectFilter)
	}
	if spec.Domain != "" {
		domainFilter, err := GetSingleValueEqualityFilter(entity, shared.Domain, spec.Domain)
		if err != nil {
			return nil, err
		}
		filters = append(filters, domainFilter)
	}

	if spec.Name != "" {
		nameFilter, err := GetSingleValueEqualityFilter(entity, shared.Name, spec.Name)
		if err != nil {
			return nil, err
		}
		filters = append(filters, nameFilter)
	}
	return filters, nil
}

func AddRequestFilters(requestFilters string, primaryEntity common.Entity, existingFilters []common.InlineFilter) (
	[]common.InlineFilter, error) {

	if requestFilters == "" {
		return existingFilters, nil
	}
	var additionalFilters []common.InlineFilter
	additionalFilters, err := ParseFilters(requestFilters, primaryEntity)
	if err != nil {
		return nil, err
	}
	updatedFilters := append(existingFilters, additionalFilters...)
	return updatedFilters, nil
}

// Consolidates request params and filters to a single list of filters. This consolidation is necessary since the db is
// agnostic to required request parameters and additional filter arguments.
func GetDbFilters(spec FilterSpec, primaryEntity common.Entity) ([]common.InlineFilter, error) {
	filters, err := getIdentifierFilters(primaryEntity, spec)
	if err != nil {
		return nil, err
	}

	// Append any request filters.
	if spec.RequestFilters != "" {
		filters, err = AddRequestFilters(spec.RequestFilters, primaryEntity, filters)
		if err != nil {
			return nil, err
		}
	}

	return filters, nil
}

func GetWorkflowExecutionIdentifierFilters(
	ctx context.Context, workflowExecutionIdentifier core.WorkflowExecutionIdentifier) ([]common.InlineFilter, error) {
	identifierFilters := make([]common.InlineFilter, 3)
	identifierProjectFilter, err := GetSingleValueEqualityFilter(
		common.Execution, shared.Project, workflowExecutionIdentifier.Project)
	if err != nil {
		logger.Warningf(ctx, "Failed to create execution identifier filter for project: %s with identifier [%+v]",
			workflowExecutionIdentifier.Project, workflowExecutionIdentifier)
		return nil, err
	}
	identifierFilters[0] = identifierProjectFilter

	identifierDomainFilter, err := GetSingleValueEqualityFilter(
		common.Execution, shared.Domain, workflowExecutionIdentifier.Domain)
	if err != nil {
		logger.Warningf(ctx, "Failed to create execution identifier filter for domain: %s with identifier [%+v]",
			workflowExecutionIdentifier.Domain, workflowExecutionIdentifier)
		return nil, err
	}
	identifierFilters[1] = identifierDomainFilter

	identifierNameFilter, err := GetSingleValueEqualityFilter(
		common.Execution, shared.Name, workflowExecutionIdentifier.Name)
	if err != nil {
		logger.Warningf(ctx, "Failed to create execution identifier filter for domain: %s with identifier [%+v]",
			workflowExecutionIdentifier.Name, workflowExecutionIdentifier)
		return nil, err
	}
	identifierFilters[2] = identifierNameFilter
	return identifierFilters, nil
}

// All inputs to this function must be validated.
func GetNodeExecutionIdentifierFilters(
	ctx context.Context, nodeExecutionIdentifier core.NodeExecutionIdentifier) ([]common.InlineFilter, error) {
	workflowExecutionIdentifierFilters, err :=
		GetWorkflowExecutionIdentifierFilters(ctx, *nodeExecutionIdentifier.ExecutionId)
	if err != nil {
		return nil, err
	}
	nodeIDFilter, err := GetSingleValueEqualityFilter(
		common.NodeExecution, shared.NodeID, nodeExecutionIdentifier.NodeId)
	if err != nil {
		logger.Warningf(ctx, "Failed to create node execution identifier filter for node id: %s with identifier [%+v]",
			nodeExecutionIdentifier.NodeId, nodeExecutionIdentifier)
	}
	return append(workflowExecutionIdentifierFilters, nodeIDFilter), nil
}
