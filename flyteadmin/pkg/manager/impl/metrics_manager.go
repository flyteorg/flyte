package impl

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"

	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	RequestLimit uint32 = 50

	nodeIdle         = "NODE_IDLE"
	nodeReset        = "NODE_RESET"
	nodeSetup        = "NODE_SETUP"
	nodeTeardown     = "NODE_TEARDOWN"
	nodeTransition   = "NODE_TRANSITION"
	taskRuntime      = "TASK_RUNTIME"
	taskSetup        = "TASK_SETUP"
	taskTeardown     = "TASK_TEARDOWN"
	workflowSetup    = "WORKFLOW_SETUP"
	workflowTeardown = "WORKFLOW_TEARDOWN"
)

var (
	emptyDuration *duration.Duration = &duration.Duration{
		Seconds: 0,
		Nanos:   0,
	}
	emptyTimestamp *timestamp.Timestamp = &timestamp.Timestamp{
		Seconds: 0,
		Nanos:   0,
	}
)

type metrics struct {
	Scope promutils.Scope
}

// MetricsManager handles computation of workflow, node, and task execution metrics.
type MetricsManager struct {
	workflowManager      interfaces.WorkflowInterface
	executionManager     interfaces.ExecutionInterface
	nodeExecutionManager interfaces.NodeExecutionInterface
	taskExecutionManager interfaces.TaskExecutionInterface
	metrics              metrics
}

// createOperationSpan returns a Span defined by the provided arguments.
func createOperationSpan(startTime, endTime *timestamp.Timestamp, operation string) *core.Span {
	return &core.Span{
		StartTime: startTime,
		EndTime:   endTime,
		Id: &core.Span_OperationId{
			OperationId: operation,
		},
	}
}

// getBranchNode searches the provided BranchNode definition for the Node identified by nodeID.
func getBranchNode(nodeID string, branchNode *core.BranchNode) *core.Node {
	if branchNode.IfElse.Case.ThenNode.Id == nodeID {
		return branchNode.IfElse.Case.ThenNode
	}

	for _, other := range branchNode.IfElse.Other {
		if other.ThenNode.Id == nodeID {
			return other.ThenNode
		}
	}

	if elseNode, ok := branchNode.IfElse.Default.(*core.IfElseBlock_ElseNode); ok {
		if elseNode.ElseNode.Id == nodeID {
			return elseNode.ElseNode
		}
	}

	return nil
}

// getLatestUpstreamNodeExecution returns the NodeExecution with the latest UpdatedAt timestamp that is an upstream
// dependency of the provided nodeID. This is useful for computing the duration between when a node is first available
// for scheduling and when it is actually scheduled.
func (m *MetricsManager) getLatestUpstreamNodeExecution(nodeID string, upstreamNodeIds map[string]*core.ConnectionSet_IdList,
	nodeExecutions map[string]*admin.NodeExecution) *admin.NodeExecution {

	var nodeExecution *admin.NodeExecution
	var latestUpstreamUpdatedAt = time.Unix(0, 0)
	if connectionSet, exists := upstreamNodeIds[nodeID]; exists {
		for _, upstreamNodeID := range connectionSet.Ids {
			upstreamNodeExecution, exists := nodeExecutions[upstreamNodeID]
			if !exists {
				continue
			}

			t := upstreamNodeExecution.Closure.UpdatedAt.AsTime()
			if t.After(latestUpstreamUpdatedAt) {
				nodeExecution = upstreamNodeExecution
				latestUpstreamUpdatedAt = t
			}
		}
	}

	return nodeExecution
}

// getNodeExecutions queries the nodeExecutionManager for NodeExecutions adhering to the specified request.
func (m *MetricsManager) getNodeExecutions(ctx context.Context, request admin.NodeExecutionListRequest) (map[string]*admin.NodeExecution, error) {
	nodeExecutions := make(map[string]*admin.NodeExecution)
	for {
		response, err := m.nodeExecutionManager.ListNodeExecutions(ctx, request)
		if err != nil {
			return nil, err
		}

		for _, nodeExecution := range response.NodeExecutions {
			nodeExecutions[nodeExecution.Metadata.SpecNodeId] = nodeExecution
		}

		if len(response.NodeExecutions) < int(request.Limit) {
			break
		}

		request.Token = response.Token
	}

	return nodeExecutions, nil
}

