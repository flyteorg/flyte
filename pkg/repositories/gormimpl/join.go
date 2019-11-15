package gormimpl

import (
	"fmt"
	"strings"

	"github.com/lyft/datacatalog/pkg/common"
	"github.com/lyft/datacatalog/pkg/repositories/errors"
	"github.com/lyft/datacatalog/pkg/repositories/models"
)

const (
	joinCondition = "JOIN %s ON %s"
	joinEquals    = "%s.%s = %s.%s"
	joinSeparator = " AND "
)

// JoinOnMap is a map of the properties for joining source table to joining table
type JoinOnMap map[string]string

// This provides the field names needed for joining a source Model to joining Model
var joinFieldNames = map[common.Entity]map[common.Entity]JoinOnMap{
	common.Artifact: {
		common.Partition: JoinOnMap{"artifact_id": "artifact_id"},
		common.Tag:       JoinOnMap{"artifact_id": "artifact_id"},
	},
}

// Contains the details to construct GORM JOINs in the format:
// JOIN sourceTable ON sourceTable.sourceField = joiningTable.joiningField
type gormJoinConditionImpl struct {
	// The source entity type
	sourceEntity common.Entity
	// The joining entity type
	joiningEntity common.Entity
}

// Get the GORM expression to JOIN two entities. The output should be a valid input into tx.Join()
func (g *gormJoinConditionImpl) GetJoinOnDBQueryExpression(sourceTableName string, joiningTableName string) (string, error) {
	joinOnFieldMap, err := g.getJoinOnFields()

	if err != nil {
		return "", err
	}

	joinFields := make([]string, 0, len(joinOnFieldMap))
	for sourceField, joiningField := range joinOnFieldMap {
		joinFieldCondition := fmt.Sprintf(joinEquals, sourceTableName, sourceField, joiningTableName, joiningField)
		joinFields = append(joinFields, joinFieldCondition)
	}

	return fmt.Sprintf(joinCondition, joiningTableName, strings.Join(joinFields, joinSeparator)), nil
}

// Get the properties necessary to join two GORM models
func (g *gormJoinConditionImpl) getJoinOnFields() (JoinOnMap, error) {
	joiningEntityMap, ok := joinFieldNames[g.sourceEntity]
	if !ok {
		return nil, errors.GetInvalidEntityRelationshipError(g.sourceEntity, g.joiningEntity)
	}

	fieldMap, ok := joiningEntityMap[g.joiningEntity]
	if !ok {
		return nil, errors.GetInvalidEntityRelationshipError(g.sourceEntity, g.joiningEntity)
	}

	return fieldMap, nil
}

func NewGormJoinCondition(sourceEntity common.Entity, joiningEntity common.Entity) models.ModelJoinCondition {
	return &gormJoinConditionImpl{
		joiningEntity: joiningEntity,
		sourceEntity:  sourceEntity,
	}
}
