package gormimpl

import (
	"fmt"

	"github.com/lyft/datacatalog/pkg/common"
	"github.com/lyft/datacatalog/pkg/repositories/errors"
	"github.com/lyft/datacatalog/pkg/repositories/models"
)

// String formats for various GORM expression queries
const (
	equalQuery = "%s.%s = ?"
)

type gormValueFilterImpl struct {
	comparisonOperator common.ComparisonOperator
	field              string
	value              interface{}
}

// Get the GORM expression to filter by a model's property. The output should be a valid input into tx.Where()
func (g *gormValueFilterImpl) GetDBQueryExpression(tableName string) (models.DBQueryExpr, error) {
	switch g.comparisonOperator {
	case common.Equal:
		return models.DBQueryExpr{
			Query: fmt.Sprintf(equalQuery, tableName, g.field),
			Args:  g.value,
		}, nil
	}
	return models.DBQueryExpr{}, errors.GetUnsupportedFilterExpressionErr(g.comparisonOperator)
}

// Construct the container necessary to issue a db query to filter in GORM
func NewGormValueFilter(comparisonOperator common.ComparisonOperator, field string, value interface{}) models.ModelValueFilter {
	return &gormValueFilterImpl{
		comparisonOperator: comparisonOperator,
		field:              field,
		value:              value,
	}
}