// getTaskExecutions queries the taskExecutionManager for TaskExecutions adhering to the specified request.
func (m *MetricsManager) getTaskExecutions(ctx context.Context, request admin.TaskExecutionListRequest) ([]*admin.TaskExecution, error) {
	taskExecutions := make([]*admin.TaskExecution, 0)
	for {
		response, err := m.taskExecutionManager.ListTaskExecutions(ctx, request)
		if err != nil {
			return nil, err
		}

		taskExecutions = append(taskExecutions, response.TaskExecutions...)

		if len(response.TaskExecutions) < int(request.Limit) {
			break
		}

		request.Token = response.Token
	}

	return taskExecutions, nil
}

// parseBranchNodeExecution partitions the BranchNode execution into a collection of Categorical and Reference Spans
// which are appended to the provided spans argument.
func (m *MetricsManager) parseBranchNodeExecution(ctx context.Context,
	nodeExecution *admin.NodeExecution, branchNode *core.BranchNode, spans *[]*core.Span, depth int) error {

	// retrieve node execution(s)
	nodeExecutions, err := m.getNodeExecutions(ctx, admin.NodeExecutionListRequest{
		WorkflowExecutionId: nodeExecution.Id.ExecutionId,
		Limit:               RequestLimit,
		UniqueParentId:      nodeExecution.Id.NodeId,
	})
	if err != nil {
		return err
	}

	// check if the node started
	if len(nodeExecutions) == 0 {
		*spans = append(*spans, createOperationSpan(nodeExecution.Closure.CreatedAt, nodeExecution.Closure.UpdatedAt, nodeSetup))
	} else {
		// parse branchNode
		if len(nodeExecutions) != 1 {
			return fmt.Errorf("invalid branch node execution: expected 1 but found %d node execution(s)", len(nodeExecutions))
		}

		var branchNodeExecution *admin.NodeExecution
		for _, e := range nodeExecutions {
			branchNodeExecution = e
		}

		node := getBranchNode(branchNodeExecution.Metadata.SpecNodeId, branchNode)
		if node == nil {
			return fmt.Errorf("failed to identify branch node final node definition for nodeID '%s' and branchNode '%+v'",
				branchNodeExecution.Metadata.SpecNodeId, branchNode)
		}

		// frontend overhead
		*spans = append(*spans, createOperationSpan(nodeExecution.Closure.CreatedAt, branchNodeExecution.Closure.CreatedAt, nodeSetup))

		// node execution
		nodeExecutionSpan, err := m.parseNodeExecution(ctx, branchNodeExecution, node, depth)
		if err != nil {
			return err
		}

		*spans = append(*spans, nodeExecutionSpan)

		// backened overhead
		if !nodeExecution.Closure.UpdatedAt.AsTime().Before(branchNodeExecution.Closure.UpdatedAt.AsTime()) {
			*spans = append(*spans, createOperationSpan(branchNodeExecution.Closure.UpdatedAt,
				nodeExecution.Closure.UpdatedAt, nodeTeardown))
		}
	}

	return nil
}

