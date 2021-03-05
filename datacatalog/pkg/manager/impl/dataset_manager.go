package impl

import (
	"context"
	"strconv"
	"time"

	"github.com/flyteorg/datacatalog/pkg/common"
	"github.com/flyteorg/datacatalog/pkg/errors"
	"github.com/flyteorg/datacatalog/pkg/manager/impl/validators"
	"github.com/flyteorg/datacatalog/pkg/manager/interfaces"
	"github.com/flyteorg/datacatalog/pkg/repositories"
	"github.com/flyteorg/datacatalog/pkg/repositories/transformers"
	datacatalog "github.com/flyteorg/datacatalog/protos/gen"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flytestdlib/storage"
	"google.golang.org/grpc/codes"
)

type datasetMetrics struct {
	scope                   promutils.Scope
	createResponseTime      labeled.StopWatch
	getResponseTime         labeled.StopWatch
	createSuccessCounter    labeled.Counter
	createErrorCounter      labeled.Counter
	getSuccessCounter       labeled.Counter
	getErrorCounter         labeled.Counter
	listSuccessCounter      labeled.Counter
	listFailureCounter      labeled.Counter
	transformerErrorCounter labeled.Counter
	validationErrorCounter  labeled.Counter
	alreadyExistsCounter    labeled.Counter
	doesNotExistCounter     labeled.Counter
}

type datasetManager struct {
	repo          repositories.RepositoryInterface
	store         *storage.DataStore
	systemMetrics datasetMetrics
}

func (dm *datasetManager) validateCreateRequest(request *datacatalog.CreateDatasetRequest) error {
	errorSet := make([]error, 0)
	err := validators.ValidateDatasetID(request.Dataset.Id)
	if err != nil {
		errorSet = append(errorSet, err)
	}

	err = validators.ValidateUniquePartitionKeys(request.Dataset.PartitionKeys)
	if err != nil {
		errorSet = append(errorSet, err)
	}

	if len(errorSet) > 0 {
		return errors.NewCollectedErrors(codes.InvalidArgument, errorSet)
	}

	return nil
}

// Create a Dataset with optional metadata. If one already exists a grpc AlreadyExists err will be returned
func (dm *datasetManager) CreateDataset(ctx context.Context, request *datacatalog.CreateDatasetRequest) (*datacatalog.CreateDatasetResponse, error) {
	timer := dm.systemMetrics.createResponseTime.Start(ctx)
	defer timer.Stop()

	err := dm.validateCreateRequest(request)
	if err != nil {
		return nil, err
	}

	datasetModel, err := transformers.CreateDatasetModel(request.Dataset)
	if err != nil {
		logger.Errorf(ctx, "Unable to transform create dataset request %+v err: %v", request, err)
		dm.systemMetrics.transformerErrorCounter.Inc(ctx)
		return nil, err
	}

	err = dm.repo.DatasetRepo().Create(ctx, *datasetModel)
	if err != nil {
		if errors.IsAlreadyExistsError(err) {
			logger.Warnf(ctx, "Dataset already exists key: %+v, err %v", request.Dataset, err)
			dm.systemMetrics.alreadyExistsCounter.Inc(ctx)
		} else {
			logger.Errorf(ctx, "Failed to create dataset model: %+v err: %v", datasetModel, err)
			dm.systemMetrics.createErrorCounter.Inc(ctx)
		}
		return nil, err
	}

	logger.Debugf(ctx, "Successfully created dataset %+v", request.Dataset)
	dm.systemMetrics.createSuccessCounter.Inc(ctx)
	return &datacatalog.CreateDatasetResponse{}, nil
}

