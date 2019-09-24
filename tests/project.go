// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

func TestCreateProject(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	req := admin.ProjectRegisterRequest{
		Project: &admin.Project{
			Id:   "potato",
			Name: "spud",
		},
	}

	_, err := client.RegisterProject(ctx, &req)
	assert.Nil(t, err)

	projects, err := client.ListProjects(ctx, &admin.ProjectListRequest{})
	assert.Nil(t, err)
	assert.NotEmpty(t, projects.Projects)
	var sawNewProject bool
	for _, project := range projects.Projects {
		if project.Id == "potato" {
			sawNewProject = true
			assert.Equal(t, "spud", project.Name)
		}
		// Assert all projects include expected system-wide domains
		for _, domain := range project.Domains {
			assert.Contains(t, []string{"development", "domain", "staging", "production"}, domain.Id)
			assert.Contains(t, []string{"development", "domain", "staging", "production"}, domain.Name)
		}
	}
	assert.True(t, sawNewProject)
}