// parseDynamicNodeExecution partitions the DynamicNode execution into a collection of Categorical and Reference Spans
// which are appended to the provided spans argument.
func (m *MetricsManager) parseDynamicNodeExecution(ctx context.Context, nodeExecution *admin.NodeExecution, spans *[]*core.Span, depth int) error {
	taskExecutions, err := m.getTaskExecutions(ctx, admin.TaskExecutionListRequest{
		NodeExecutionId: nodeExecution.Id,
		Limit:           RequestLimit,
	})
	if err != nil {
		return err
	}

	// if no task executions then everything is execution overhead
	if len(taskExecutions) == 0 {
		*spans = append(*spans, createOperationSpan(nodeExecution.Closure.CreatedAt, nodeExecution.Closure.UpdatedAt, nodeSetup))
	} else {
		// frontend overhead
		*spans = append(*spans, createOperationSpan(nodeExecution.Closure.CreatedAt, taskExecutions[0].Closure.CreatedAt, nodeSetup))

		// task execution(s)
		parseTaskExecutions(taskExecutions, spans, depth)

		nodeExecutions, err := m.getNodeExecutions(ctx, admin.NodeExecutionListRequest{
			WorkflowExecutionId: nodeExecution.Id.ExecutionId,
			Limit:               RequestLimit,
			UniqueParentId:      nodeExecution.Id.NodeId,
		})
		if err != nil {
			return err
		}

		lastTask := taskExecutions[len(taskExecutions)-1]
		if len(nodeExecutions) == 0 {
			if !nodeExecution.Closure.UpdatedAt.AsTime().Before(lastTask.Closure.UpdatedAt.AsTime()) {
				*spans = append(*spans, createOperationSpan(lastTask.Closure.UpdatedAt, nodeExecution.Closure.UpdatedAt, nodeReset))
			}
		} else {
			// between task execution(s) and node execution(s) overhead
			startNode := nodeExecutions[v1alpha1.StartNodeID]
			*spans = append(*spans, createOperationSpan(taskExecutions[len(taskExecutions)-1].Closure.UpdatedAt,
				startNode.Closure.UpdatedAt, nodeReset))

			// node execution(s)
			getDataRequest := admin.NodeExecutionGetDataRequest{Id: nodeExecution.Id}
			nodeExecutionData, err := m.nodeExecutionManager.GetNodeExecutionData(ctx, getDataRequest)
			if err != nil {
				return err
			}

			if err := m.parseNodeExecutions(ctx, nodeExecutions, nodeExecutionData.DynamicWorkflow.CompiledWorkflow, spans, depth); err != nil {
				return err
			}

			// backened overhead
			latestUpstreamNode := m.getLatestUpstreamNodeExecution(v1alpha1.EndNodeID,
				nodeExecutionData.DynamicWorkflow.CompiledWorkflow.Primary.Connections.Upstream, nodeExecutions)
			if latestUpstreamNode != nil && !nodeExecution.Closure.UpdatedAt.AsTime().Before(latestUpstreamNode.Closure.UpdatedAt.AsTime()) {
				*spans = append(*spans, createOperationSpan(latestUpstreamNode.Closure.UpdatedAt, nodeExecution.Closure.UpdatedAt, nodeTeardown))
			}
		}
	}

	return nil
}

// parseExecution partitions the workflow execution into a collection of Categorical and Reference Spans which are
// returned as a hierarchical breakdown of the workflow execution.
func (m *MetricsManager) parseExecution(ctx context.Context, execution *admin.Execution, depth int) (*core.Span, error) {
	spans := make([]*core.Span, 0)
	if depth != 0 {
		// retrieve workflow and node executions
		workflowRequest := admin.ObjectGetRequest{Id: execution.Closure.WorkflowId}
		workflow, err := m.workflowManager.GetWorkflow(ctx, workflowRequest)
		if err != nil {
			return nil, err
		}

		nodeExecutions, err := m.getNodeExecutions(ctx, admin.NodeExecutionListRequest{
			WorkflowExecutionId: execution.Id,
			Limit:               RequestLimit,
		})
		if err != nil {
			return nil, err
		}

		// check if workflow has started
		startNode := nodeExecutions[v1alpha1.StartNodeID]
		if startNode.Closure.UpdatedAt == nil || reflect.DeepEqual(startNode.Closure.UpdatedAt, emptyTimestamp) {
			spans = append(spans, createOperationSpan(execution.Closure.CreatedAt, execution.Closure.UpdatedAt, workflowSetup))
		} else {
			// compute frontend overhead
			spans = append(spans, createOperationSpan(execution.Closure.CreatedAt, startNode.Closure.UpdatedAt, workflowSetup))

			// iterate over nodes and compute overhead
			if err := m.parseNodeExecutions(ctx, nodeExecutions, workflow.Closure.CompiledWorkflow, &spans, depth-1); err != nil {
				return nil, err
			}

			// compute backend overhead
			latestUpstreamNode := m.getLatestUpstreamNodeExecution(v1alpha1.EndNodeID,
				workflow.Closure.CompiledWorkflow.Primary.Connections.Upstream, nodeExecutions)
			if latestUpstreamNode != nil && !execution.Closure.UpdatedAt.AsTime().Before(latestUpstreamNode.Closure.UpdatedAt.AsTime()) {
				spans = append(spans, createOperationSpan(latestUpstreamNode.Closure.UpdatedAt,
					execution.Closure.UpdatedAt, workflowTeardown))
			}
		}
	}

	return &core.Span{
		StartTime: execution.Closure.CreatedAt,
		EndTime:   execution.Closure.UpdatedAt,
		Id: &core.Span_WorkflowId{
			WorkflowId: execution.Id,
		},
		Spans: spans,
	}, nil
}

