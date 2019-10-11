package flytek8s

import (
	"encoding/json"
	"fmt"

	"github.com/lyft/flyteplugins/go/tasks/v1/types"
)

const stateKey = "os"

// This status internal state of the object not read/updated by upstream components (eg. Node manager)
type K8sObjectStatus int

const (
	k8sObjectUnknown K8sObjectStatus = iota
	k8sObjectExists
	k8sObjectDeleted
)

func (q K8sObjectStatus) String() string {
	switch q {
	case k8sObjectUnknown:
		return "NotStarted"
	case k8sObjectExists:
		return "Running"
	case k8sObjectDeleted:
		return "Deleted"
	}
	return "IllegalK8sObjectStatus"
}

// This status internal state of the object not read/updated by upstream components (eg. Node manager)
type K8sObjectState struct {
	Status        K8sObjectStatus `json:"s"`
	TerminalPhase types.TaskPhase `json:"tp"`
}

func retrieveK8sObjectState(customState map[string]interface{}) (K8sObjectStatus, types.TaskPhase, error) {
	v, found := customState[stateKey]
	if !found {
		return k8sObjectUnknown, types.TaskPhaseUnknown, nil
	}

	state, err := convertToState(v)
	if err != nil {
		return k8sObjectUnknown, types.TaskPhaseUnknown, err
	}
	return state.Status, state.TerminalPhase, nil
}

func storeK8sObjectState(status K8sObjectStatus, phase types.TaskPhase) map[string]interface{} {
	customState := make(map[string]interface{})
	customState[stateKey] = K8sObjectState{Status: status, TerminalPhase: phase}
	return customState
}

func convertToState(iface interface{}) (K8sObjectState, error) {
	raw, err := json.Marshal(iface)
	if err != nil {
		return K8sObjectState{}, err
	}

	item := &K8sObjectState{}
	err = json.Unmarshal(raw, item)
	if err != nil {
		return K8sObjectState{}, fmt.Errorf("failed to unmarshal state into K8sObjectState")
	}

	return *item, nil
}
