package task

import (
	"testing"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

type dummySetupCtx struct {
	interfaces.SetupContext
	testScopeName string
}

func (d dummySetupCtx) MetricsScope() promutils.Scope {
	return promutils.NewScope(d.testScopeName)
}

func Test_nameSpacedSetupCtx_MetricsScope(t *testing.T) {
	r := &mocks.ResourceRegistrar{}
	ns := newNameSpacedSetupCtx(&setupContext{SetupContext: &dummySetupCtx{testScopeName: "test-scope-1"}}, r, "p1")
	scope := ns.MetricsScope()
	assert.NotNil(t, scope)
	assert.Equal(t, "test_scope_1:p1:", scope.CurrentScope())
}
