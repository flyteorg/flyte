package executioncluster

import (
	flyteclient "github.com/lyft/flytepropeller/pkg/client/clientset/versioned"
	"github.com/lyft/flytestdlib/random"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Spec to determine the execution target
type ExecutionTargetSpec struct {
	TargetID    string
	ExecutionID string
	Project     string
	Domain      string
	Workflow    string
	LaunchPlan  string
}

// Client object of the target execution cluster
type ExecutionTarget struct {
	ID          string
	FlyteClient flyteclient.Interface
	Client      client.Client
	Enabled     bool
}

func (e ExecutionTarget) Compare(to random.Comparable) bool {
	return e.ID < to.(ExecutionTarget).ID
}
