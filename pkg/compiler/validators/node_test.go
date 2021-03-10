package validators

import (
	"testing"

	"github.com/flyteorg/flytepropeller/pkg/compiler/common"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytepropeller/pkg/compiler/common/mocks"
	"github.com/flyteorg/flytepropeller/pkg/compiler/errors"
	"github.com/stretchr/testify/assert"
)

func TestValidateBranchNode(t *testing.T) {
	t.Run("No Case No Default", func(t *testing.T) {
		n := &mocks.NodeBuilder{}
		n.OnGetBranchNode().Return(&core.BranchNode{
			IfElse: &core.IfElseBlock{},
		})
		n.OnGetId().Return("node1")

		wf := &mocks.WorkflowBuilder{}
		errs := errors.NewCompileErrors()
		_, _, ok := ValidateBranchNode(wf, n, false, errs)
		assert.False(t, ok)
		if !errs.HasErrors() {
			assert.Error(t, errs)
		} else {
			t.Log(errs)
			errsList := errs.Errors().List()
			assert.Len(t, errsList, 2)
			assert.Equal(t, errors.BranchNodeHasNoCondition, errsList[0].Code())
			assert.Equal(t, errors.BranchNodeHasNoDefault, errsList[1].Code())
		}
	})
}

func TestValidateNode(t *testing.T) {
	n := &mocks.NodeBuilder{}
	n.OnGetId().Return(common.StartNodeID)

	wf := &mocks.WorkflowBuilder{}
	errs := errors.NewCompileErrors()
	ValidateNode(wf, n, true, errs)
	if errs.HasErrors() {
		assert.NoError(t, errs)
	}
}
