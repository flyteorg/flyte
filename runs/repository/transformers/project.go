package transformers

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// NewProjectModel transforms a project proto into a project DB model.
func NewProjectModel(p *project.Project) (*models.Project, error) {
	var labelsBytes []byte
	if p.GetLabels() != nil {
		var err error
		labelsBytes, err = proto.Marshal(p.GetLabels())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal project labels: %w", err)
		}
	}

	return &models.Project{
		ProjectKey: models.ProjectKey{
			Org: p.GetOrg(),
			ID:  p.GetId(),
		},
		Name:        p.GetName(),
		Description: p.GetDescription(),
		Labels:      labelsBytes,
		State:       int32(p.GetState()),
	}, nil
}

// ProjectModelToProject transforms a project DB model into a project proto with injected domains.
func ProjectModelToProject(model *models.Project, domains []*project.Domain) (*project.Project, error) {
	var labels *task.Labels
	if len(model.Labels) > 0 {
		labels = &task.Labels{}
		if err := proto.Unmarshal(model.Labels, labels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal project labels: %w", err)
		}
	}

	return &project.Project{
		Id:          model.ID,
		Name:        model.Name,
		Domains:     domains,
		Description: model.Description,
		Labels:      labels,
		State:       project.ProjectState(model.State),
		Org:         model.Org,
	}, nil
}