// parseGateNodeExecution partitions the GateNode execution into a collection of Categorical and Reference Spans
// which are appended to the provided spans argument.
func (m *MetricsManager) parseGateNodeExecution(_ context.Context, nodeExecution *admin.NodeExecution, spans *[]*core.Span) {
	// check if node has started yet
	if nodeExecution.Closure.StartedAt == nil || reflect.DeepEqual(nodeExecution.Closure.StartedAt, emptyTimestamp) {
		*spans = append(*spans, createOperationSpan(nodeExecution.Closure.CreatedAt, nodeExecution.Closure.UpdatedAt, nodeSetup))
	} else {
		// frontend overhead
		*spans = append(*spans, createOperationSpan(nodeExecution.Closure.CreatedAt, nodeExecution.Closure.StartedAt, nodeSetup))

		// check if plugin has completed yet
		if nodeExecution.Closure.Duration == nil || reflect.DeepEqual(nodeExecution.Closure.Duration, emptyDuration) {
			*spans = append(*spans, createOperationSpan(nodeExecution.Closure.StartedAt,
				nodeExecution.Closure.UpdatedAt, nodeIdle))
		} else {
			// idle time
			nodeEndTime := timestamppb.New(nodeExecution.Closure.StartedAt.AsTime().Add(nodeExecution.Closure.Duration.AsDuration()))
			*spans = append(*spans, createOperationSpan(nodeExecution.Closure.StartedAt, nodeEndTime, nodeIdle))

			// backend overhead
			*spans = append(*spans, createOperationSpan(nodeEndTime, nodeExecution.Closure.UpdatedAt, nodeTeardown))
		}
	}
}

// parseLaunchPlanNodeExecution partitions the LaunchPlanNode execution into a collection of Categorical and Reference
// Spans which are appended to the provided spans argument.
func (m *MetricsManager) parseLaunchPlanNodeExecution(ctx context.Context, nodeExecution *admin.NodeExecution, spans *[]*core.Span, depth int) error {
	// check if workflow started yet
	workflowNode := nodeExecution.Closure.GetWorkflowNodeMetadata()
	if workflowNode == nil {
		*spans = append(*spans, createOperationSpan(nodeExecution.Closure.CreatedAt, nodeExecution.Closure.UpdatedAt, nodeSetup))
	} else {
		// retrieve execution
		executionRequest := admin.WorkflowExecutionGetRequest{Id: workflowNode.ExecutionId}
		execution, err := m.executionManager.GetExecution(ctx, executionRequest)
		if err != nil {
			return err
		}

		// frontend overhead
		*spans = append(*spans, createOperationSpan(nodeExecution.Closure.CreatedAt, execution.Closure.CreatedAt, nodeSetup))

		// execution
		span, err := m.parseExecution(ctx, execution, depth)
		if err != nil {
			return err
		}

		*spans = append(*spans, span)

		// backend overhead
		if !nodeExecution.Closure.UpdatedAt.AsTime().Before(execution.Closure.UpdatedAt.AsTime()) {
			*spans = append(*spans, createOperationSpan(execution.Closure.UpdatedAt, nodeExecution.Closure.UpdatedAt, nodeTeardown))
		}
	}

	return nil
}

