package impl

import (
	"context"
	"time"

	"github.com/lyft/datacatalog/pkg/manager/impl/validators"
	"github.com/lyft/datacatalog/pkg/manager/interfaces"
	"github.com/lyft/datacatalog/pkg/repositories"

	"github.com/lyft/datacatalog/pkg/repositories/models"
	"github.com/lyft/datacatalog/pkg/repositories/transformers"
	datacatalog "github.com/lyft/datacatalog/protos/gen"

	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/lyft/flytestdlib/storage"
)

type tagMetrics struct {
	scope                  promutils.Scope
	addTagSuccessCounter   labeled.Counter
	addTagFailureCounter   labeled.Counter
	addTagResponseTime     labeled.StopWatch
	validationErrorCounter labeled.Counter
}

type tagManager struct {
	repo          repositories.RepositoryInterface
	store         *storage.DataStore
	systemMetrics tagMetrics
}

func (m *tagManager) AddTag(ctx context.Context, request datacatalog.AddTagRequest) (*datacatalog.AddTagResponse, error) {

	if err := validators.ValidateTag(request.Tag); err != nil {
		m.systemMetrics.validationErrorCounter.Inc(ctx)
		return nil, err
	}

	// verify the artifact exists before adding a tag to it
	datasetID := *request.Tag.Dataset
	artifactKey := transformers.ToArtifactKey(datasetID, request.Tag.ArtifactId)
	_, err := m.repo.ArtifactRepo().Get(ctx, artifactKey)
	if err != nil {
		m.systemMetrics.addTagFailureCounter.Inc(ctx)
		return nil, err
	}

	tagKey := transformers.ToTagKey(datasetID, request.Tag.Name)
	err = m.repo.TagRepo().Create(ctx, models.Tag{
		TagKey:     tagKey,
		ArtifactID: request.Tag.ArtifactId,
	})
	if err != nil {
		m.systemMetrics.addTagFailureCounter.Inc(ctx)
		return nil, err
	}

	m.systemMetrics.addTagSuccessCounter.Inc(ctx)
	return &datacatalog.AddTagResponse{}, nil
}

func NewTagManager(repo repositories.RepositoryInterface, store *storage.DataStore, tagScope promutils.Scope) interfaces.TagManager {
	systemMetrics := tagMetrics{
		scope:                  tagScope,
		addTagResponseTime:     labeled.NewStopWatch("add_tag_duration", "The duration of tagging an artifact.", time.Millisecond, tagScope, labeled.EmitUnlabeledMetric),
		addTagSuccessCounter:   labeled.NewCounter("add_tag_success_count", "The number of times an artifact was tagged successfully", tagScope, labeled.EmitUnlabeledMetric),
		addTagFailureCounter:   labeled.NewCounter("add_tag_failure_count", "The number of times we failed  to tag an artifact", tagScope, labeled.EmitUnlabeledMetric),
		validationErrorCounter: labeled.NewCounter("validation_error_count", "The number of times we failed validate a tag", tagScope, labeled.EmitUnlabeledMetric),
	}

	return &tagManager{
		repo:          repo,
		store:         store,
		systemMetrics: systemMetrics,
	}
}
