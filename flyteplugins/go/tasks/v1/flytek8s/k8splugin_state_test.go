package flytek8s

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"
	"encoding/json"
)

func TestRetrieveK8sObjectStatus(t *testing.T) {
	status := k8sObjectExists
	phase := types.TaskPhaseRunning
	customState := storeK8sObjectState(status, phase)

	raw, err := json.Marshal(customState)
	assert.NoError(t, err)

	unmarshalledCustomState := make(map[string]interface{})
	err = json.Unmarshal(raw, &unmarshalledCustomState)
	assert.NoError(t, err)

	retrievedStatus, retrievedPhase, err := retrieveK8sObjectState(unmarshalledCustomState)
	assert.NoError(t, err)
	assert.Equal(t, status, retrievedStatus)
	assert.Equal(t, phase, retrievedPhase)
}
