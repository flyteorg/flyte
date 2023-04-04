package impl

import (
	"context"
	"time"

	"github.com/flyteorg/datacatalog/pkg/manager/impl/validators"
	"github.com/flyteorg/datacatalog/pkg/manager/interfaces"
	"github.com/flyteorg/datacatalog/pkg/repositories"

	"github.com/flyteorg/datacatalog/pkg/repositories/models"
	"github.com/flyteorg/datacatalog/pkg/repositories/transformers"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/datacatalog"

	"github.com/flyteorg/datacatalog/pkg/errors"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flytestdlib/storage"
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

	if err := validators.ValidateTag(request.Tag); err != nil {
		logger.Warnf(ctx, "Invalid get tag request %+v err: %v", request, err)
		m.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	// verify the artifact and dataset exists before adding a tag to it
	datasetID := request.Tag.Dataset
	ctx = contextutils.WithProjectDomain(ctx, datasetID.Project, datasetID.Domain)

	datasetKey := transformers.FromDatasetID(datasetID)
	dataset, err := m.repo.DatasetRepo().Get(ctx, datasetKey)
	if err != nil {
		m.systemMetrics.addTagFailureCounter.Inc(ctx)
		return nil, err
	}

	artifactKey := transformers.ToArtifactKey(datasetID, request.Tag.ArtifactId)
	_, err = m.repo.ArtifactRepo().Get(ctx, artifactKey)
	if err != nil {
		m.systemMetrics.addTagFailureCounter.Inc(ctx)
		return nil, err
	}

	tagKey := transformers.ToTagKey(datasetID, request.Tag.Name)
	err = m.repo.TagRepo().Create(ctx, models.Tag{
		TagKey:      tagKey,
		ArtifactID:  request.Tag.ArtifactId,
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
