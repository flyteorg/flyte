package impl

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/datacatalog/pkg/errors"
	"github.com/flyteorg/flyte/datacatalog/pkg/manager/impl/validators"
	"github.com/flyteorg/flyte/datacatalog/pkg/manager/interfaces"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/flyte/datacatalog/pkg/repositories/transformers"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type futureArtifactMetrics struct {
	scope                    promutils.Scope
	createResponseTime       labeled.StopWatch
	getResponseTime          labeled.StopWatch
	createSuccessCounter     labeled.Counter
	createFailureCounter     labeled.Counter
	getSuccessCounter        labeled.Counter
	getFailureCounter        labeled.Counter
	createDataFailureCounter labeled.Counter
	createDataSuccessCounter labeled.Counter
	transformerErrorCounter  labeled.Counter
	validationErrorCounter   labeled.Counter
	alreadyExistsCounter     labeled.Counter
	doesNotExistCounter      labeled.Counter
	updateResponseTime       labeled.StopWatch
	updateSuccessCounter     labeled.Counter
	updateFailureCounter     labeled.Counter
	updateDataSuccessCounter labeled.Counter
	updateDataFailureCounter labeled.Counter
}

type futureArtifactManager struct {
	repo          repositories.RepositoryInterface
	artifactStore ArtifactDataStore
	systemMetrics futureArtifactMetrics
}

