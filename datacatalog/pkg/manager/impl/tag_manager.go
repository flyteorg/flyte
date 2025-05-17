package impl

import (
	"context"
	"time"

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

type tagMetrics struct {
	scope                  promutils.Scope
	createResponseTime     labeled.StopWatch
	addTagSuccessCounter   labeled.Counter
	addTagFailureCounter   labeled.Counter
	validationErrorCounter labeled.Counter
	alreadyExistsCounter   labeled.Counter
}

type tagManager struct {
	repo          repositories.RepositoryInterface
	store         *storage.DataStore
	systemMetrics tagMetrics
}

func (m *tagManager) AddTag(ctx context.Context, request *datacatalog.AddTagRequest) (*datacatalog.AddTagResponse, error) {
	timer := m.systemMetrics.createResponseTime.Start(ctx)
	defer timer.Stop()

	if err := validators.ValidateTag(request.GetTag()); err != nil {
		logger.Warnf(ctx, "Invalid get tag request %+v err: %v", request, err)
		m.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	// verify the artifact and dataset exists before adding a tag to it
	datasetID := request.GetTag().GetDataset()
	ctx = contextutils.WithProjectDomain(ctx, datasetID.GetProject(), datasetID.GetDomain())

	datasetKey := transformers.FromDatasetID(datasetID)
	dataset, err := m.repo.DatasetRepo().Get(ctx, datasetKey)
	if err != nil {
		m.systemMetrics.addTagFailureCounter.Inc(ctx)
		return nil, err
	}

	artifactKey := transformers.ToArtifactKey(datasetID, request.GetTag().GetArtifactId())
	_, err = m.repo.ArtifactRepo().GetAndFilterExpired(ctx, artifactKey)
	if err != nil {
		m.systemMetrics.addTagFailureCounter.Inc(ctx)
		return nil, err
	}

	tagKey := transformers.ToTagKey(datasetID, request.GetTag().GetName())
	err = m.repo.TagRepo().Create(ctx, models.Tag{
		TagKey:      tagKey,
		ArtifactID:  request.GetTag().GetArtifactId(),
		DatasetUUID: dataset.UUID,
	})
	if err != nil {
		if errors.IsAlreadyExistsError(err) {
			logger.Warnf(ctx, "Tag already exists key: %+v, err %v", request, err)
			m.systemMetrics.alreadyExistsCounter.Inc(ctx)
		} else {
			logger.Errorf(ctx, "Failed to tag artifact: %+v err: %v", request, err)
			m.systemMetrics.addTagFailureCounter.Inc(ctx)
		}

		return nil, err
	}

	m.systemMetrics.addTagSuccessCounter.Inc(ctx)
	return &datacatalog.AddTagResponse{}, nil
}

func NewTagManager(repo repositories.RepositoryInterface, store *storage.DataStore, tagScope promutils.Scope) interfaces.TagManager {
	systemMetrics := tagMetrics{
		scope:                  tagScope,
		createResponseTime:     labeled.NewStopWatch("create_duration", "The duration of the add tag calls.", time.Millisecond, tagScope, labeled.EmitUnlabeledMetric),
		addTagSuccessCounter:   labeled.NewCounter("create_success_count", "The number of times an artifact was tagged successfully", tagScope, labeled.EmitUnlabeledMetric),
		addTagFailureCounter:   labeled.NewCounter("create_failure_count", "The number of times we failed  to tag an artifact", tagScope, labeled.EmitUnlabeledMetric),
		validationErrorCounter: labeled.NewCounter("validation_failed_count", "The number of times we failed validate a tag", tagScope, labeled.EmitUnlabeledMetric),
		alreadyExistsCounter:   labeled.NewCounter("already_exists_count", "The number of times an tag already exists", tagScope, labeled.EmitUnlabeledMetric),
	}

	return &tagManager{
		repo:          repo,
		store:         store,
		systemMetrics: systemMetrics,
	}
}
