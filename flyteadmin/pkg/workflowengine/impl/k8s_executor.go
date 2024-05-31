package impl

import (
	"context"

	"google.golang.org/grpc/codes"
	k8_api_err "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/executioncluster"
	execClusterInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/executioncluster/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/workflowengine/interfaces"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var deletePropagationBackground = v1.DeletePropagationBackground

const defaultIdentifier = "DefaultK8sExecutor"

// K8sWorkflowExecutor directly creates and delete Flyte workflow execution CRD objects using the configured execution
// cluster interface.
type K8sWorkflowExecutor struct {
	config           runtimeInterfaces.Configuration
	executionCluster execClusterInterfaces.ClusterInterface
	workflowBuilder  interfaces.FlyteWorkflowBuilder
	storageClient    *storage.DataStore
}

func (e K8sWorkflowExecutor) ID() string {
	return defaultIdentifier
}

func (e K8sWorkflowExecutor) Execute(ctx context.Context, data interfaces.ExecutionData) (interfaces.ExecutionResponse, error) {
	flyteWf, err := e.workflowBuilder.Build(data.WorkflowClosure, data.ExecutionParameters.Inputs, data.ExecutionID, data.Namespace)
	if err != nil {
		logger.Infof(ctx, "failed to build the workflow [%+v] %v",
			data.WorkflowClosure.Primary.Template.Id, err)
		return interfaces.ExecutionResponse{}, err
	}
	err = PrepareFlyteWorkflow(data, flyteWf)
	if err != nil {
		return interfaces.ExecutionResponse{}, err
	}

	if e.config.ApplicationConfiguration().GetTopLevelConfig().UseOffloadedWorkflowClosure {
		// if offloading workflow closure is enabled we set the WorkflowClosureReference and remove
		// the closure generated static fields from the FlyteWorkflow CRD. They are read from the
		// storage client and temporarily repopulated during execution to reduce the CRD size.
		flyteWf.WorkflowClosureReference = data.WorkflowClosureReference
		flyteWf.WorkflowSpec = nil
		flyteWf.SubWorkflows = nil
		flyteWf.Tasks = nil
	}

	if e.config.ApplicationConfiguration().GetTopLevelConfig().ConsoleURL != "" {
		flyteWf.ConsoleURL = e.config.ApplicationConfiguration().GetTopLevelConfig().ConsoleURL
	}

	executionTargetSpec := executioncluster.ExecutionTargetSpec{
		Project:               data.ExecutionID.Project,
		Domain:                data.ExecutionID.Domain,
		Workflow:              data.ReferenceWorkflowName,
		LaunchPlan:            data.ReferenceWorkflowName,
		ExecutionID:           data.ExecutionID.Name,
		ExecutionClusterLabel: data.ExecutionParameters.ExecutionClusterLabel,
	}
	targetCluster, err := e.executionCluster.GetTarget(ctx, &executionTargetSpec)
	if err != nil {
		return interfaces.ExecutionResponse{}, errors.NewFlyteAdminErrorf(codes.Internal, "failed to create workflow in propeller %v", err)
	}
	_, err = targetCluster.FlyteClient.FlyteworkflowV1alpha1().FlyteWorkflows(data.Namespace).Create(ctx, flyteWf, v1.CreateOptions{})
	if err != nil {
		if !k8_api_err.IsAlreadyExists(err) {
			logger.Debugf(context.TODO(), "Failed to create execution [%+v] in cluster: %s", data.ExecutionID, targetCluster.ID)
			return interfaces.ExecutionResponse{}, errors.NewFlyteAdminErrorf(codes.Internal, "failed to create workflow in propeller %v", err)
		}
	}
	return interfaces.ExecutionResponse{
		Cluster: targetCluster.ID,
	}, nil
}

func (e K8sWorkflowExecutor) Abort(ctx context.Context, data interfaces.AbortData) error {
	target, err := e.executionCluster.GetTarget(ctx, &executioncluster.ExecutionTargetSpec{
		TargetID: data.Cluster,
	})
	if err != nil {
		return errors.NewFlyteAdminErrorf(codes.Internal, err.Error())
	}
	err = target.FlyteClient.FlyteworkflowV1alpha1().FlyteWorkflows(data.Namespace).Delete(ctx, data.ExecutionID.GetName(), v1.DeleteOptions{
		PropagationPolicy: &deletePropagationBackground,
	})
	// An IsNotFound error indicates the resource is already deleted.
	if err != nil && !k8_api_err.IsNotFound(err) {
		return errors.NewFlyteAdminErrorf(codes.Internal, "failed to terminate execution: %v with err %v", data.ExecutionID, err)
	}
	return nil
}

func NewK8sWorkflowExecutor(config runtimeInterfaces.Configuration, executionCluster execClusterInterfaces.ClusterInterface, workflowBuilder interfaces.FlyteWorkflowBuilder, client *storage.DataStore) *K8sWorkflowExecutor {

	return &K8sWorkflowExecutor{
		config:           config,
		executionCluster: executionCluster,
		workflowBuilder:  workflowBuilder,
		storageClient:    client,
	}
}