// parseNodeExecution partitions the node execution into a collection of Categorical and Reference Spans which are
// returned as a hierarchical breakdown of the node execution.
func (m *MetricsManager) parseNodeExecution(ctx context.Context, nodeExecution *admin.NodeExecution, node *core.Node, depth int) (*core.Span, error) {
	spans := make([]*core.Span, 0)
	if depth != 0 {

		// parse node
		var err error
		switch target := node.Target.(type) {
		case *core.Node_BranchNode:
			// handle branch node
			err = m.parseBranchNodeExecution(ctx, nodeExecution, target.BranchNode, &spans, depth-1)
		case *core.Node_GateNode:
			// handle gate node
			m.parseGateNodeExecution(ctx, nodeExecution, &spans)
		case *core.Node_TaskNode:
			if nodeExecution.Metadata.IsParentNode {
				// handle dynamic node
				err = m.parseDynamicNodeExecution(ctx, nodeExecution, &spans, depth-1)
			} else {
				// handle task node
				err = m.parseTaskNodeExecution(ctx, nodeExecution, &spans, depth-1)
			}
		case *core.Node_WorkflowNode:
			switch workflow := target.WorkflowNode.Reference.(type) {
			case *core.WorkflowNode_LaunchplanRef:
				// handle launch plan
				err = m.parseLaunchPlanNodeExecution(ctx, nodeExecution, &spans, depth-1)
			case *core.WorkflowNode_SubWorkflowRef:
				// handle subworkflow
				err = m.parseSubworkflowNodeExecution(ctx, nodeExecution, workflow.SubWorkflowRef, &spans, depth-1)
			default:
				err = fmt.Errorf("failed to identify workflow node type for node: %+v", target)
			}
		default:
			err = fmt.Errorf("failed to identify node type for node: %+v", target)
		}

		if err != nil {
			return nil, err
		}
	}

	return &core.Span{
		StartTime: nodeExecution.Closure.CreatedAt,
		EndTime:   nodeExecution.Closure.UpdatedAt,
		Id: &core.Span_NodeId{
			NodeId: nodeExecution.Id,
		},
		Spans: spans,
	}, nil
}

// parseNodeExecutions partitions the node executions into a collection of Categorical and Reference Spans which are
// appended to the provided spans argument.
func (m *MetricsManager) parseNodeExecutions(ctx context.Context, nodeExecutions map[string]*admin.NodeExecution,
	compiledWorkflowClosure *core.CompiledWorkflowClosure, spans *[]*core.Span, depth int) error {

	// sort node executions
	sortedNodeExecutions := make([]*admin.NodeExecution, 0, len(nodeExecutions))
	for _, nodeExecution := range nodeExecutions {
		sortedNodeExecutions = append(sortedNodeExecutions, nodeExecution)
	}
	sort.Slice(sortedNodeExecutions, func(i, j int) bool {
		x := sortedNodeExecutions[i].Closure.CreatedAt.AsTime()
		y := sortedNodeExecutions[j].Closure.CreatedAt.AsTime()
		return x.Before(y)
	})

	// iterate over sorted node executions
	for _, nodeExecution := range sortedNodeExecutions {
		specNodeID := nodeExecution.Metadata.SpecNodeId
		if specNodeID == v1alpha1.StartNodeID || specNodeID == v1alpha1.EndNodeID {
			continue
		}

		// get node definition from workflow
		var node *core.Node
		for _, n := range compiledWorkflowClosure.Primary.Template.Nodes {
			if n.Id == specNodeID {
				node = n
			}
		}

		if node == nil {
			return fmt.Errorf("failed to discover workflow node '%s' in workflow '%+v'",
				specNodeID, compiledWorkflowClosure.Primary.Template.Id)
		}

		// parse node execution
		nodeExecutionSpan, err := m.parseNodeExecution(ctx, nodeExecution, node, depth)
		if err != nil {
			return err
		}

		// prepend nodeExecution spans with node transition time
		latestUpstreamNode := m.getLatestUpstreamNodeExecution(specNodeID,
			compiledWorkflowClosure.Primary.Connections.Upstream, nodeExecutions)
		if latestUpstreamNode != nil {
			nodeExecutionSpan.Spans = append([]*core.Span{createOperationSpan(latestUpstreamNode.Closure.UpdatedAt,
				nodeExecution.Closure.CreatedAt, nodeTransition)}, nodeExecutionSpan.Spans...)
		}

		*spans = append(*spans, nodeExecutionSpan)
	}

	return nil
}

