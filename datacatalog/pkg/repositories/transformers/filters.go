package transformers

import (
	"context"

	"github.com/flyteorg/datacatalog/pkg/common"

	"github.com/flyteorg/datacatalog/pkg/manager/impl/validators"
	"github.com/flyteorg/datacatalog/pkg/repositories/gormimpl"
	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flytestdlib/logger"
)

const (
	partitionKeyFieldName   = "key"
	partitionValueFieldName = "value"
	tagNameFieldName        = "tag_name"
	projectFieldName        = "project"
	domainFieldName         = "domain"
	nameFieldName           = "name"
	versionFieldName        = "version"
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
	var modelFilter models.ModelFilter

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

			modelFilter = models.ModelFilter{
				Entity:        common.Partition,
				ValueFilters:  modelValueFilters,
				JoinCondition: gormimpl.NewGormJoinCondition(sourceEntity, common.Partition),
			}
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

			modelFilter = models.ModelFilter{
				Entity:        common.Tag,
				ValueFilters:  modelValueFilters,
				JoinCondition: gormimpl.NewGormJoinCondition(sourceEntity, common.Tag),
			}
		}
	case *datacatalog.SinglePropertyFilter_DatasetFilter:
		switch datasetProperty := propertyFilter.DatasetFilter.GetProperty().(type) {
		case *datacatalog.DatasetPropertyFilter_Project:
			project := datasetProperty.Project
			logger.Debugf(ctx, "Constructing Dataset filter project:[%v]", project)
			if err := validators.ValidateEmptyStringField(datasetProperty.Project, "project"); err != nil {
				return models.ModelFilter{}, err
			}
			projectFilter := gormimpl.NewGormValueFilter(operator, projectFieldName, project)
			modelValueFilters := []models.ModelValueFilter{projectFilter}

			modelFilter = models.ModelFilter{
				Entity:       common.Dataset,
				ValueFilters: modelValueFilters,
			}
		case *datacatalog.DatasetPropertyFilter_Domain:
			domain := datasetProperty.Domain
			logger.Debugf(ctx, "Constructing Dataset filter domain:[%v]", domain)
			if err := validators.ValidateEmptyStringField(datasetProperty.Domain, "domain"); err != nil {
				return models.ModelFilter{}, err
			}
			domainFilter := gormimpl.NewGormValueFilter(operator, domainFieldName, domain)
			modelValueFilters := []models.ModelValueFilter{domainFilter}

			modelFilter = models.ModelFilter{
				Entity:       common.Dataset,
				ValueFilters: modelValueFilters,
			}
		case *datacatalog.DatasetPropertyFilter_Name:
			name := datasetProperty.Name
			logger.Debugf(ctx, "Constructing Dataset filter name:[%v]", name)
			if err := validators.ValidateEmptyStringField(datasetProperty.Name, "name"); err != nil {
				return models.ModelFilter{}, err
			}
			nameFilter := gormimpl.NewGormValueFilter(operator, nameFieldName, name)
			modelValueFilters := []models.ModelValueFilter{nameFilter}

			modelFilter = models.ModelFilter{
				Entity:       common.Dataset,
				ValueFilters: modelValueFilters,
			}
		case *datacatalog.DatasetPropertyFilter_Version:
			version := datasetProperty.Version
			logger.Debugf(ctx, "Constructing Dataset filter version:[%v]", version)
			if err := validators.ValidateEmptyStringField(datasetProperty.Version, "version"); err != nil {
				return models.ModelFilter{}, err
			}
			versionFilter := gormimpl.NewGormValueFilter(operator, versionFieldName, version)
			modelValueFilters := []models.ModelValueFilter{versionFilter}

			modelFilter = models.ModelFilter{
				Entity:       common.Dataset,
				ValueFilters: modelValueFilters,
			}
		}
	}

	return modelFilter, nil
}
