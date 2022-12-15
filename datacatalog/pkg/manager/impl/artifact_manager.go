package impl

import (
	"context"
	"strconv"
	"time"

	"github.com/flyteorg/datacatalog/pkg/errors"
	"github.com/flyteorg/datacatalog/pkg/manager/impl/validators"
	"github.com/flyteorg/datacatalog/pkg/manager/interfaces"
	"github.com/flyteorg/datacatalog/pkg/repositories"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"

	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/datacatalog/pkg/repositories/transformers"

	"github.com/flyteorg/datacatalog/pkg/common"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flytestdlib/storage"
	"google.golang.org/grpc/codes"
)

type artifactMetrics struct {
	scope                    promutils.Scope
	createResponseTime       labeled.StopWatch
	getResponseTime          labeled.StopWatch
	createSuccessCounter     labeled.Counter
	createFailureCounter     labeled.Counter
	getSuccessCounter        labeled.Counter
	getFailureCounter        labeled.Counter
	listSuccessCounter       labeled.Counter
	listFailureCounter       labeled.Counter
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
	deleteDataSuccessCounter labeled.Counter
	deleteDataFailureCounter labeled.Counter
	deleteResponseTime       labeled.StopWatch
	deleteSuccessCounter     labeled.Counter
	deleteFailureCounter     labeled.Counter
}

type artifactManager struct {
	repo          repositories.RepositoryInterface
	artifactStore ArtifactDataStore
	systemMetrics artifactMetrics
}

