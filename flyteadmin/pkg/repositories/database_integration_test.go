//go:build integration

package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteadmin/pkg/config"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
)

func TestInsertionOfTask(t *testing.T) {
	// Because this is an integration test, we'll need to load the version of configuration that's actually being run
	// instead of the sample one in this repo
	config.Init("/etc/flyte/config/flyteadmin_config.yaml")

	repository := GetRepository(POSTGRES)
	taskRepository := repository.TaskRepo()

	err := taskRepository.Create(models.Task{
		Name:    "sometesttask",
		Closure: []byte("in bytes"),
		Domain:  "testdev",
		Project: "flyte",
		Version: "0.0.0",
	})

	assert.NoError(t, err)
}
