// Util around parsing request filters
package util

import (
	"fmt"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"regexp"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
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

var allowedJoinEntities = map[common.Entity]sets.Set[common.Entity]{
	common.Execution:           sets.New[common.Entity](common.Execution, common.LaunchPlan, common.Workflow, common.Task, common.AdminTag),
	common.LaunchPlan:          sets.New[common.Entity](common.LaunchPlan, common.Workflow),
	common.NodeExecution:       sets.New[common.Entity](common.NodeExecution, common.Execution),
	common.NodeExecutionEvent:  sets.New[common.Entity](common.NodeExecutionEvent),
	common.Task:                sets.New[common.Entity](common.Task),
	common.TaskExecution:       sets.New[common.Entity](common.TaskExecution, common.Task, common.Execution, common.NodeExecution),
	common.Workflow:            sets.New[common.Entity](common.Workflow),
	common.NamedEntity:         sets.New[common.Entity](common.NamedEntity, common.NamedEntityMetadata),
	common.NamedEntityMetadata: sets.New[common.Entity](common.NamedEntityMetadata),
	common.Project:             sets.New[common.Entity](common.Project),
	common.Signal:              sets.New[common.Entity](common.Signal),
	common.AdminTag:            sets.New[common.Entity](common.AdminTag),
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

func GetRequestFilters(requestFilters string, primaryEntity common.Entity) (
	filters []common.InlineFilter, err error) {
	if requestFilters == "" {
		return filters, nil
	}
	filters, err = ParseFilters(requestFilters, primaryEntity)
	if err != nil {
		return nil, err
	}
	return filters, nil
}

type NamedEntityLike interface {
	GetProject() string
	GetDomain() string
	GetName() string
	GetOrg() string
}

func GetIdentifierScope(id NamedEntityLike) *admin.NamedEntityIdentifier {
	return &admin.NamedEntityIdentifier{
		Project: id.GetProject(),
		Domain:  id.GetDomain(),
		Name:    id.GetName(),
		Org:     id.GetOrg(),
	}
}