// Create an Artifact along with the associated ArtifactData. The ArtifactData will be stored in an offloaded location.
func (m *artifactManager) CreateArtifact(ctx context.Context, request *datacatalog.CreateArtifactRequest) (*datacatalog.CreateArtifactResponse, error) {
	timer := m.systemMetrics.createResponseTime.Start(ctx)
	defer timer.Stop()

	artifact := request.Artifact
	err := validators.ValidateArtifact(artifact)
	if err != nil {
		logger.Warningf(ctx, "Invalid create artifact request %v, err: %v", request, err)
		m.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	ctx = contextutils.WithProjectDomain(ctx, artifact.Dataset.Project, artifact.Dataset.Domain)
	datasetKey := transformers.FromDatasetID(artifact.Dataset)

	// The dataset must exist for the artifact, let's verify that first
	dataset, err := m.repo.DatasetRepo().Get(ctx, datasetKey)
	if err != nil {
		logger.Warnf(ctx, "Failed to get dataset for artifact creation %v, err: %v", datasetKey, err)
		m.systemMetrics.createFailureCounter.Inc(ctx)
		return nil, err
	}

	// TODO: when adding a tag, need to verify one tag per partition combo
	// check that the artifact's partitions are the same partition values of the dataset
	datasetPartitionKeys := transformers.FromPartitionKeyModel(dataset.PartitionKeys)
	err = validators.ValidatePartitions(datasetPartitionKeys, artifact.Partitions)
	if err != nil {
		logger.Warnf(ctx, "Invalid artifact partitions %v, err: %+v", artifact.Partitions, err)
		m.systemMetrics.createFailureCounter.Inc(ctx)
		return nil, err
	}

	// create Artifact Data offloaded storage files
	artifactDataModels := make([]models.ArtifactData, len(request.Artifact.Data))
	for i, artifactData := range request.Artifact.Data {
		dataLocation, err := m.artifactStore.PutData(ctx, artifact, artifactData)
		if err != nil {
			logger.Errorf(ctx, "Failed to store artifact data err: %v", err)
			m.systemMetrics.createDataFailureCounter.Inc(ctx)
			return nil, err
		}

		artifactDataModels[i].Name = artifactData.Name
		artifactDataModels[i].Location = dataLocation.String()
		m.systemMetrics.createDataSuccessCounter.Inc(ctx)
	}

	logger.Debugf(ctx, "Stored %v data for artifact %+v", len(artifactDataModels), artifact.Id)

	artifactModel, err := transformers.CreateArtifactModel(request, artifactDataModels, dataset)
	if err != nil {
		logger.Errorf(ctx, "Failed to transform artifact err: %v", err)
		m.systemMetrics.transformerErrorCounter.Inc(ctx)
		return nil, err
	}

	err = m.repo.ArtifactRepo().Create(ctx, artifactModel)
	if err != nil {
		if errors.IsAlreadyExistsError(err) {
			logger.Warnf(ctx, "Artifact already exists key: %+v, err %v", artifact.Id, err)
			m.systemMetrics.alreadyExistsCounter.Inc(ctx)
		} else {
			logger.Errorf(ctx, "Failed to create artifact %v, err: %v", artifactDataModels, err)
			m.systemMetrics.createFailureCounter.Inc(ctx)
		}
		return nil, err
	}

	logger.Debugf(ctx, "Successfully created artifact id: %v", artifact.Id)

	m.systemMetrics.createSuccessCounter.Inc(ctx)
	return &datacatalog.CreateArtifactResponse{}, nil
}

// Get the Artifact and its associated ArtifactData. The request can query by ArtifactID or TagName.
func (m *artifactManager) GetArtifact(ctx context.Context, request *datacatalog.GetArtifactRequest) (*datacatalog.GetArtifactResponse, error) {
	timer := m.systemMetrics.getResponseTime.Start(ctx)
	defer timer.Stop()

	err := validators.ValidateGetArtifactRequest(request)
	if err != nil {
		logger.Warningf(ctx, "Invalid get artifact request %v, err: %v", request, err)
		m.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	datasetID := request.Dataset

	artifactModel, err := m.findArtifact(ctx, datasetID, request)
	if err != nil {
		logger.Errorf(ctx, "Failed to retrieve artifact for get artifact request %v, err: %v", request, err)
		m.systemMetrics.getFailureCounter.Inc(ctx)
		return nil, err
	}

	artifact, err := transformers.FromArtifactModel(artifactModel)
	if err != nil {
		logger.Errorf(ctx, "Error in transforming get artifact request %+v, err %v", artifactModel, err)
		m.systemMetrics.transformerErrorCounter.Inc(ctx)
		return nil, err
	}

	artifactDataList, err := m.getArtifactDataList(ctx, artifactModel.ArtifactData)
	if err != nil {
		m.systemMetrics.getFailureCounter.Inc(ctx)
		return nil, err
	}
	artifact.Data = artifactDataList

	logger.Debugf(ctx, "Retrieved artifact dataset %v, id: %v", artifact.Dataset, artifact.Id)
	m.systemMetrics.getSuccessCounter.Inc(ctx)
	return &datacatalog.GetArtifactResponse{
		Artifact: artifact,
	}, nil
}

type artifactQueryHandle interface {
	GetArtifactId() string
	GetTagName() string
}

func (m *artifactManager) findArtifact(ctx context.Context, datasetID *datacatalog.DatasetID, queryHandle artifactQueryHandle) (models.Artifact, error) {
	var artifactModel models.Artifact

	key := queryHandle.GetArtifactId()
	if len(key) > 0 {
		logger.Debugf(ctx, "Get artifact by id %v", key)
		artifactKey := transformers.ToArtifactKey(datasetID, key)
		var err error
		artifactModel, err = m.repo.ArtifactRepo().Get(ctx, artifactKey)

		if err != nil {
			if errors.IsDoesNotExistError(err) {
				logger.Warnf(ctx, "Artifact does not exist id: %+v, err %v", key, err)
				m.systemMetrics.doesNotExistCounter.Inc(ctx)
			} else {
				logger.Errorf(ctx, "Unable to retrieve artifact by id: %+v, err %v", key, err)
			}
			return models.Artifact{}, err
		}
	} else {
		key = queryHandle.GetTagName()

		logger.Debugf(ctx, "Get artifact by tag %v", key)
		tagKey := transformers.ToTagKey(datasetID, key)
		tag, err := m.repo.TagRepo().Get(ctx, tagKey)

		if err != nil {
			if errors.IsDoesNotExistError(err) {
				logger.Infof(ctx, "Artifact does not exist tag: %+v, err %v", key, err)
				m.systemMetrics.doesNotExistCounter.Inc(ctx)
			} else {
				logger.Errorf(ctx, "Unable to retrieve Artifact by tag %v, err: %v", key, err)
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

func (m *artifactManager) getArtifactDataList(ctx context.Context, artifactDataModels []models.ArtifactData) ([]*datacatalog.ArtifactData, error) {
	artifactDataList := make([]*datacatalog.ArtifactData, len(artifactDataModels))
	for i, artifactData := range artifactDataModels {
		value, err := m.artifactStore.GetData(ctx, artifactData)
		if err != nil {
			logger.Errorf(ctx, "Error in getting artifact data from datastore %+v, err %v", artifactData.Location, err)
			return nil, err
		}

		artifactDataList[i] = &datacatalog.ArtifactData{
			Name:  artifactData.Name,
			Value: value,
		}
	}

	return artifactDataList, nil
}

func (m *artifactManager) ListArtifacts(ctx context.Context, request *datacatalog.ListArtifactsRequest) (*datacatalog.ListArtifactsResponse, error) {
	err := validators.ValidateListArtifactRequest(request)
	if err != nil {
		logger.Warningf(ctx, "Invalid list artifact request %v, err: %v", request, err)
		m.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	// Verify the dataset exists before listing artifacts
	datasetKey := transformers.FromDatasetID(request.Dataset)
	dataset, err := m.repo.DatasetRepo().Get(ctx, datasetKey)
	if err != nil {
		logger.Warnf(ctx, "Failed to get dataset for listing artifacts %v, err: %v", datasetKey, err)
		m.systemMetrics.listFailureCounter.Inc(ctx)
		return nil, err
	}

	// Get the list inputs
	listInput, err := transformers.FilterToListInput(ctx, common.Artifact, request.GetFilter())
	if err != nil {
		logger.Warningf(ctx, "Invalid list artifact request %v, err: %v", request, err)
		m.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	err = transformers.ApplyPagination(request.Pagination, &listInput)
	if err != nil {
		logger.Warningf(ctx, "Invalid pagination options in list artifact request %v, err: %v", request, err)
		m.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	// Perform the list with the dataset and listInput filters
	artifactModels, err := m.repo.ArtifactRepo().List(ctx, dataset.DatasetKey, listInput)
	if err != nil {
		logger.Errorf(ctx, "Unable to list Artifacts err: %v", err)
		m.systemMetrics.listFailureCounter.Inc(ctx)
		return nil, err
	}

	// convert returned models into entity list
	artifactsList, err := transformers.FromArtifactModels(artifactModels)
	if err != nil {
		logger.Errorf(ctx, "Unable to transform Artifacts %+v err: %v", artifactModels, err)
		m.systemMetrics.listFailureCounter.Inc(ctx)
		return nil, err
	}

	// Get the artifact data for the artifacts. It retrieves the data from storage and
	// unmarshals the data.
	for i, artifact := range artifactsList {
		artifactDataList, err := m.getArtifactDataList(ctx, artifactModels[i].ArtifactData)
		if err != nil {
			logger.Errorf(ctx, "Unable to transform Artifacts %+v err: %v", artifactModels, err)
			m.systemMetrics.listFailureCounter.Inc(ctx)
			return nil, err
		}
		artifact.Data = artifactDataList
	}

	token := strconv.Itoa(listInput.Offset + len(artifactsList))

	logger.Debugf(ctx, "Listed %v matching artifacts successfully", len(artifactsList))
	m.systemMetrics.listSuccessCounter.Inc(ctx)
	return &datacatalog.ListArtifactsResponse{Artifacts: artifactsList, NextToken: token}, nil
}

// UpdateArtifact updates the given artifact, currently only allowing the associated ArtifactData to be replaced. All
// stored data will be overwritten in the underlying blob storage, no longer existing data (based on ArtifactData name)
// will be deleted.
func (m *artifactManager) UpdateArtifact(ctx context.Context, request *datacatalog.UpdateArtifactRequest) (*datacatalog.UpdateArtifactResponse, error) {
	ctx = contextutils.WithProjectDomain(ctx, request.Dataset.Project, request.Dataset.Domain)

	timer := m.systemMetrics.updateResponseTime.Start(ctx)
	defer timer.Stop()

	err := validators.ValidateUpdateArtifactRequest(request)
	if err != nil {
		logger.Warningf(ctx, "Invalid update artifact request %v, err: %v", request, err)
		m.systemMetrics.validationErrorCounter.Inc(ctx)
		m.systemMetrics.updateFailureCounter.Inc(ctx)
		return nil, err
	}

	// artifact must already exist, verify first
	artifactModel, err := m.findArtifact(ctx, request.GetDataset(), request)
	if err != nil {
		logger.Errorf(ctx, "Failed to get artifact for update artifact request %v, err: %v", request, err)
		m.systemMetrics.updateFailureCounter.Inc(ctx)
		return nil, err
	}

	artifact, err := transformers.FromArtifactModel(artifactModel)
	if err != nil {
		logger.Errorf(ctx, "Error in transforming update artifact request %+v, err %v", artifactModel, err)
		m.systemMetrics.transformerErrorCounter.Inc(ctx)
		m.systemMetrics.updateFailureCounter.Inc(ctx)
		return nil, err
	}

	// overwrite existing artifact data and upload new entries, building a map of artifact data names to remove
	// deleted entries from the blob storage after the upload completed
	artifactDataNames := make(map[string]struct{})
	artifactDataModels := make([]models.ArtifactData, len(request.Data))
	for i, artifactData := range request.Data {
		artifactDataNames[artifactData.Name] = struct{}{}

		dataLocation, err := m.artifactStore.PutData(ctx, artifact, artifactData)
		if err != nil {
			logger.Errorf(ctx, "Failed to store artifact data [%v] during update, err: %v", artifactData.Name, err)
			m.systemMetrics.updateDataFailureCounter.Inc(ctx)
			m.systemMetrics.updateFailureCounter.Inc(ctx)
			return nil, err
		}

		artifactDataModels[i].Name = artifactData.Name
		artifactDataModels[i].Location = dataLocation.String()
		m.systemMetrics.updateDataSuccessCounter.Inc(ctx)
	}

	removedArtifactData := make([]models.ArtifactData, 0)
	for _, artifactData := range artifactModel.ArtifactData {
		if _, ok := artifactDataNames[artifactData.Name]; !ok {
			removedArtifactData = append(removedArtifactData, artifactData)
		}
	}

	// update artifact in DB, also replaces/upserts associated artifact data
	artifactModel.ArtifactData = artifactDataModels
	err = m.repo.ArtifactRepo().Update(ctx, artifactModel)
	if err != nil {
		if errors.IsDoesNotExistError(err) {
			logger.Warnf(ctx, "Artifact does not exist key: %+v, err %v", artifact.Id, err)
			m.systemMetrics.doesNotExistCounter.Inc(ctx)
		} else {
			logger.Errorf(ctx, "Failed to update artifact %v, err: %v", artifactModel, err)
		}
		m.systemMetrics.updateFailureCounter.Inc(ctx)
		return nil, err
	}

	// delete all artifact data no longer present in the updated artifact from the blob storage.
	// blob storage data is removed last in case the DB update fail, which would leave us with artifact data DB entries
	// without underlying blob data. this might still leave orphaned data in blob storage, however we can more easily
	// clean that up periodically and don't risk serving artifact data records that will fail when retrieved.
	for _, artifactData := range removedArtifactData {
		if err := m.artifactStore.DeleteData(ctx, artifactData); err != nil {
			logger.Errorf(ctx, "Failed to delete artifact data during update, err: %v", err)
			m.systemMetrics.deleteDataFailureCounter.Inc(ctx)
			m.systemMetrics.updateFailureCounter.Inc(ctx)
			return nil, err
		}

		m.systemMetrics.deleteDataSuccessCounter.Inc(ctx)
	}

	logger.Debugf(ctx, "Successfully updated artifact id: %v", artifact.Id)

	m.systemMetrics.updateSuccessCounter.Inc(ctx)
	return &datacatalog.UpdateArtifactResponse{
		ArtifactId: artifact.Id,
	}, nil
}

// DeleteArtifact deletes the given artifact, removing all stored artifact data from the underlying blob storage.
func (m *artifactManager) DeleteArtifact(ctx context.Context, request *datacatalog.DeleteArtifactRequest) (*datacatalog.DeleteArtifactResponse, error) {
	ctx = contextutils.WithProjectDomain(ctx, request.Dataset.Project, request.Dataset.Domain)

	timer := m.systemMetrics.deleteResponseTime.Start(ctx)
	defer timer.Stop()

	err := validators.ValidateDeleteArtifactRequest(request)
	if err != nil {
		logger.Warningf(ctx, "Invalid delete artifact request %v, err: %v", request, err)
		m.systemMetrics.validationErrorCounter.Inc(ctx)
		m.systemMetrics.deleteFailureCounter.Inc(ctx)
		return nil, err
	}

	// artifact must already exist, verify first
	artifactModel, err := m.findArtifact(ctx, request.GetDataset(), request)
	if err != nil {
		logger.Errorf(ctx, "Failed to get artifact for delete artifact request %v, err: %v", request, err)
		m.systemMetrics.deleteFailureCounter.Inc(ctx)
		return nil, err
	}

	// delete all artifact data from the blob storage
	for _, artifactData := range artifactModel.ArtifactData {
		if err := m.artifactStore.DeleteData(ctx, artifactData); err != nil {
			logger.Errorf(ctx, "Failed to delete artifact data [%v] during delete, err: %v", artifactData.Name, err)
			m.systemMetrics.deleteDataFailureCounter.Inc(ctx)
			m.systemMetrics.deleteFailureCounter.Inc(ctx)
			return nil, err
		}

		m.systemMetrics.deleteDataSuccessCounter.Inc(ctx)
	}

	// delete artifact from DB, also removed associated artifact data entries
	err = m.repo.ArtifactRepo().Delete(ctx, artifactModel.ArtifactKey)
	if err != nil {
		if errors.IsDoesNotExistError(err) {
			logger.Warnf(ctx, "Artifact does not exist key: %+v, err %v", artifactModel.ArtifactID, err)
			m.systemMetrics.doesNotExistCounter.Inc(ctx)
		} else {
			logger.Errorf(ctx, "Failed to delete artifact %v, err: %v", artifactModel, err)
		}
		m.systemMetrics.deleteFailureCounter.Inc(ctx)
		return nil, err
	}

	logger.Debugf(ctx, "Successfully deleted artifact id: %v", artifactModel.ArtifactID)

	m.systemMetrics.deleteSuccessCounter.Inc(ctx)
	return &datacatalog.DeleteArtifactResponse{}, nil
}

func NewArtifactManager(repo repositories.RepositoryInterface, store *storage.DataStore, storagePrefix storage.DataReference, artifactScope promutils.Scope) interfaces.ArtifactManager {
	artifactMetrics := artifactMetrics{
		scope:                    artifactScope,
		createResponseTime:       labeled.NewStopWatch("create_duration", "The duration of the create artifact calls.", time.Millisecond, artifactScope, labeled.EmitUnlabeledMetric),
		getResponseTime:          labeled.NewStopWatch("get_duration", "The duration of the get artifact calls.", time.Millisecond, artifactScope, labeled.EmitUnlabeledMetric),
		createSuccessCounter:     labeled.NewCounter("create_success_count", "The number of times create artifact succeeded", artifactScope, labeled.EmitUnlabeledMetric),
		getSuccessCounter:        labeled.NewCounter("get_success_count", "The number of times get artifact succeeded", artifactScope, labeled.EmitUnlabeledMetric),
		createFailureCounter:     labeled.NewCounter("create_failure_count", "The number of times create artifact failed", artifactScope, labeled.EmitUnlabeledMetric),
		getFailureCounter:        labeled.NewCounter("get_failure_count", "The number of times get artifact failed", artifactScope, labeled.EmitUnlabeledMetric),
		createDataFailureCounter: labeled.NewCounter("create_data_failure_count", "The number of times create artifact data failed", artifactScope, labeled.EmitUnlabeledMetric),
		createDataSuccessCounter: labeled.NewCounter("create_data_success_count", "The number of times create artifact data succeeded", artifactScope, labeled.EmitUnlabeledMetric),
		transformerErrorCounter:  labeled.NewCounter("transformer_failed_count", "The number of times transformations failed", artifactScope, labeled.EmitUnlabeledMetric),
		validationErrorCounter:   labeled.NewCounter("validation_failed_count", "The number of times validation failed", artifactScope, labeled.EmitUnlabeledMetric),
		alreadyExistsCounter:     labeled.NewCounter("already_exists_count", "The number of times an artifact already exists", artifactScope, labeled.EmitUnlabeledMetric),
		doesNotExistCounter:      labeled.NewCounter("does_not_exists_count", "The number of times an artifact was not found", artifactScope, labeled.EmitUnlabeledMetric),
		listSuccessCounter:       labeled.NewCounter("list_success_count", "The number of times list artifact succeeded", artifactScope, labeled.EmitUnlabeledMetric),
		listFailureCounter:       labeled.NewCounter("list_failure_count", "The number of times list artifact failed", artifactScope, labeled.EmitUnlabeledMetric),
		updateResponseTime:       labeled.NewStopWatch("update_duration", "The duration of the update artifact calls.", time.Millisecond, artifactScope, labeled.EmitUnlabeledMetric),
		updateSuccessCounter:     labeled.NewCounter("update_success_count", "The number of times update artifact succeeded", artifactScope, labeled.EmitUnlabeledMetric),
		updateFailureCounter:     labeled.NewCounter("update_failure_count", "The number of times update artifact failed", artifactScope, labeled.EmitUnlabeledMetric),
		updateDataSuccessCounter: labeled.NewCounter("update_data_success_count", "The number of times update artifact data succeeded", artifactScope, labeled.EmitUnlabeledMetric),
		updateDataFailureCounter: labeled.NewCounter("update_data_failure_count", "The number of times update artifact data failed", artifactScope, labeled.EmitUnlabeledMetric),
		deleteDataSuccessCounter: labeled.NewCounter("delete_data_success_count", "The number of times delete artifact data succeeded", artifactScope, labeled.EmitUnlabeledMetric),
		deleteDataFailureCounter: labeled.NewCounter("delete_data_failure_count", "The number of times delete artifact data failed", artifactScope, labeled.EmitUnlabeledMetric),
		deleteResponseTime:       labeled.NewStopWatch("delete_duration", "The duration of the delete artifact calls.", time.Millisecond, artifactScope, labeled.EmitUnlabeledMetric),
		deleteSuccessCounter:     labeled.NewCounter("delete_success_count", "The number of times delete artifact succeeded", artifactScope, labeled.EmitUnlabeledMetric),
		deleteFailureCounter:     labeled.NewCounter("delete_failure_count", "The number of times delete artifact failed", artifactScope, labeled.EmitUnlabeledMetric),
	}

	return &artifactManager{
		repo:          repo,
		artifactStore: NewArtifactDataStore(store, storagePrefix),
		systemMetrics: artifactMetrics,
	}
}
