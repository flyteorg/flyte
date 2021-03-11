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

	projectModel := CreateProjectModel(&admin.Project{
		Id:          "project_id",
		Name:        "project_name",
		Description: "project_description",
		State:       admin.Project_ACTIVE,
	})

	activeState := int32(admin.Project_ACTIVE)
	assert.Equal(t, models.Project{
		Identifier:  "project_id",
		Name:        "project_name",
		Description: "project_description",
		Labels: []uint8{
			0xa, 0xa, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x5f, 0x69, 0x64, 0x12, 0xc, 0x70, 0x72, 0x6f,
			0x6a, 0x65, 0x63, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x13, 0x70, 0x72, 0x6f, 0x6a, 0x65,
			0x63, 0x74, 0x5f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e,
		},
		State: &activeState,
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