// Get a Dataset with the given DatasetID if it exists. If none exist a grpc NotFound err will be returned
func (dm *datasetManager) GetDataset(ctx context.Context, request *datacatalog.GetDatasetRequest) (*datacatalog.GetDatasetResponse, error) {
	timer := dm.systemMetrics.getResponseTime.Start(ctx)
	defer timer.Stop()

	err := validators.ValidateDatasetID(request.Dataset)
	if err != nil {
		logger.Warnf(ctx, "Invalid get dataset request %+v err: %v", request, err)
		dm.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	datasetKey := transformers.FromDatasetID(request.Dataset)
	datasetModel, err := dm.repo.DatasetRepo().Get(ctx, datasetKey)

	if err != nil {
		if errors.IsDoesNotExistError(err) {
			logger.Warnf(ctx, "Dataset does not exist key: %+v, err %v", datasetKey, err)
			dm.systemMetrics.doesNotExistCounter.Inc(ctx)
		} else {
			logger.Errorf(ctx, "Unable to get dataset request %+v err: %v", request, err)
			dm.systemMetrics.getErrorCounter.Inc(ctx)
		}
		return nil, err
	}

	datasetResponse, err := transformers.FromDatasetModel(datasetModel)
	if err != nil {
		dm.systemMetrics.transformerErrorCounter.Inc(ctx)
		return nil, err
	}

	dm.systemMetrics.getSuccessCounter.Inc(ctx)
	return &datacatalog.GetDatasetResponse{
		Dataset: datasetResponse,
	}, nil
}

// List Datasets with optional filtering and pagination
func (dm *datasetManager) ListDatasets(ctx context.Context, request *datacatalog.ListDatasetsRequest) (*datacatalog.ListDatasetsResponse, error) {
	err := validators.ValidateListDatasetsRequest(request)
	if err != nil {
		logger.Warningf(ctx, "Invalid list datasets request %v, err: %v", request, err)
		dm.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	// Get the list inputs
	listInput, err := transformers.FilterToListInput(ctx, common.Dataset, request.GetFilter())
	if err != nil {
		logger.Warningf(ctx, "Invalid list datasets request %v, err: %v", request, err)
		dm.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	err = transformers.ApplyPagination(request.Pagination, &listInput)
	if err != nil {
		logger.Warningf(ctx, "Invalid pagination options in list datasets request %v, err: %v", request, err)
		dm.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	// Perform the list with the dataset and listInput filters
	datasetModels, err := dm.repo.DatasetRepo().List(ctx, listInput)
	if err != nil {
		logger.Errorf(ctx, "Unable to list Datasets err: %v", err)
		dm.systemMetrics.listFailureCounter.Inc(ctx)
		return nil, err
	}

	// convert returned models into entity list
	datasetList := make([]*datacatalog.Dataset, len(datasetModels))
	transformerErrs := make([]error, 0)
	for idx, datasetModel := range datasetModels {
		dataset, err := transformers.FromDatasetModel(datasetModel)
		if err != nil {
			logger.Errorf(ctx, "Unable to transform Dataset %+v err: %v", dataset.Id, err)
			transformerErrs = append(transformerErrs, err)
		}

		datasetList[idx] = dataset
	}

	if len(transformerErrs) > 0 {
		dm.systemMetrics.listFailureCounter.Inc(ctx)
		return nil, errors.NewCollectedErrors(codes.Internal, transformerErrs)
	}

	token := strconv.Itoa(int(listInput.Offset) + len(datasetList))

	logger.Debugf(ctx, "Listed %v matching datasets successfully", len(datasetList))
	dm.systemMetrics.listSuccessCounter.Inc(ctx)
	return &datacatalog.ListDatasetsResponse{Datasets: datasetList, NextToken: token}, nil
}

func NewDatasetManager(repo repositories.RepositoryInterface, store *storage.DataStore, datasetScope promutils.Scope) interfaces.DatasetManager {
	return &datasetManager{
		repo:  repo,
		store: store,
		systemMetrics: datasetMetrics{
			scope:                   datasetScope,
			createResponseTime:      labeled.NewStopWatch("create_duration", "The duration of the create dataset calls.", time.Millisecond, datasetScope, labeled.EmitUnlabeledMetric),
			getResponseTime:         labeled.NewStopWatch("get_duration", "The duration of the get dataset calls.", time.Millisecond, datasetScope, labeled.EmitUnlabeledMetric),
			createSuccessCounter:    labeled.NewCounter("create_success_count", "The number of times create dataset was called", datasetScope, labeled.EmitUnlabeledMetric),
			getSuccessCounter:       labeled.NewCounter("get_success_count", "The number of times get dataset was called", datasetScope, labeled.EmitUnlabeledMetric),
			createErrorCounter:      labeled.NewCounter("create_failed_count", "The number of times create dataset failed", datasetScope, labeled.EmitUnlabeledMetric),
			getErrorCounter:         labeled.NewCounter("get_failed_count", "The number of times get dataset failed", datasetScope, labeled.EmitUnlabeledMetric),
			transformerErrorCounter: labeled.NewCounter("transformer_failed_count", "The number of times transformations failed", datasetScope, labeled.EmitUnlabeledMetric),
			validationErrorCounter:  labeled.NewCounter("validation_failed_count", "The number of times validation failed", datasetScope, labeled.EmitUnlabeledMetric),
			alreadyExistsCounter:    labeled.NewCounter("already_exists_count", "The number of times a dataset already exists", datasetScope, labeled.EmitUnlabeledMetric),
			doesNotExistCounter:     labeled.NewCounter("does_not_exists_count", "The number of times a dataset was not found", datasetScope, labeled.EmitUnlabeledMetric),
			listSuccessCounter:      labeled.NewCounter("list_success_count", "The number of times list dataset succeeded", datasetScope, labeled.EmitUnlabeledMetric),
			listFailureCounter:      labeled.NewCounter("list_failure_count", "The number of times list dataset failed", datasetScope, labeled.EmitUnlabeledMetric),
		},
	}
}
