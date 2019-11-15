package models

import "github.com/lyft/datacatalog/pkg/common"

// Inputs to specify to list models
type ListModelsInput struct {
	// JoinEntityToConditionMap for the list query. It is represented as a 1:1 mapping between a joiningEntity and a joinCondition
	JoinEntityToConditionMap map[common.Entity]ModelJoinCondition
	// Value filters for the list query
	Filters []ModelValueFilter
	// The number of models to list
	Limit uint32
	// The token to offset results by
	Offset uint32
	// Parameter to sort by
	SortParameter SortParameter
}

type SortParameter interface {
	GetDBOrderExpression(tableName string) string
}

// Generates db filter expressions for model values
type ModelValueFilter interface {
	GetDBEntity() common.Entity
	GetDBQueryExpression(tableName string) (DBQueryExpr, error)
}

type ModelJoinCondition interface {
	GetJoinOnDBQueryExpression(sourceTableName string, joiningTableName string) (string, error)
}

// Encapsulates the query and necessary arguments to issue a DB query.
type DBQueryExpr struct {
	Query string
	Args  interface{}
}
