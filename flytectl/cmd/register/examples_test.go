package register

import (
	"testing"

	"github.com/flyteorg/flyte/flytectl/cmd/testutils"
	"github.com/stretchr/testify/assert"
)

func TestRegisterExamplesFunc(t *testing.T) {
	s := testutils.Setup(t)

	registerFilesSetup()
	args := []string{""}
	err := registerExamplesFunc(s.Ctx, args, s.CmdCtx)
	assert.NotNil(t, err)
}
func TestRegisterExamplesFuncErr(t *testing.T) {
	s := testutils.Setup(t)

	registerFilesSetup()
	flytesnacks = "testingsnacks"
	args := []string{""}
	err := registerExamplesFunc(s.Ctx, args, s.CmdCtx)
	// TODO (Yuvraj) make test to success after fixing flytesnacks bug
	assert.NotNil(t, err)
	flytesnacks = "flytesnacks"
}
