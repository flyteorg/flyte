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

// GormQueryExpr is a container for arguments necessary to issue a GORM query.
type GormQueryExpr struct {
	Query string
	Args  []interface{}
}

type Filter interface {
	GormQueryExpression(table string) (GormQueryExpr, error)
	And(filter Filter) Filter
	Or(filter Filter) Filter
}