// CreateFutureArtifact will create a future.pb cache in blob storage
func (f *futureArtifactManager) CreateFutureArtifact(ctx context.Context, request *datacatalog.CreateArtifactRequest) (*datacatalog.CreateArtifactResponse, error) {
	timer := f.systemMetrics.createResponseTime.Start(ctx)
	defer timer.Stop()

	artifact := request.GetArtifact()
	err := validators.ValidateArtifact(artifact)
	if err != nil {
		logger.Warningf(ctx, "Invalid create artifact request %v, err: %v", request, err)
		f.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	datasetKey := transformers.FromDatasetID(artifact.GetDataset())
	dataset, err := f.repo.DatasetRepo().Get(ctx, datasetKey)
	if err != nil {
		logger.Warnf(ctx, "Failed to get dataset for artifact creation %v, err: %v", datasetKey, err)
		f.systemMetrics.createFailureCounter.Inc(ctx)
		return nil, err
	}

	// check that the artifact's partitions are the same partition values of the dataset
	datasetPartitionKeys := transformers.FromPartitionKeyModel(dataset.PartitionKeys)
	err = validators.ValidatePartitions(datasetPartitionKeys, artifact.GetPartitions())
	if err != nil {
		logger.Warnf(ctx, "Invalid artifact partitions %v, err: %+v", artifact.GetPartitions(), err)
		f.systemMetrics.createFailureCounter.Inc(ctx)
		return nil, err
	}
	// create future Artifact Data offloaded storage files
	artifactDataModels := make([]models.ArtifactData, len(request.GetArtifact().GetData()))
	for i, artifactData := range request.GetArtifact().GetData() {
		if artifactData.GetName() == "future" {
			djSpec, err := transformers.FromFutureLiteralToDynamicJobSpec(ctx, artifactData.GetValue())
			if err != nil {
				logger.Errorf(ctx, "Failed to convert future literal to dynamic job spec err: %v", err)
				f.systemMetrics.createDataFailureCounter.Inc(ctx)
				return nil, err
			}
			dataLocation, err := f.artifactStore.PutFutureData(ctx, artifact, djSpec)
			if err != nil {
				logger.Errorf(ctx, "Failed to store artifact data err: %v", err)
				f.systemMetrics.createDataFailureCounter.Inc(ctx)
				return nil, err
			}
			artifactDataModels[i].Name = artifactData.GetName()
			artifactDataModels[i].Location = dataLocation.String()
			f.systemMetrics.createDataSuccessCounter.Inc(ctx)
		}
	}

	if len(artifactDataModels) == 0 {
		err := errors.NewDataCatalogErrorf(codes.InvalidArgument, "No future literal found in request")
		logger.Error(ctx, "Invalid create request: %v", err)
		f.systemMetrics.createFailureCounter.Inc(ctx)
		return nil, err
	}

	logger.Debugf(ctx, "Stored %v data for artifact %+v", len(artifactDataModels), artifact.GetId())

	artifactModel, err := transformers.CreateArtifactModel(request, artifactDataModels, dataset)
	if err != nil {
		logger.Errorf(ctx, "Failed to transform artifact err: %v", err)
		f.systemMetrics.transformerErrorCounter.Inc(ctx)
		return nil, err
	}

	err = f.repo.ArtifactRepo().Create(ctx, artifactModel)
	if err != nil {
		if errors.IsAlreadyExistsError(err) {
			logger.Warnf(ctx, "Artifact already exists key: %+v, err %v", artifact.GetId(), err)
			f.systemMetrics.alreadyExistsCounter.Inc(ctx)
		} else {
			logger.Errorf(ctx, "Failed to create artifact %v, err: %v", artifactDataModels, err)
			f.systemMetrics.createFailureCounter.Inc(ctx)
		}
		return nil, err
	}

	logger.Debugf(ctx, "Successfully created artifact id: %v", artifact.GetId())

	f.systemMetrics.createSuccessCounter.Inc(ctx)
	return &datacatalog.CreateArtifactResponse{}, nil
}

// GetFutureArtifact retrieves an artifact that contains future data.
func (f *futureArtifactManager) GetFutureArtifact(ctx context.Context, request *datacatalog.GetArtifactRequest) (*datacatalog.GetArtifactResponse, error) {
	timer := f.systemMetrics.getResponseTime.Start(ctx)
	defer timer.Stop()

	err := validators.ValidateGetArtifactRequest(request)
	if err != nil {
		logger.Warningf(ctx, "Invalid get future artifact request %v, err: %v", request, err)
		f.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	artifactModel, err := f.findFutureArtifact(ctx, request.GetDataset(), request)
	if err != nil {
		logger.Errorf(ctx, "Failed to retrieve artifact for get artifact request %v, err: %v", request, err)
		f.systemMetrics.getFailureCounter.Inc(ctx)
		return nil, err
	}

	artifact, err := transformers.FromArtifactModel(artifactModel)
	if err != nil {
		logger.Errorf(ctx, "Error in transforming get artifact request %+v, err %v", artifactModel, err)
		f.systemMetrics.transformerErrorCounter.Inc(ctx)
		return nil, err
	}

	artifactDataList, err := f.getFutureArtifactDataList(ctx, artifactModel.ArtifactData)
	if err != nil {
		f.systemMetrics.getFailureCounter.Inc(ctx)
		return nil, err
	}
	artifact.Data = artifactDataList

	logger.Debugf(ctx, "Retrieved artifact dataset %v, id: %v", artifact.GetDataset(), artifact.GetId())
	f.systemMetrics.getSuccessCounter.Inc(ctx)
	return &datacatalog.GetArtifactResponse{
		Artifact: artifact,
	}, nil
}

func (f *futureArtifactManager) getFutureArtifactDataList(ctx context.Context, artifactDataModels []models.ArtifactData) ([]*datacatalog.ArtifactData, error) {
	artifactDataList := make([]*datacatalog.ArtifactData, len(artifactDataModels))
	for i, artifactData := range artifactDataModels {
		if artifactData.Name == futureDataName {
			value, err := f.artifactStore.GetFutureData(ctx, artifactData)
			if err != nil {
				logger.Errorf(ctx, "Error in getting future artifact data from datastore %+v, err %v", artifactData.Location, err)
				return nil, err
			}
			// we must convert the DynamicJobSpec to a Literal before inserting into the ArtifactData list
			literal, err := transformers.FromDynamicJobSpecToLiteral(value)
			if err != nil {
				logger.Errorf(ctx, "Error in converting DynamicJobSpec to Literal for future artifact data %+v, err %v", artifactData.Location, err)
				f.systemMetrics.transformerErrorCounter.Inc(ctx)
				return nil, err
			}
			artifactDataList[i] = &datacatalog.ArtifactData{
				Name:  artifactData.Name,
				Value: literal,
			}
		}
	}

	return artifactDataList, nil
}

type futureArtifactQueryHandle interface {
	GetArtifactId() string
	GetTagName() string
}

func (f *futureArtifactManager) findFutureArtifact(ctx context.Context, datasetID *datacatalog.DatasetID, queryHandle futureArtifactQueryHandle) (models.Artifact, error) {
	var artifactModel models.Artifact

	key := queryHandle.GetArtifactId()
	if len(key) > 0 {
		logger.Debugf(ctx, "Get future artifact by id %v", key)
		artifactKey := transformers.ToArtifactKey(datasetID, key)
		var err error
		artifactModel, err = f.repo.ArtifactRepo().Get(ctx, artifactKey)

		if err != nil {
			if errors.IsDoesNotExistError(err) {
				logger.Warnf(ctx, "Future artifact does not exist id: %+v, err %v", key, err)
				f.systemMetrics.doesNotExistCounter.Inc(ctx)
			} else {
				logger.Errorf(ctx, "Unable to retrieve future artifact by id: %+v, err %v", key, err)
			}
			return models.Artifact{}, err
		}
	} else {
		key = queryHandle.GetTagName()

		logger.Debugf(ctx, "Get future artifact by tag %v", key)
		tagKey := transformers.ToTagKey(datasetID, key)
		tag, err := f.repo.TagRepo().Get(ctx, tagKey)

		if err != nil {
			if errors.IsDoesNotExistError(err) {
				logger.Infof(ctx, "Future artifact does not exist tag: %+v, err %v", key, err)
				f.systemMetrics.doesNotExistCounter.Inc(ctx)
			} else {
				logger.Errorf(ctx, "Unable to retrieve future artifact by tag %v, err: %v", key, err)
			}
			return models.Artifact{}, err
		}

		artifactModel = tag.Artifact
	}

	if len(artifactModel.ArtifactData) == 0 {
		return models.Artifact{}, errors.NewDataCatalogErrorf(codes.Internal, "artifact [%+v] with key %v does not have artifact data associated", artifactModel, key)
	}

	return artifactModel, nil
}

func (f *futureArtifactManager) UpdateFutureArtifact(ctx context.Context, request *datacatalog.UpdateArtifactRequest) (*datacatalog.UpdateArtifactResponse, error) {
	ctx = contextutils.WithProjectDomain(ctx, request.GetDataset().GetProject(), request.GetDataset().GetDomain())

	timer := f.systemMetrics.updateResponseTime.Start(ctx)
	defer timer.Stop()

	err := validators.ValidateUpdateArtifactRequest(request)
	if err != nil {
		logger.Warningf(ctx, "Invalid update artifact request %v, err: %v", request, err)
		f.systemMetrics.validationErrorCounter.Inc(ctx)
		f.systemMetrics.updateFailureCounter.Inc(ctx)
		return nil, err
	}

	// artifact must already exist, verify first
	artifactModel, err := f.findFutureArtifact(ctx, request.GetDataset(), request)
	if err != nil {
		logger.Errorf(ctx, "Failed to get artifact for update artifact request %v, err: %v", request, err)
		f.systemMetrics.updateFailureCounter.Inc(ctx)
		return nil, err
	}

	// artifactModel needs to be updated with new SerializedMetadata
	serializedMetadata, err := transformers.SerializedMetadata(request.GetMetadata())
	if err != nil {
		logger.Errorf(ctx, "Error in transforming Metadata from request %+v, err %v", request.GetMetadata(), err)
		f.systemMetrics.transformerErrorCounter.Inc(ctx)
		f.systemMetrics.updateFailureCounter.Inc(ctx)
		return nil, err
	}
	artifactModel.SerializedMetadata = serializedMetadata

	artifact, err := transformers.FromArtifactModel(artifactModel)
	if err != nil {
		logger.Errorf(ctx, "Error in transforming update artifact request %+v, err %v", artifactModel, err)
		f.systemMetrics.transformerErrorCounter.Inc(ctx)
		f.systemMetrics.updateFailureCounter.Inc(ctx)
		return nil, err
	}

	// overwrite existing future artifact data and upload new entries, building a map of artifact data names to remove
	// deleted entries from the blob storage after the upload completed
	artifactDataModels := make([]models.ArtifactData, len(request.GetData()))
	for i, artifactData := range request.GetData() {
		if artifactData.GetName() == futureDataName {
			djSpec, err := transformers.FromFutureLiteralToDynamicJobSpec(ctx, artifactData.GetValue())
			if err != nil {
				logger.Errorf(ctx, "Failed to convert future literal to dynamic job spec err: %v", err)
				f.systemMetrics.createDataFailureCounter.Inc(ctx)
				return nil, err
			}
			dataLocation, err := f.artifactStore.PutFutureData(ctx, artifact, djSpec)
			if err != nil {
				logger.Errorf(ctx, "Failed to store future artifact data during update, err: %v", err)
				f.systemMetrics.updateDataFailureCounter.Inc(ctx)
				f.systemMetrics.updateFailureCounter.Inc(ctx)
				return nil, err
			}
			artifactDataModels[i].Name = artifactData.GetName()
			artifactDataModels[i].Location = dataLocation.String()
			f.systemMetrics.updateDataSuccessCounter.Inc(ctx)
		}
	}

	if len(artifactDataModels) == 0 {
		err := errors.NewDataCatalogErrorf(codes.InvalidArgument, "No future literal found in request")
		logger.Error(ctx, "Invalid update request: %v", err)
		f.systemMetrics.updateFailureCounter.Inc(ctx)
		return nil, err
	}

	// update future artifact in DB, also replaces/upserts associated artifact data
	artifactModel.ArtifactData = artifactDataModels
	logger.Debugf(ctx, "Updating Future ArtifactModel with %+v", artifactModel)

	err = f.repo.ArtifactRepo().Update(ctx, artifactModel)
	if err != nil {
		if errors.IsDoesNotExistError(err) {
			logger.Warnf(ctx, "Future Artifact does not exist key: %+v, err %v", artifact.GetId(), err)
			f.systemMetrics.doesNotExistCounter.Inc(ctx)
		} else {
			logger.Errorf(ctx, "Failed to update future artifact %v, err: %v", artifactModel, err)
		}
		f.systemMetrics.updateFailureCounter.Inc(ctx)
		return nil, err
	}

	logger.Debugf(ctx, "Successfully updated future artifact id: %v", artifact.GetId())
	f.systemMetrics.updateSuccessCounter.Inc(ctx)
	return &datacatalog.UpdateArtifactResponse{
		ArtifactId: artifact.GetId(),
	}, nil
}

func NewFutureArtifactManager(repo repositories.RepositoryInterface, store *storage.DataStore, storagePrefix storage.DataReference, futureArtifactScope promutils.Scope) interfaces.FutureArtifactManager {
	futureArtifactMetrics := futureArtifactMetrics{
		scope:                    futureArtifactScope,
		createResponseTime:       labeled.NewStopWatch("create_duration", "The duration of the create artifact calls.", time.Millisecond, futureArtifactScope, labeled.EmitUnlabeledMetric),
		getResponseTime:          labeled.NewStopWatch("get_duration", "The duration of the get artifact calls.", time.Millisecond, futureArtifactScope, labeled.EmitUnlabeledMetric),
		createSuccessCounter:     labeled.NewCounter("create_success_count", "The number of times create artifact succeeded", futureArtifactScope, labeled.EmitUnlabeledMetric),
		getSuccessCounter:        labeled.NewCounter("get_success_count", "The number of times get artifact succeeded", futureArtifactScope, labeled.EmitUnlabeledMetric),
		createFailureCounter:     labeled.NewCounter("create_failure_count", "The number of times create artifact failed", futureArtifactScope, labeled.EmitUnlabeledMetric),
		getFailureCounter:        labeled.NewCounter("get_failure_count", "The number of times get artifact failed", futureArtifactScope, labeled.EmitUnlabeledMetric),
		createDataFailureCounter: labeled.NewCounter("create_data_failure_count", "The number of times create artifact data failed", futureArtifactScope, labeled.EmitUnlabeledMetric),
		createDataSuccessCounter: labeled.NewCounter("create_data_success_count", "The number of times create artifact data succeeded", futureArtifactScope, labeled.EmitUnlabeledMetric),
		transformerErrorCounter:  labeled.NewCounter("transformer_failed_count", "The number of times transformations failed", futureArtifactScope, labeled.EmitUnlabeledMetric),
		validationErrorCounter:   labeled.NewCounter("validation_failed_count", "The number of times validation failed", futureArtifactScope, labeled.EmitUnlabeledMetric),
		alreadyExistsCounter:     labeled.NewCounter("already_exists_count", "The number of times an artifact already exists", futureArtifactScope, labeled.EmitUnlabeledMetric),
		doesNotExistCounter:      labeled.NewCounter("does_not_exists_count", "The number of times an artifact was not found", futureArtifactScope, labeled.EmitUnlabeledMetric),
		updateResponseTime:       labeled.NewStopWatch("update_duration", "The duration of the update artifact calls.", time.Millisecond, futureArtifactScope, labeled.EmitUnlabeledMetric),
		updateSuccessCounter:     labeled.NewCounter("update_success_count", "The number of times update artifact succeeded", futureArtifactScope, labeled.EmitUnlabeledMetric),
		updateFailureCounter:     labeled.NewCounter("update_failure_count", "The number of times update artifact failed", futureArtifactScope, labeled.EmitUnlabeledMetric),
		updateDataSuccessCounter: labeled.NewCounter("update_data_success_count", "The number of times update artifact data succeeded", futureArtifactScope, labeled.EmitUnlabeledMetric),
		updateDataFailureCounter: labeled.NewCounter("update_data_failure_count", "The number of times update artifact data failed", futureArtifactScope, labeled.EmitUnlabeledMetric),
	}

	return &futureArtifactManager{
		repo:          repo,
		artifactStore: NewArtifactDataStore(store, storagePrefix),
		systemMetrics: futureArtifactMetrics,
	}
}