// parseSubworkflowNodeExecutions partitions the SubworkflowNode execution into a collection of Categorical and
// Reference Spans which are appended to the provided spans argument.
func (m *MetricsManager) parseSubworkflowNodeExecution(ctx context.Context,
	nodeExecution *admin.NodeExecution, identifier *core.Identifier, spans *[]*core.Span, depth int) error {

	// retrieve node execution(s)
	nodeExecutions, err := m.getNodeExecutions(ctx, admin.NodeExecutionListRequest{
		WorkflowExecutionId: nodeExecution.Id.ExecutionId,
		Limit:               RequestLimit,
		UniqueParentId:      nodeExecution.Id.NodeId,
	})
	if err != nil {
		return err
	}

	// check if the subworkflow started
	if len(nodeExecutions) == 0 {
		*spans = append(*spans, createOperationSpan(nodeExecution.Closure.CreatedAt, nodeExecution.Closure.UpdatedAt, nodeSetup))
	} else {
		// frontend overhead
		startNode := nodeExecutions[v1alpha1.StartNodeID]
		*spans = append(*spans, createOperationSpan(nodeExecution.Closure.CreatedAt, startNode.Closure.UpdatedAt, nodeSetup))

		// retrieve workflow
		workflowRequest := admin.ObjectGetRequest{Id: identifier}
		workflow, err := m.workflowManager.GetWorkflow(ctx, workflowRequest)
		if err != nil {
			return err
		}

		// node execution(s)
		if err := m.parseNodeExecutions(ctx, nodeExecutions, workflow.Closure.CompiledWorkflow, spans, depth); err != nil {
			return err
		}

		// backened overhead
		latestUpstreamNode := m.getLatestUpstreamNodeExecution(v1alpha1.EndNodeID,
			workflow.Closure.CompiledWorkflow.Primary.Connections.Upstream, nodeExecutions)
		if latestUpstreamNode != nil && !nodeExecution.Closure.UpdatedAt.AsTime().Before(latestUpstreamNode.Closure.UpdatedAt.AsTime()) {
			*spans = append(*spans, createOperationSpan(latestUpstreamNode.Closure.UpdatedAt, nodeExecution.Closure.UpdatedAt, nodeTeardown))
		}
	}

	return nil
}

// parseTaskExecution partitions the task execution into a collection of Categorical and Reference Spans which are
// returned as a hierarchical breakdown of the task execution.
func parseTaskExecution(taskExecution *admin.TaskExecution) *core.Span {
	spans := make([]*core.Span, 0)

	// check if plugin has started yet
	if taskExecution.Closure.StartedAt == nil || reflect.DeepEqual(taskExecution.Closure.StartedAt, emptyTimestamp) {
		spans = append(spans, createOperationSpan(taskExecution.Closure.CreatedAt, taskExecution.Closure.UpdatedAt, taskSetup))
	} else {
		// frontend overhead
		spans = append(spans, createOperationSpan(taskExecution.Closure.CreatedAt, taskExecution.Closure.StartedAt, taskSetup))

		// check if plugin has completed yet
		if taskExecution.Closure.Duration == nil || reflect.DeepEqual(taskExecution.Closure.Duration, emptyDuration) {
			spans = append(spans, createOperationSpan(taskExecution.Closure.StartedAt, taskExecution.Closure.UpdatedAt, taskRuntime))
		} else {
			// plugin execution
			taskEndTime := timestamppb.New(taskExecution.Closure.StartedAt.AsTime().Add(taskExecution.Closure.Duration.AsDuration()))
			spans = append(spans, createOperationSpan(taskExecution.Closure.StartedAt, taskEndTime, taskRuntime))

			// backend overhead
			if !taskExecution.Closure.UpdatedAt.AsTime().Before(taskEndTime.AsTime()) {
				spans = append(spans, createOperationSpan(taskEndTime, taskExecution.Closure.UpdatedAt, taskTeardown))
			}
		}
	}

	return &core.Span{
		StartTime: taskExecution.Closure.CreatedAt,
		EndTime:   taskExecution.Closure.UpdatedAt,
		Id: &core.Span_TaskId{
			TaskId: taskExecution.Id,
		},
		Spans: spans,
	}
}

