package project

import (
	"errors"
	"testing"

	"github.com/flyteorg/flytectl/clierrors"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/stretchr/testify/assert"
)

func TestGetProjectSpec(t *testing.T) {
	t.Run("Successful get project spec", func(t *testing.T) {
		c := &ConfigProject{
			Name: "flytesnacks",
		}
		response, err := c.GetProjectSpec("flytesnacks")
		assert.Nil(t, err)
		assert.NotNil(t, response)
	})
	t.Run("Successful get request spec from file", func(t *testing.T) {
		c := &ConfigProject{
			File: "testdata/project.yaml",
		}
		response, err := c.GetProjectSpec("flytesnacks")
		assert.Nil(t, err)
		assert.Equal(t, "flytesnacks", response.Name)
		assert.Equal(t, "flytesnacks test", response.Description)
	})
}

func TestMapToAdminState(t *testing.T) {
	t.Run("Successful mapToAdminState with archive", func(t *testing.T) {
		c := &ConfigProject{
			Archive: true,
		}
		state, err := c.MapToAdminState()
		assert.Nil(t, err)
		assert.Equal(t, admin.Project_ARCHIVED, state)
	})
	t.Run("Successful mapToAdminState with activate", func(t *testing.T) {
		c := &ConfigProject{
			Activate: true,
		}
		state, err := c.MapToAdminState()
		assert.Nil(t, err)
		assert.Equal(t, admin.Project_ACTIVE, state)
	})
	t.Run("Invalid state", func(t *testing.T) {
		c := &ConfigProject{
			Activate: true,
			Archive:  true,
		}
		state, err := c.MapToAdminState()
		assert.NotNil(t, err)
		assert.Equal(t, errors.New(clierrors.ErrInvalidStateUpdate), err)
		assert.Equal(t, admin.Project_ACTIVE, state)
	})
	t.Run("deprecated Flags Test", func(t *testing.T) {
		c := &ConfigProject{
			ActivateProject: true,
		}
		state, err := c.MapToAdminState()
		assert.Nil(t, err)
		assert.Equal(t, admin.Project_ACTIVE, state)
	})
}
