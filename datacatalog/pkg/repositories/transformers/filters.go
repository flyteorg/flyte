package transformers

import (
	"context"

	"github.com/lyft/datacatalog/pkg/common"

	"github.com/lyft/datacatalog/pkg/manager/impl/validators"
	"github.com/lyft/datacatalog/pkg/repositories/gormimpl"
	"github.com/lyft/datacatalog/pkg/repositories/models"
	datacatalog "github.com/lyft/datacatalog/protos/gen"
	"github.com/lyft/flytestdlib/logger"
)

const (
	partitionKeyFieldName   = "key"
	partitionValueFieldName = "value"
	tagNameFieldName        = "tag_name"
)

var comparisonOperatorMap = map[datacatalog.SinglePropertyFilter_ComparisonOperator]common.ComparisonOperator{
	datacatalog.SinglePropertyFilter_EQUALS: common.Equal,
}

func FilterToListInput(ctx context.Context, sourceEntity common.Entity, filterExpression *datacatalog.FilterExpression) (models.ListModelsInput, error) {
	// ListInput is composed of ModelFilters and ModelJoins. Let's construct those filters and joins.
	modelFilters := make([]models.ModelValueFilter, 0, len(filterExpression.GetFilters()))
	joinModelMap := make(map[common.Entity]models.ModelJoinCondition)

	// First construct the model filters
	for _, filter := range filterExpression.GetFilters() {
		modelPropertyFilters, err := constructModelValueFilters(ctx, filter)
		if err != nil {
			return models.ListModelsInput{}, err
		}

		modelFilters = append(modelFilters, modelPropertyFilters...)
	}

	// Then add necessary joins if there are filters that are not the source entity
	for _, modelFilter := range modelFilters {
		joinEntity := modelFilter.GetDBEntity()
		if _, exists := joinModelMap[joinEntity]; !exists && sourceEntity != joinEntity {
			joinModelMap[joinEntity] = gormimpl.NewGormJoinCondition(sourceEntity, joinEntity)
		}

	}

	// Need to add limit/offset/Sort
	return models.ListModelsInput{
		Filters:                  modelFilters,
		JoinEntityToConditionMap: joinModelMap,
	}, nil
}

func constructModelValueFilters(ctx context.Context, singleFilter *datacatalog.SinglePropertyFilter) ([]models.ModelValueFilter, error) {
	modelValueFilters := make([]models.ModelValueFilter, 0, 1)
	operator := comparisonOperatorMap[singleFilter.Operator]

	switch propertyFilter := singleFilter.GetPropertyFilter().(type) {
	case *datacatalog.SinglePropertyFilter_PartitionFilter:
		partitionPropertyFilter := singleFilter.GetPartitionFilter()

		switch partitionProperty := partitionPropertyFilter.GetProperty().(type) {
		case *datacatalog.PartitionPropertyFilter_KeyVal:
			key := partitionProperty.KeyVal.Key
			value := partitionProperty.KeyVal.Value

			logger.Debugf(ctx, "Constructing partition key:[%v], val:[%v] filter", key, value)
			if err := validators.ValidateEmptyStringField(key, "PartitionKey"); err != nil {
				return nil, err
			}
			if err := validators.ValidateEmptyStringField(value, "PartitionValue"); err != nil {
				return nil, err
			}
			partitionKeyFilter := gormimpl.NewGormValueFilter(common.Partition, operator, partitionKeyFieldName, key)
			partitionValueFilter := gormimpl.NewGormValueFilter(common.Partition, operator, partitionValueFieldName, value)
			modelValueFilters = append(modelValueFilters, partitionKeyFilter, partitionValueFilter)
		}
	case *datacatalog.SinglePropertyFilter_TagFilter:
		switch tagProperty := propertyFilter.TagFilter.GetProperty().(type) {
		case *datacatalog.TagPropertyFilter_TagName:
			tagName := tagProperty.TagName
			logger.Debugf(ctx, "Constructing Tag filter name:[%v]", tagName)
			if err := validators.ValidateEmptyStringField(tagProperty.TagName, "TagName"); err != nil {
				return nil, err
			}
			tagNameFilter := gormimpl.NewGormValueFilter(common.Tag, operator, tagNameFieldName, tagName)
			modelValueFilters = append(modelValueFilters, tagNameFilter)
		}
	}
	return modelValueFilters, nil
}