// parseTaskExecutions partitions the task executions into a collection of Categorical and Reference Spans which are
// appended to the provided spans argument.
func parseTaskExecutions(taskExecutions []*admin.TaskExecution, spans *[]*core.Span, depth int) {
	// sort task executions
	sort.Slice(taskExecutions, func(i, j int) bool {
		x := taskExecutions[i].Closure.CreatedAt.AsTime()
		y := taskExecutions[j].Closure.CreatedAt.AsTime()
		return x.Before(y)
	})

	// iterate over task executions
	for index, taskExecution := range taskExecutions {
		if index > 0 {
			*spans = append(*spans, createOperationSpan(taskExecutions[index-1].Closure.UpdatedAt, taskExecution.Closure.CreatedAt, nodeReset))
		}

		if depth != 0 {
			*spans = append(*spans, parseTaskExecution(taskExecution))
		}
	}
}

// parseTaskNodeExecutions partitions the TaskNode execution into a collection of Categorical and Reference Spans which
// are appended to the provided spans argument.
func (m *MetricsManager) parseTaskNodeExecution(ctx context.Context, nodeExecution *admin.NodeExecution, spans *[]*core.Span, depth int) error {
	// retrieve task executions
	taskExecutions, err := m.getTaskExecutions(ctx, admin.TaskExecutionListRequest{
		NodeExecutionId: nodeExecution.Id,
		Limit:           RequestLimit,
	})
	if err != nil {
		return err
	}

	// if no task executions then everything is execution overhead
	if len(taskExecutions) == 0 {
		*spans = append(*spans, createOperationSpan(nodeExecution.Closure.CreatedAt, nodeExecution.Closure.UpdatedAt, nodeSetup))
	} else {
		// frontend overhead
		*spans = append(*spans, createOperationSpan(nodeExecution.Closure.CreatedAt, taskExecutions[0].Closure.CreatedAt, nodeSetup))

		// parse task executions
		parseTaskExecutions(taskExecutions, spans, depth)

		// backend overhead
		lastTask := taskExecutions[len(taskExecutions)-1]
		if !nodeExecution.Closure.UpdatedAt.AsTime().Before(lastTask.Closure.UpdatedAt.AsTime()) {
			*spans = append(*spans, createOperationSpan(taskExecutions[len(taskExecutions)-1].Closure.UpdatedAt,
				nodeExecution.Closure.UpdatedAt, nodeTeardown))
		}
	}

	return nil
}

// GetExecutionMetrics returns a Span hierarchically breaking down the workflow execution into a collection of
// Categorical and Reference Spans.
func (m *MetricsManager) GetExecutionMetrics(ctx context.Context,
	request admin.WorkflowExecutionGetMetricsRequest) (*admin.WorkflowExecutionGetMetricsResponse, error) {

	// retrieve workflow execution
	executionRequest := admin.WorkflowExecutionGetRequest{Id: request.Id}
	execution, err := m.executionManager.GetExecution(ctx, executionRequest)
	if err != nil {
		return nil, err
	}

	span, err := m.parseExecution(ctx, execution, int(request.Depth))
	if err != nil {
		return nil, err
	}

	return &admin.WorkflowExecutionGetMetricsResponse{Span: span}, nil
}

// NewMetricsManager returns a new MetricsManager constructed with the provided arguments.
func NewMetricsManager(
	workflowManager interfaces.WorkflowInterface,
	executionManager interfaces.ExecutionInterface,
	nodeExecutionManager interfaces.NodeExecutionInterface,
	taskExecutionManager interfaces.TaskExecutionInterface,
	scope promutils.Scope) interfaces.MetricsInterface {
	metrics := metrics{
		Scope: scope,
	}

	return &MetricsManager{
		workflowManager:      workflowManager,
		executionManager:     executionManager,
		nodeExecutionManager: nodeExecutionManager,
		taskExecutionManager: taskExecutionManager,
		metrics:              metrics,
	}
}
