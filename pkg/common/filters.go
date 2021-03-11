// Defines a common set of filters that can be applied in queries for supported databases types.
package common

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flytestdlib/logger"
	"google.golang.org/grpc/codes"
)

type FilterExpression int

// Container for arguments necessary to issue a GORM query.
type GormQueryExpr struct {
	Query string
	Args  interface{}
}

// Complete set of filters available for database queries.
const (
	Contains FilterExpression = iota
	GreaterThan
	GreaterThanOrEqual
	LessThan
	LessThanOrEqual
	Equal
	NotEqual
	ValueIn
)

// String formats for various filter expression queries
const (
	joinArgsFormat          = "%s.%s"
	containsQuery           = "%s LIKE ?"
	containsArgs            = "%%%s%%"
	greaterThanQuery        = "%s > ?"
	greaterThanOrEqualQuery = "%s >= ?"
	lessThanQuery           = "%s < ?"
	lessThanOrEqualQuery    = "%s <= ?"
	equalQuery              = "%s = ?"
	notEqualQuery           = "%s <> ?"
	valueInQuery            = "%s in (?)"
)

// Set of available filters which exclusively accept a single argument value.
var singleValueFilters = map[FilterExpression]bool{
	Contains:           true,
	GreaterThan:        true,
	GreaterThanOrEqual: true,
	LessThan:           true,
	LessThanOrEqual:    true,
	Equal:              true,
	NotEqual:           true,
}

// Set of available filters which exclusively accept repeated argument values.
var repeatedValueFilters = map[FilterExpression]bool{
	ValueIn: true,
}

const EqualExpression = "eq"

var filterNameMappings = map[string]FilterExpression{
	"contains":      Contains,
	"gt":            GreaterThan,
	"gte":           GreaterThanOrEqual,
	"lt":            LessThan,
	"lte":           LessThanOrEqual,
	EqualExpression: Equal,
	"ne":            NotEqual,
	"value_in":      ValueIn,
}

var executionIdentifierFields = map[string]bool{
	"project": true,
	"domain":  true,
	"name":    true,
}

var entityMetadataFields = map[string]bool{
	"description": true,
	"state":       true,
}

const unrecognizedFilterFunction = "unrecognized filter function: %s"
const unsupportedFilterExpression = "unsupported filter expression: %s"
const invalidSingleValueFilter = "invalid single value filter expression: %s"
const invalidRepeatedValueFilter = "invalid repeated value filter expression: %s"

func getFilterExpressionName(expression FilterExpression) string {
	switch expression {
	case Contains:
		return "contains"
	case GreaterThan:
		return "greater than"
	case GreaterThanOrEqual:
		return "greater than or equal"
	case LessThan:
		return "less than"
	case LessThanOrEqual:
		return "less than or equal"
	case Equal:
		return "equal"
	case NotEqual:
		return "not equal"
	case ValueIn:
		return "value in"
	default:
		return ""
	}
}

func GetUnrecognizedFilterFunctionErr(function string) error {
	return errors.NewFlyteAdminErrorf(codes.InvalidArgument, unrecognizedFilterFunction,
		function)
}

func GetUnsupportedFilterExpressionErr(expression FilterExpression) error {
	return errors.NewFlyteAdminErrorf(codes.InvalidArgument, unsupportedFilterExpression,
		getFilterExpressionName(expression))
}

func GetInvalidSingleValueFilterErr(expression FilterExpression) error {
	return errors.NewFlyteAdminErrorf(codes.InvalidArgument, invalidSingleValueFilter,
		getFilterExpressionName(expression))
}

func GetInvalidRepeatedValueFilterErr(expression FilterExpression) error {
	return errors.NewFlyteAdminErrorf(codes.InvalidArgument, invalidRepeatedValueFilter,
		getFilterExpressionName(expression))
}

// Interface for a single filter expression.
type InlineFilter interface {
	// Returns the entity for which this filter should be applied.
	GetEntity() Entity
	// Returns the column filtered on.
	GetField() string
	// Generates fields necessary to add a filter to a gorm database query.
	GetGormQueryExpr() (GormQueryExpr, error)
	// Generates fields necessary to add a filter on a gorm database join query.
	GetGormJoinTableQueryExpr(tableName string) (GormQueryExpr, error)

	// Include other db-specific methods here.
}

// FilterInterface implementation. Only one of value or repeatedValue should ever be populated (based on the function).
type inlineFilterImpl struct {
	entity        Entity
	function      FilterExpression
	field         string
	value         interface{}
	repeatedValue interface{}
}

func (f *inlineFilterImpl) GetEntity() Entity {
	return f.entity
}

func (f *inlineFilterImpl) GetField() string {
	return f.field
}

