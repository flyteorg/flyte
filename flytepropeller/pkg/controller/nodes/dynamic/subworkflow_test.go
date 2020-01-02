package dynamic

import (
	"context"
	"testing"

	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
)

func TestNewContextualWorkflow(t *testing.T) {
	wf := &mocks.ExecutableWorkflow{}
	calledBase := false
	wf.On("GetAnnotations").Return(map[string]string{}).Run(func(_ mock.Arguments) {
		calledBase = true
	})

	wf.On("GetExecutionStatus").Return(&mocks.ExecutableWorkflowStatus{})

	subwf := &mocks.ExecutableSubWorkflow{}
	cWF := newContextualWorkflow(wf, subwf, nil, nil, nil, nil)
	cWF.GetAnnotations()

	assert.True(t, calledBase)
}

func TestConstructNodeDataDir(t *testing.T) {
	wf := &mocks.ExecutableWorkflow{}
	wf.On("GetExecutionStatus").Return(&mocks.ExecutableWorkflowStatus{})

	wfStatus := &mocks.ExecutableWorkflowStatus{}
	wfStatus.On("GetDataDir").Return(storage.DataReference("fk://wrong/")).Run(func(_ mock.Arguments) {
		assert.FailNow(t, "Should call the override")
	})

	nodeStatus := &mocks.ExecutableNodeStatus{}
	nodeStatus.On("GetDataDir").Return(storage.DataReference("fk://right/"))

	ds, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)
	cWF := newContextualWorkflowStatus(wfStatus, nodeStatus, ds)

	dataDir, err := cWF.ConstructNodeDataDir(context.TODO(), "my_node")
	assert.NoError(t, err)
	assert.NotNil(t, dataDir)
	assert.Equal(t, "fk://right/my_node", dataDir.String())
}
