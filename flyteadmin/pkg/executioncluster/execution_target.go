package executioncluster

import (
	"k8s.io/client-go/dynamic"
	restclient "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	flyteclient "github.com/flyteorg/flyte/flytepropeller/pkg/client/clientset/versioned"
	"github.com/flyteorg/flyte/flytestdlib/random"
)

// Spec to determine the execution target
type ExecutionTargetSpec struct {
	TargetID              string
	ExecutionID           string
	Project               string
	Domain                string
	Workflow              string
	LaunchPlan            string
	ExecutionClusterLabel *admin.ExecutionClusterLabel
}

// Client object of the target execution cluster
type ExecutionTarget struct {
	ID            string
	FlyteClient   flyteclient.Interface
	Client        client.Client
	DynamicClient dynamic.Interface
	Enabled       bool
	Config        restclient.Config
}

func (e ExecutionTarget) Compare(to random.Comparable) bool {
	return e.ID < to.(ExecutionTarget).ID
}
