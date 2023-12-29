package ioutils

import (
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func TestSimpleInputFilePath_GetInputPath(t *testing.T) {
	dataStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	s := SimpleInputFilePath{
		pathPrefix: "s3://flyteorg-modelbuilder/metadata/propeller/staging/flyteexamples-development-jf193q0cqo/odd-nums-task/data",
		store:      dataStore,
	}

	assert.Equal(t, "s3://flyteorg-modelbuilder/metadata/propeller/staging/flyteexamples-development-jf193q0cqo/odd-nums-task/data/inputs.pb", s.GetInputDataPath().String())
}
