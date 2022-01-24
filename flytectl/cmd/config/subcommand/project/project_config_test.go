package project

import (
	"testing"

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
