package transformers

import (
	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

type CreateProjectModelInput struct {
	Identifier  string
	Name        string
	Description string
}

func CreateProjectModel(project *admin.Project) models.Project {
	stateInt := int32(project.GetState())
	if project.GetLabels() == nil {
		return models.Project{
			Identifier:  project.GetId(),
			Name:        project.GetName(),
			Description: project.GetDescription(),
			State:       &stateInt,
		}
	}
	projectBytes, err := proto.Marshal(project)
	if err != nil {
		return models.Project{}
	}
	return models.Project{
		Identifier:  project.GetId(),
		Name:        project.GetName(),
		Description: project.GetDescription(),
		Labels:      projectBytes,
		State:       &stateInt,
	}
}

func FromProjectModel(projectModel models.Project, domains []*admin.Domain) *admin.Project {
	projectDeserialized := &admin.Project{}
	err := proto.Unmarshal(projectModel.Labels, projectDeserialized)
	if err != nil {
		return &admin.Project{}
	}
	project := &admin.Project{
		Id:          projectModel.Identifier,
		Name:        projectModel.Name,
		Description: projectModel.Description,
		Labels:      projectDeserialized.GetLabels(),
		State:       admin.Project_ProjectState(*projectModel.State),
	}
	project.Domains = domains
	return project
}

func FromProjectModels(projectModels []models.Project, domains []*admin.Domain) []*admin.Project {
	projects := make([]*admin.Project, len(projectModels))
	for index, projectModel := range projectModels {
		project := FromProjectModel(projectModel, domains)
		projects[index] = project
	}
	return projects
}
