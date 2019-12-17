package impl

import (
	"context"
	"strconv"
	"time"

	"github.com/lyft/datacatalog/pkg/errors"
	"github.com/lyft/datacatalog/pkg/manager/impl/validators"
	"github.com/lyft/datacatalog/pkg/manager/interfaces"
	"github.com/lyft/datacatalog/pkg/repositories"
	datacatalog "github.com/lyft/datacatalog/protos/gen"

	"github.com/lyft/datacatalog/pkg/repositories/models"
	"github.com/lyft/datacatalog/pkg/repositories/transformers"

	"github.com/lyft/datacatalog/pkg/common"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/lyft/flytestdlib/storage"
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
}

type artifactManager struct {
	repo          repositories.RepositoryInterface
	artifactStore ArtifactDataStore
	systemMetrics artifactMetrics
}

// Create an Artifact along with the associated ArtifactData. The ArtifactData will be stored in an offloaded location.
func (m *artifactManager) CreateArtifact(ctx context.Context, request datacatalog.CreateArtifactRequest) (*datacatalog.CreateArtifactResponse, error) {
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
	datasetKey := transformers.FromDatasetID(*artifact.Dataset)

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
		dataLocation, err := m.artifactStore.PutData(ctx, *artifact, *artifactData)
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
func (m *artifactManager) GetArtifact(ctx context.Context, request datacatalog.GetArtifactRequest) (*datacatalog.GetArtifactResponse, error) {
	timer := m.systemMetrics.getResponseTime.Start(ctx)
	defer timer.Stop()

	err := validators.ValidateGetArtifactRequest(request)
	if err != nil {
		logger.Warningf(ctx, "Invalid get artifact request %v, err: %v", request, err)
		m.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	datasetID := request.Dataset

	var artifactModel models.Artifact
	switch request.QueryHandle.(type) {
	case *datacatalog.GetArtifactRequest_ArtifactId:
		logger.Debugf(ctx, "Get artifact by id %v", request.GetArtifactId())
		artifactKey := transformers.ToArtifactKey(datasetID, request.GetArtifactId())
		artifactModel, err = m.repo.ArtifactRepo().Get(ctx, artifactKey)

		if err != nil {
			if errors.IsDoesNotExistError(err) {
				logger.Warnf(ctx, "Artifact does not exist id: %+v, err %v", request.GetArtifactId(), err)
				m.systemMetrics.doesNotExistCounter.Inc(ctx)
			} else {
				logger.Errorf(ctx, "Unable to retrieve artifact by id: %+v, err %v", request.GetArtifactId(), err)
				m.systemMetrics.getFailureCounter.Inc(ctx)
			}
			return nil, err
		}
	case *datacatalog.GetArtifactRequest_TagName:
		logger.Debugf(ctx, "Get artifact by tag %v", request.GetTagName())
		tagKey := transformers.ToTagKey(*datasetID, request.GetTagName())
		tag, err := m.repo.TagRepo().Get(ctx, tagKey)

		if err != nil {
			if errors.IsDoesNotExistError(err) {
				logger.Warnf(ctx, "Artifact does not exist tag: %+v, err %v", request.GetTagName(), err)
				m.systemMetrics.doesNotExistCounter.Inc(ctx)
			} else {
				logger.Errorf(ctx, "Unable to retrieve Artifact by tag %v, err: %v", request.GetTagName(), err)
				m.systemMetrics.getFailureCounter.Inc(ctx)
			}
			return nil, err
		}

		artifactModel = tag.Artifact
	}

	if len(artifactModel.ArtifactData) == 0 {
		return nil, errors.NewDataCatalogErrorf(codes.Internal, "artifact [%+v] does not have artifact data associated", request)
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
		Artifact: &artifact,
	}, nil
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

func (m *artifactManager) ListArtifacts(ctx context.Context, request datacatalog.ListArtifactsRequest) (*datacatalog.ListArtifactsResponse, error) {
	err := validators.ValidateListArtifactRequest(&request)
	if err != nil {
		logger.Warningf(ctx, "Invalid list artifact request %v, err: %v", request, err)
		m.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	// Verify the dataset exists before listing artifacts
	datasetKey := transformers.FromDatasetID(*request.Dataset)
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

	token := strconv.Itoa(int(listInput.Offset) + len(artifactsList))

	logger.Debugf(ctx, "Listed %v matching artifacts successfully", len(artifactsList))
	m.systemMetrics.listSuccessCounter.Inc(ctx)
	return &datacatalog.ListArtifactsResponse{Artifacts: artifactsList, NextToken: token}, nil
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
	}

	return &artifactManager{
		repo:          repo,
		artifactStore: NewArtifactDataStore(store, storagePrefix),
		systemMetrics: artifactMetrics,
	}
}
