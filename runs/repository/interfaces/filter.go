package interfaces

type FilterExpression = int

// Set of filters available for database queries.
const (
	FilterExpressionEqual FilterExpression = iota
	FilterExpressionNotEqual
	FilterExpressionGreaterThan
	FilterExpressionGreaterThanOrEqual
	FilterExpressionLessThan
	FilterExpressionLessThanOrEqual
	FilterExpressionContains
	FilterExpressionValueIn
	FilterExpressionEndsWith
	FilterExpressionNotEndsWith
	FilterExpressionContainsCaseInsensitive
)

// QueryExpr is a container for arguments necessary to issue a query.
type QueryExpr struct {
	Query string
	Args  []interface{}
}

type Filter interface {
	QueryExpression(table string) (QueryExpr, error)
	And(filter Filter) Filter
	Or(filter Filter) Filter
}