func (f *inlineFilterImpl) getGormQueryExpr(formattedField string) (GormQueryExpr, error) {

	// ValueIn is special because it uses repeating values.
	if f.function == ValueIn {
		queryStr := fmt.Sprintf(valueInQuery, formattedField)
		return GormQueryExpr{
			Query: queryStr,
			Args:  f.repeatedValue,
		}, nil
	}

	switch f.function {
	case Contains:
		return GormQueryExpr{
			// WHERE field LIKE %value%
			Query: fmt.Sprintf(containsQuery, formattedField),
			// args renders to something like: "%value%"
			Args: fmt.Sprintf(containsArgs, f.value),
		}, nil
	case GreaterThan:
		return GormQueryExpr{
			// WHERE field > value
			Query: fmt.Sprintf(greaterThanQuery, formattedField),
			Args:  f.value,
		}, nil
	case GreaterThanOrEqual:
		return GormQueryExpr{
			// WHERE field >= value
			Query: fmt.Sprintf(greaterThanOrEqualQuery, formattedField),
			Args:  f.value,
		}, nil
	case LessThan:
		return GormQueryExpr{
			// WHERE field < value
			Query: fmt.Sprintf(lessThanQuery, formattedField),
			Args:  f.value,
		}, nil
	case LessThanOrEqual:
		return GormQueryExpr{
			// WHERE field <= value
			Query: fmt.Sprintf(lessThanOrEqualQuery, formattedField),
			Args:  f.value,
		}, nil
	case Equal:
		return GormQueryExpr{
			// WHERE field = value
			Query: fmt.Sprintf(equalQuery, formattedField),
			Args:  f.value,
		}, nil
	case NotEqual:
		return GormQueryExpr{
			// WHERE field <> value
			Query: fmt.Sprintf(notEqualQuery, formattedField),
			Args:  f.value,
		}, nil
	}
	logger.Debugf(context.Background(), "can't create gorm query expr for %s", getFilterExpressionName(f.function))
	return GormQueryExpr{}, GetUnsupportedFilterExpressionErr(f.function)
}

func (f *inlineFilterImpl) GetGormQueryExpr() (GormQueryExpr, error) {
	return f.getGormQueryExpr(f.field)
}

func (f *inlineFilterImpl) GetGormJoinTableQueryExpr(tableName string) (GormQueryExpr, error) {
	formattedField := fmt.Sprintf(joinArgsFormat, tableName, f.field)
	return f.getGormQueryExpr(formattedField)
}

func customizeField(field string, entity Entity) string {
	// Execution identifier fields have to be customized because we differ from convention in those column names.
	if entity == Execution && executionIdentifierFields[field] {
		return fmt.Sprintf("execution_%s", field)
	}
	return field
}

func customizeEntity(field string, entity Entity) Entity {
	// NamedEntity is considered a single object, but the metdata
	// is stored using a different entity type.
	if entity == NamedEntity && entityMetadataFields[field] {
		return NamedEntityMetadata
	}
	return entity
}

// Returns a filter which uses a single argument value.
func NewSingleValueFilter(entity Entity, function FilterExpression, field string, value interface{}) (InlineFilter, error) {
	if _, ok := singleValueFilters[function]; !ok {
		return nil, GetInvalidSingleValueFilterErr(function)
	}
	customizedField := customizeField(field, entity)
	customizedEntity := customizeEntity(field, entity)
	return &inlineFilterImpl{
		entity:   customizedEntity,
		function: function,
		field:    customizedField,
		value:    value,
	}, nil
}

// Returns a filter which uses a repeated argument value.
func NewRepeatedValueFilter(entity Entity, function FilterExpression, field string, repeatedValue interface{}) (InlineFilter, error) {
	if _, ok := repeatedValueFilters[function]; !ok {
		return nil, GetInvalidRepeatedValueFilterErr(function)
	}
	customizedField := customizeField(field, entity)
	return &inlineFilterImpl{
		entity:        entity,
		function:      function,
		field:         customizedField,
		repeatedValue: repeatedValue,
	}, nil
}

func NewInlineFilter(entity Entity, function string, field string, value interface{}) (InlineFilter, error) {
	expression, ok := filterNameMappings[function]
	if !ok {
		logger.Debugf(context.Background(), "can't create filter for unrecognized function: %s", function)
		return nil, GetUnrecognizedFilterFunctionErr(function)
	}
	if isSingleValueFilter := singleValueFilters[expression]; isSingleValueFilter {
		return NewSingleValueFilter(entity, expression, field, value)
	}
	return NewRepeatedValueFilter(entity, expression, field, value)
}

// Interface for a map filter expression.
type MapFilter interface {
	GetFilter() map[string]interface{}
}

type mapFilterImpl struct {
	filter map[string]interface{}
}

func (m mapFilterImpl) GetFilter() map[string]interface{} {
	return m.filter
}

func NewMapFilter(filter map[string]interface{}) MapFilter {
	return &mapFilterImpl{
		filter: filter,
	}
}

const queryWithDefaultFmt = "COALESCE(%s, %v)"

type withDefaultValueFilter struct {
	inlineFilterImpl
	defaultValue interface{}
}

func (f *withDefaultValueFilter) GetGormQueryExpr() (GormQueryExpr, error) {
	formattedField := fmt.Sprintf(queryWithDefaultFmt, f.GetField(), f.defaultValue)
	return f.getGormQueryExpr(formattedField)
}

func (f *withDefaultValueFilter) GetGormJoinTableQueryExpr(tableName string) (GormQueryExpr, error) {
	formattedField := fmt.Sprintf(queryWithDefaultFmt, fmt.Sprintf(joinArgsFormat, tableName, f.GetField()), f.defaultValue)
	return f.getGormQueryExpr(formattedField)
}

func NewWithDefaultValueFilter(defaultValue interface{}, filter InlineFilter) (InlineFilter, error) {
	inlineFilter, ok := filter.(*inlineFilterImpl)
	if !ok {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal,
			"Unable to create default value filter for [%s] because the system encountered an unknown filter type",
			filter.GetField())
	}
	return &withDefaultValueFilter{
		inlineFilterImpl: *inlineFilter,
		defaultValue:     defaultValue,
	}, nil
}
