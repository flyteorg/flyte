package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitializeClientFunc(t *testing.T) {
	cs := initializeClientFunc()
	assert.NotNil(t, cs)
	assert.NotNil(t, cs.getAgentClient)
	assert.NotNil(t, cs.getAgentMetadataClient)
}
