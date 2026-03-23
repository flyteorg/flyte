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

	stateInt := int32(p.GetState())

	return &models.Project{
		Identifier:  p.GetId(),
		Name:        p.GetName(),
		Description: p.GetDescription(),
		Labels:      labelsBytes,
		State:       &stateInt,
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

	state := project.ProjectState_PROJECT_STATE_ACTIVE
	if model.State != nil {
		state = project.ProjectState(*model.State)
	}

	return &project.Project{
		Id:          model.Identifier,
		Name:        model.Name,
		Domains:     domains,
		Description: model.Description,
		Labels:      labels,
		State:       state,
	}, nil
}
