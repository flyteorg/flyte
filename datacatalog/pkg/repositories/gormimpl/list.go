package gormimpl

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/lyft/datacatalog/pkg/common"
	"github.com/lyft/datacatalog/pkg/repositories/errors"
	"github.com/lyft/datacatalog/pkg/repositories/models"
)

const (
	tableAliasFormat = "%s%d" // Table Alias is the "<table name><index>"
)

var entityToModel = map[common.Entity]interface{}{
	common.Artifact:  models.Artifact{},
	common.Dataset:   models.Dataset{},
	common.Partition: models.Partition{},
	common.Tag:       models.Tag{},
}

// Apply the list query on the source model. This method will apply the necessary joins, filters and
// pagination on the database for the given ListModelInputs.
func applyListModelsInput(tx *gorm.DB, sourceEntity common.Entity, in models.ListModelsInput) (*gorm.DB, error) {
	sourceModel, ok := entityToModel[sourceEntity]
	if !ok {
		return nil, errors.GetInvalidEntityError(sourceEntity)
	}

	sourceTableName := tx.NewScope(sourceModel).TableName()
	for modelIndex, modelFilter := range in.ModelFilters {
		entity := modelFilter.Entity
		filterModel, ok := entityToModel[entity]
		if !ok {
			return nil, errors.GetInvalidEntityError(entity)
		}
		tableName := tx.NewScope(filterModel).TableName()
		tableAlias := tableName

		// Optionally add the join condition if the entity we need isn't the source
		if sourceEntity != modelFilter.Entity {
			// if there is a join associated with the filter, we should use an alias
			joinCondition := modelFilter.JoinCondition
			tableAlias = fmt.Sprintf(tableAliasFormat, tableName, modelIndex)
			joinExpression, err := joinCondition.GetJoinOnDBQueryExpression(sourceTableName, tableName, tableAlias)
			if err != nil {
				return nil, err
			}
			tx = tx.Joins(joinExpression)
		}

		for _, whereFilter := range modelFilter.ValueFilters {
			dbQueryExpr, err := whereFilter.GetDBQueryExpression(tableAlias)

			if err != nil {
				return nil, err
			}
			tx = tx.Where(dbQueryExpr.Query, dbQueryExpr.Args)
		}
	}

	tx = tx.Limit(in.Limit)
	tx = tx.Offset(in.Offset)

	if in.SortParameter != nil {
		tx = tx.Order(in.SortParameter.GetDBOrderExpression(sourceTableName))
	}
	return tx, nil
}
