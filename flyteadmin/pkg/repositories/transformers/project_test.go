package transformers

import (
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

func TestCreateProjectModel(t *testing.T) {

	projectModel := CreateProjectModel("project_id", "project_name")

	assert.Equal(t, models.Project{
		Identifier: "project_id",
		Name:       "project_name",
	}, projectModel)
}

func TestFromProjectModel(t *testing.T) {
	projectModel := models.Project{
		Identifier: "proj_id",
		Name:       "proj_name",
	}
	domains := []*admin.Domain{
		{
			Id:   "domain_id",
			Name: "domain_name",
		},
		{
			Id:   "domain2_id",
			Name: "domain2_name",
		},
	}
	project := FromProjectModel(projectModel, domains)
	assert.True(t, proto.Equal(&admin.Project{
		Id:      "proj_id",
		Name:    "proj_name",
		Domains: domains,
	}, &project))
}

func TestFromProjectModels(t *testing.T) {
	projectModels := []models.Project{
		{
			Identifier: "proj1_id",
			Name:       "proj1_name",
		},
		{
			Identifier: "proj2_id",
			Name:       "proj2_name",
		},
	}
	domains := []*admin.Domain{
		{
			Id:   "domain_id",
			Name: "domain_name",
		},
		{
			Id:   "domain2_id",
			Name: "domain2_name",
		},
	}
	projects := FromProjectModels(projectModels, domains)
	assert.Len(t, projects, 2)
	for index, project := range projects {
		assert.Equal(t, fmt.Sprintf("proj%v_id", index+1), project.Id)
		assert.Equal(t, fmt.Sprintf("proj%v_name", index+1), project.Name)
		assert.EqualValues(t, domains, project.Domains)
	}
}
