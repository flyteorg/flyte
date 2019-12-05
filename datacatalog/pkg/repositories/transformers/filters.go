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
	// ListInput is composed of filters and joins for multiple entities, lets construct that
	modelFilters := make([]models.ModelFilter, 0, len(filterExpression.GetFilters()))

	// Construct the ModelFilter for each PropertyFilter
	for _, filter := range filterExpression.GetFilters() {
		modelFilter, err := constructModelFilter(ctx, filter, sourceEntity)
		if err != nil {
			return models.ListModelsInput{}, err
		}
		modelFilters = append(modelFilters, modelFilter)
	}

	return models.ListModelsInput{
		ModelFilters: modelFilters,
	}, nil
}

func constructModelFilter(ctx context.Context, singleFilter *datacatalog.SinglePropertyFilter, sourceEntity common.Entity) (models.ModelFilter, error) {
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
				return models.ModelFilter{}, err
			}
			if err := validators.ValidateEmptyStringField(value, "PartitionValue"); err != nil {
				return models.ModelFilter{}, err
			}
			partitionKeyFilter := gormimpl.NewGormValueFilter(operator, partitionKeyFieldName, key)
			partitionValueFilter := gormimpl.NewGormValueFilter(operator, partitionValueFieldName, value)
			modelValueFilters := []models.ModelValueFilter{partitionKeyFilter, partitionValueFilter}

			return models.ModelFilter{
				Entity:        common.Partition,
				ValueFilters:  modelValueFilters,
				JoinCondition: gormimpl.NewGormJoinCondition(sourceEntity, common.Partition),
			}, nil
		}
	case *datacatalog.SinglePropertyFilter_TagFilter:
		switch tagProperty := propertyFilter.TagFilter.GetProperty().(type) {
		case *datacatalog.TagPropertyFilter_TagName:
			tagName := tagProperty.TagName
			logger.Debugf(ctx, "Constructing Tag filter name:[%v]", tagName)
			if err := validators.ValidateEmptyStringField(tagProperty.TagName, "TagName"); err != nil {
				return models.ModelFilter{}, err
			}
			tagNameFilter := gormimpl.NewGormValueFilter(operator, tagNameFieldName, tagName)
			modelValueFilters := []models.ModelValueFilter{tagNameFilter}

			return models.ModelFilter{
				Entity:        common.Tag,
				ValueFilters:  modelValueFilters,
				JoinCondition: gormimpl.NewGormJoinCondition(sourceEntity, common.Tag),
			}, nil
		}
	}
	return models.ModelFilter{}, nil
}
