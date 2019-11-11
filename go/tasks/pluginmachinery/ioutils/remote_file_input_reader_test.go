package ioutils

import (
	"testing"

	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
)

func TestSimpleInputFilePath_GetInputPath(t *testing.T) {
	s := SimpleInputFilePath{
		pathPrefix: "s3://lyft-modelbuilder/metadata/propeller/staging/flyteexamples-development-jf193q0cqo/odd-nums-task/data",
		store:      storage.URLPathConstructor{},
	}

	assert.Equal(t, "s3://lyft-modelbuilder/metadata/propeller/staging/flyteexamples-development-jf193q0cqo/odd-nums-task/data/inputs.pb", s.GetInputPath().String())
}
