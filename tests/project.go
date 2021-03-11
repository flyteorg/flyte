// +build integration

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

func TestCreateProject(t *testing.T) {
	truncateAllTablesForTestingOnly()
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

func TestUpdateProjectDescription(t *testing.T) {
	truncateAllTablesForTestingOnly()
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	// Create a new project.
	req := admin.ProjectRegisterRequest{
		Project: &admin.Project{
			Id:   "potato",
			Name: "spud",
			Labels: &admin.Labels{
				Values: map[string]string{
					"foo": "bar",
					"bar": "baz",
				},
			},
		},
	}
	_, err := client.RegisterProject(ctx, &req)
	assert.Nil(t, err)

	// Verify the project has been registered.
	projects, err := client.ListProjects(ctx, &admin.ProjectListRequest{})
	assert.Nil(t, err)
	assert.NotEmpty(t, projects.Projects)

	// Attempt to modify the name of the Project. Labels should be a no-op.
	// Name and Description should modify just fine.
	_, err = client.UpdateProject(ctx, &admin.Project{
		Id:          "potato",
		Name:        "foobar",
		Description: "a-new-description",
	})

	// Assert that update went through without an error.
	assert.Nil(t, err)

	// Fetch updated projects.
	projectsUpdated, err := client.ListProjects(ctx, &admin.ProjectListRequest{})
	assert.Nil(t, err)
	assert.NotEmpty(t, projectsUpdated.Projects)

	// Verify that the project's Name has not been modified but the Description has.
	updatedProject := projectsUpdated.Projects[0]
	assert.Equal(t, updatedProject.Id, "potato")                     // unchanged
	assert.Equal(t, updatedProject.Name, "foobar")                   // changed
	assert.Equal(t, updatedProject.Description, "a-new-description") // changed

	// Verify that project labels are not removed.
	labelsMap := updatedProject.Labels
	fooVal, fooExists := labelsMap.Values["foo"]
	barVal, barExists := labelsMap.Values["bar"]
	assert.Equal(t, fooExists, true)
	assert.Equal(t, fooVal, "bar")
	assert.Equal(t, barExists, true)
	assert.Equal(t, barVal, "baz")
}

func TestUpdateProjectLabels(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	// Create a new project.
	req := admin.ProjectRegisterRequest{
		Project: &admin.Project{
			Id:   "potato",
			Name: "spud",
		},
	}
	_, err := client.RegisterProject(ctx, &req)
	assert.Nil(t, err)

	// Verify the project has been registered.
	projects, err := client.ListProjects(ctx, &admin.ProjectListRequest{})
	assert.Nil(t, err)
	assert.NotEmpty(t, projects.Projects)

	// Attempt to modify the name of the Project. Labels and name should be
	// modified.
	_, err = client.UpdateProject(ctx, &admin.Project{
		Id:   "potato",
		Name: "foobar",
		Labels: &admin.Labels{
			Values: map[string]string{
				"foo": "bar",
				"bar": "baz",
			},
		},
	})

	// Assert that update went through without an error.
	assert.Nil(t, err)

	// Fetch updated projects.
	projectsUpdated, err := client.ListProjects(ctx, &admin.ProjectListRequest{})
	assert.Nil(t, err)
	assert.NotEmpty(t, projectsUpdated.Projects)

	// Check the name has been modified.
	// Verify that the project's Name has not been modified but the Description has.
	updatedProject := projectsUpdated.Projects[0]
	assert.Equal(t, updatedProject.Id, "potato")   // unchanged
	assert.Equal(t, updatedProject.Name, "foobar") // changed

	// Verify that the expected labels have been added to the project.
	labelsMap := updatedProject.Labels
	fooVal, fooExists := labelsMap.Values["foo"]
	barVal, barExists := labelsMap.Values["bar"]
	assert.Equal(t, fooExists, true)
	assert.Equal(t, fooVal, "bar")
	assert.Equal(t, barExists, true)
	assert.Equal(t, barVal, "baz")
}

func TestUpdateProjectLabels_BadLabels(t *testing.T) {
	ctx := context.Background()
	client, conn := GetTestAdminServiceClient()
	defer conn.Close()

	// Create a new project.
	req := admin.ProjectRegisterRequest{
		Project: &admin.Project{
			Id:   "potato",
			Name: "spud",
		},
	}
	_, err := client.RegisterProject(ctx, &req)
	assert.Nil(t, err)

	// Verify the project has been registered.
	projects, err := client.ListProjects(ctx, &admin.ProjectListRequest{})
	assert.Nil(t, err)
	assert.NotEmpty(t, projects.Projects)

	// Attempt to modify the name of the Project. Labels and name should be
	// modified.
	_, err = client.UpdateProject(ctx, &admin.Project{
		Id:   "potato",
		Name: "foobar",
		Labels: &admin.Labels{
			Values: map[string]string{
				"foo": "#bar",
				"bar": "baz",
			},
		},
	})

	// Assert that update went through without an error.
	assert.EqualError(t, err, "invalid label value [#bar]: [a DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')]")
}
