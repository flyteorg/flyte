package v1alpha1_test

import (
	"io/ioutil"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

func ReadYamlFileAsJSON(path string) ([]byte, error) {
	r, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return yaml.YAMLToJSON(r)
}

func TestWorkflowIsInterruptible(t *testing.T) {
	w := &v1alpha1.FlyteWorkflow{}

	// no execution spec or metadata defined -> interruptible defaults to false
	assert.False(t, w.IsInterruptible())

	// marked as interruptible via execution config (e.g. for a single execution)
	execConfigInterruptible := true
	w.ExecutionConfig.Interruptible = &execConfigInterruptible
	assert.True(t, w.IsInterruptible())

	// marked as not interruptible via execution config, overwriting node defaults
	execConfigInterruptible = false
	w.NodeDefaults.Interruptible = true
	assert.False(t, w.IsInterruptible())

	// marked as interruptible via execution config, overwriting node defaults
	execConfigInterruptible = true
	w.NodeDefaults.Interruptible = false
	assert.True(t, w.IsInterruptible())

	// interruptible flag retrieved from node defaults (e.g. workflow definition), no execution config override
	w.ExecutionConfig.Interruptible = nil
	w.NodeDefaults.Interruptible = true
	assert.True(t, w.IsInterruptible())
}
