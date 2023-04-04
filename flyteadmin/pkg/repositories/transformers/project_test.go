package transformers

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestCreateProjectModel(t *testing.T) {
	labels := admin.Labels{
		Values: map[string]string{
			"foo": "#badlabel",
		},
	}
	project := admin.Project{
		Id:          "project_id",
		Name:        "project_name",
		Description: "project_description",
		Labels:      &labels,
		State:       admin.Project_ACTIVE,
	}

	projectBytes, _ := proto.Marshal(&project)
	projectModel := CreateProjectModel(&project)

	activeState := int32(admin.Project_ACTIVE)
	assert.Equal(t, models.Project{
		Identifier:  "project_id",
		Name:        "project_name",
		Description: "project_description",
		Labels:      projectBytes,
		State:       &activeState,
	}, projectModel)
}

func TestFromProjectModel(t *testing.T) {
	activeState := int32(admin.Project_ACTIVE)
	projectModel := models.Project{
		Identifier:  "proj_id",
		Name:        "proj_name",
		Description: "proj_description",
		State:       &activeState,
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
		Id:          "proj_id",
		Name:        "proj_name",
		Description: "proj_description",
		Domains:     domains,
		State:       admin.Project_ACTIVE,
	}, &project))
}

func TestFromProjectModels(t *testing.T) {
	activeState := int32(admin.Project_ACTIVE)
	projectModels := []models.Project{
		{
			Identifier:  "proj1_id",
			Name:        "proj1_name",
			Description: "proj1_description",
			State:       &activeState,
		},
		{
			Identifier:  "proj2_id",
			Name:        "proj2_name",
			Description: "proj2_description",
			State:       &activeState,
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
		assert.Equal(t, fmt.Sprintf("proj%v_description", index+1), project.Description)
		assert.Equal(t, admin.Project_ACTIVE, project.State)
		assert.EqualValues(t, domains, project.Domains)
	}
}
