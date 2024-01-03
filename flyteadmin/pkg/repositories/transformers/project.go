package transformers

import (
	flyteErrs "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

type CreateProjectModelInput struct {
	Identifier  string
	Name        string
	Description string
}

func CreateProjectModel(project *admin.Project, org string) (models.Project, error) {
	stateInt := int32(project.State)
	projectModel := models.Project{
		Identifier:  project.Id,
		Name:        project.Name,
		Description: project.Description,
		State:       &stateInt,
		Org:         org,
	}
	if project.Labels != nil {
		projectBytes, err := proto.Marshal(project)
		if err != nil {
			return models.Project{}, flyteErrs.NewFlyteAdminErrorf(codes.Internal, "failed to marshal project: %v", err)
		}
		projectModel.Labels = projectBytes
	}

	return projectModel, nil
}

func FromProjectModel(projectModel models.Project, domains []*admin.Domain) admin.Project {
	projectDeserialized := &admin.Project{}
	err := proto.Unmarshal(projectModel.Labels, projectDeserialized)
	if err != nil {
		return admin.Project{}
	}
	project := admin.Project{
		Id:          projectModel.Identifier,
		Name:        projectModel.Name,
		Description: projectModel.Description,
		Labels:      projectDeserialized.Labels,
		State:       admin.Project_ProjectState(*projectModel.State),
		Org:         projectModel.Org,
	}
	project.Domains = domains
	return project
}

func FromProjectModels(projectModels []models.Project, domains []*admin.Domain) []*admin.Project {
	projects := make([]*admin.Project, len(projectModels))
	for index, projectModel := range projectModels {
		project := FromProjectModel(projectModel, domains)
		projects[index] = &project
	}
	return projects
}
