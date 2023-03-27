package impl

import (
	"context"
	"reflect"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/manager/mocks"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/stretchr/testify/assert"
)

var (
	baseDuration = &duration.Duration{
		Seconds: 400,
		Nanos:   0,
	}
	baseTimestamp = &timestamp.Timestamp{
		Seconds: 643852800,
		Nanos:   0,
	}
)

func addTimestamp(ts *timestamp.Timestamp, seconds int64) *timestamp.Timestamp {
	return &timestamp.Timestamp{
		Seconds: ts.Seconds + seconds,
		Nanos:   ts.Nanos,
	}
}

func getMockExecutionManager(execution *admin.Execution) interfaces.ExecutionInterface {
	mockExecutionManager := mocks.MockExecutionManager{}
	mockExecutionManager.SetGetCallback(
		func(ctx context.Context, request admin.WorkflowExecutionGetRequest) (*admin.Execution, error) {
			return execution, nil
		})

	return &mockExecutionManager
}

func getMockNodeExecutionManager(nodeExecutions []*admin.NodeExecution,
	dynamicWorkflow *admin.DynamicWorkflowNodeMetadata) interfaces.NodeExecutionInterface {

	mockNodeExecutionManager := mocks.MockNodeExecutionManager{}
	mockNodeExecutionManager.SetListNodeExecutionsFunc(
		func(ctx context.Context, request admin.NodeExecutionListRequest) (*admin.NodeExecutionList, error) {
			return &admin.NodeExecutionList{
				NodeExecutions: nodeExecutions,
			}, nil
		})
	mockNodeExecutionManager.SetGetNodeExecutionDataFunc(
		func(ctx context.Context, request admin.NodeExecutionGetDataRequest) (*admin.NodeExecutionGetDataResponse, error) {
			return &admin.NodeExecutionGetDataResponse{
				DynamicWorkflow: dynamicWorkflow,
			}, nil
		})

	return &mockNodeExecutionManager
}

func getMockTaskExecutionManager(taskExecutions []*admin.TaskExecution) interfaces.TaskExecutionInterface {
	mockTaskExecutionManager := mocks.MockTaskExecutionManager{}
	mockTaskExecutionManager.SetListTaskExecutionsCallback(
		func(ctx context.Context, request admin.TaskExecutionListRequest) (*admin.TaskExecutionList, error) {
			return &admin.TaskExecutionList{
				TaskExecutions: taskExecutions,
			}, nil
		})

	return &mockTaskExecutionManager
}

func getMockWorkflowManager(workflow *admin.Workflow) interfaces.WorkflowInterface {
	mockWorkflowManager := mocks.MockWorkflowManager{}
	mockWorkflowManager.SetGetCallback(
		func(ctx context.Context, request admin.ObjectGetRequest) (*admin.Workflow, error) {
			return workflow, nil
		})

	return &mockWorkflowManager
}

func parseSpans(spans []*core.Span) (map[string][]int64, int) {
	operationDurations := make(map[string][]int64)
	referenceCount := 0
	for _, span := range spans {
		switch id := span.Id.(type) {
		case *core.Span_OperationId:
			operationID := id.OperationId
			duration := span.EndTime.Seconds - span.StartTime.Seconds
			if array, exists := operationDurations[operationID]; exists {
				operationDurations[operationID] = append(array, duration)
			} else {
				operationDurations[operationID] = []int64{duration}
			}
		default:
			referenceCount++
		}
	}

	return operationDurations, referenceCount
}

func TestParseBranchNodeExecution(t *testing.T) {
	tests := []struct {
		name               string
		nodeExecution      *admin.NodeExecution
		nodeExecutions     []*admin.NodeExecution
		operationDurations map[string][]int64
		referenceCount     int
	}{
		{
			"NotStarted",
			&admin.NodeExecution{
				Id: &core.NodeExecutionIdentifier{
					ExecutionId: &core.WorkflowExecutionIdentifier{},
				},
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: emptyTimestamp,
					Duration:  emptyDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 5),
				},
			},
			nil,
			map[string][]int64{
				nodeSetup: []int64{5},
			},
			0,
		},
		{
			"Running",
			&admin.NodeExecution{
				Id: &core.NodeExecutionIdentifier{
					ExecutionId: &core.WorkflowExecutionIdentifier{},
				},
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: emptyTimestamp,
					Duration:  emptyDuration,
					UpdatedAt: baseTimestamp,
				},
			},
			[]*admin.NodeExecution{
				&admin.NodeExecution{
					Metadata: &admin.NodeExecutionMetaData{
						SpecNodeId: "foo",
					},
					Closure: &admin.NodeExecutionClosure{
						CreatedAt: addTimestamp(baseTimestamp, 10),
						StartedAt: addTimestamp(baseTimestamp, 15),
						Duration:  baseDuration,
						UpdatedAt: addTimestamp(baseTimestamp, 430),
					},
				},
			},
			map[string][]int64{
				nodeSetup: []int64{10},
			},
			1,
		},
		{
			"Completed",
			&admin.NodeExecution{
				Id: &core.NodeExecutionIdentifier{
					ExecutionId: &core.WorkflowExecutionIdentifier{},
				},
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: emptyTimestamp,
					Duration:  emptyDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 450),
				},
			},
			[]*admin.NodeExecution{
				&admin.NodeExecution{
					Metadata: &admin.NodeExecutionMetaData{
						SpecNodeId: "foo",
					},
					Closure: &admin.NodeExecutionClosure{
						CreatedAt: addTimestamp(baseTimestamp, 10),
						StartedAt: addTimestamp(baseTimestamp, 15),
						Duration:  baseDuration,
						UpdatedAt: addTimestamp(baseTimestamp, 430),
					},
				},
			},
			map[string][]int64{
				nodeSetup:    []int64{10},
				nodeTeardown: []int64{20},
			},
			1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// initialize mocks
			mockNodeExecutionManager := getMockNodeExecutionManager(test.nodeExecutions, nil)
			mockTaskExecutionManager := getMockTaskExecutionManager([]*admin.TaskExecution{})
			metricsManager := MetricsManager{
				nodeExecutionManager: mockNodeExecutionManager,
				taskExecutionManager: mockTaskExecutionManager,
			}

			// parse node execution
			branchNode := &core.BranchNode{
				IfElse: &core.IfElseBlock{
					Case: &core.IfBlock{
						ThenNode: &core.Node{
							Id: "bar",
						},
					},
					Other: []*core.IfBlock{
						&core.IfBlock{
							ThenNode: &core.Node{
								Id: "baz",
							},
						},
					},
					Default: &core.IfElseBlock_ElseNode{
						ElseNode: &core.Node{
							Id:     "foo",
							Target: &core.Node_TaskNode{},
						},
					},
				},
			}

			spans := make([]*core.Span, 0)
			err := metricsManager.parseBranchNodeExecution(context.TODO(), test.nodeExecution, branchNode, &spans, -1)
			assert.Nil(t, err)

			// validate spans
			operationDurations, referenceCount := parseSpans(spans)
			assert.True(t, reflect.DeepEqual(test.operationDurations, operationDurations))
			assert.Equal(t, test.referenceCount, referenceCount)
		})
	}
}

func TestParseDynamicNodeExecution(t *testing.T) {
	tests := []struct {
		name               string
		nodeExecution      *admin.NodeExecution
		taskExecutions     []*admin.TaskExecution
		nodeExecutions     []*admin.NodeExecution
		operationDurations map[string][]int64
		referenceCount     int
	}{
		{
			"NotStarted",
			&admin.NodeExecution{
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: emptyTimestamp,
					Duration:  emptyDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 5),
				},
			},
			nil,
			nil,
			map[string][]int64{
				nodeSetup: []int64{5},
			},
			0,
		},
		{
			"TaskRunning",
			&admin.NodeExecution{
				Id: &core.NodeExecutionIdentifier{
					ExecutionId: &core.WorkflowExecutionIdentifier{},
				},
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: emptyTimestamp,
					Duration:  emptyDuration,
					UpdatedAt: baseTimestamp,
				},
			},
			[]*admin.TaskExecution{
				&admin.TaskExecution{
					Closure: &admin.TaskExecutionClosure{
						CreatedAt: addTimestamp(baseTimestamp, 10),
						StartedAt: addTimestamp(baseTimestamp, 15),
						Duration:  baseDuration,
						UpdatedAt: addTimestamp(baseTimestamp, 420),
					},
				},
			},
			nil,
			map[string][]int64{
				nodeSetup: []int64{10},
			},
			1,
		},
		{
			"NodesRunning",
			&admin.NodeExecution{
				Id: &core.NodeExecutionIdentifier{
					ExecutionId: &core.WorkflowExecutionIdentifier{},
				},
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: emptyTimestamp,
					Duration:  emptyDuration,
					UpdatedAt: baseTimestamp,
				},
			},
			[]*admin.TaskExecution{
				&admin.TaskExecution{
					Closure: &admin.TaskExecutionClosure{
						CreatedAt: addTimestamp(baseTimestamp, 10),
						StartedAt: addTimestamp(baseTimestamp, 15),
						Duration:  baseDuration,
						UpdatedAt: addTimestamp(baseTimestamp, 420),
					},
				},
			},
			[]*admin.NodeExecution{
				&admin.NodeExecution{
					Metadata: &admin.NodeExecutionMetaData{
						SpecNodeId: "start-node",
					},
					Closure: &admin.NodeExecutionClosure{
						CreatedAt: addTimestamp(baseTimestamp, 435),
						StartedAt: emptyTimestamp,
						Duration:  emptyDuration,
						UpdatedAt: addTimestamp(baseTimestamp, 435),
					},
				},
				&admin.NodeExecution{
					Metadata: &admin.NodeExecutionMetaData{
						SpecNodeId: "foo",
					},
					Closure: &admin.NodeExecutionClosure{
						CreatedAt: addTimestamp(baseTimestamp, 445),
						StartedAt: addTimestamp(baseTimestamp, 460),
						Duration:  baseDuration,
						UpdatedAt: addTimestamp(baseTimestamp, 880),
					},
				},
			},
			map[string][]int64{
				nodeSetup: []int64{10},
				nodeReset: []int64{15},
			},
			2,
		},
		{
			"Completed",
			&admin.NodeExecution{
				Id: &core.NodeExecutionIdentifier{
					ExecutionId: &core.WorkflowExecutionIdentifier{},
				},
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: emptyTimestamp,
					Duration:  emptyDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 900),
				},
			},
			[]*admin.TaskExecution{
				&admin.TaskExecution{
					Closure: &admin.TaskExecutionClosure{
						CreatedAt: addTimestamp(baseTimestamp, 10),
						StartedAt: addTimestamp(baseTimestamp, 15),
						Duration:  baseDuration,
						UpdatedAt: addTimestamp(baseTimestamp, 420),
					},
				},
			},
			[]*admin.NodeExecution{
				&admin.NodeExecution{
					Metadata: &admin.NodeExecutionMetaData{
						SpecNodeId: "start-node",
					},
					Closure: &admin.NodeExecutionClosure{
						CreatedAt: addTimestamp(baseTimestamp, 435),
						StartedAt: emptyTimestamp,
						Duration:  emptyDuration,
						UpdatedAt: addTimestamp(baseTimestamp, 435),
					},
				},
				&admin.NodeExecution{
					Metadata: &admin.NodeExecutionMetaData{
						SpecNodeId: "foo",
					},
					Closure: &admin.NodeExecutionClosure{
						CreatedAt: addTimestamp(baseTimestamp, 445),
						StartedAt: addTimestamp(baseTimestamp, 460),
						Duration:  baseDuration,
						UpdatedAt: addTimestamp(baseTimestamp, 880),
					},
				},
			},
			map[string][]int64{
				nodeSetup:    []int64{10},
				nodeReset:    []int64{15},
				nodeTeardown: []int64{20},
			},
			2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// initialize mocks
			mockNodeExecutionManager := getMockNodeExecutionManager(
				test.nodeExecutions,
				&admin.DynamicWorkflowNodeMetadata{
					CompiledWorkflow: &core.CompiledWorkflowClosure{
						Primary: &core.CompiledWorkflow{
							Connections: &core.ConnectionSet{
								Upstream: map[string]*core.ConnectionSet_IdList{
									"foo": &core.ConnectionSet_IdList{
										Ids: []string{"start-node"},
									},
									"end-node": &core.ConnectionSet_IdList{
										Ids: []string{"foo"},
									},
								},
							},
							Template: &core.WorkflowTemplate{
								Nodes: []*core.Node{
									&core.Node{
										Id:     "foo",
										Target: &core.Node_TaskNode{},
									},
								},
							},
						},
					},
				})
			mockTaskExecutionManager := getMockTaskExecutionManager(test.taskExecutions)
			metricsManager := MetricsManager{
				nodeExecutionManager: mockNodeExecutionManager,
				taskExecutionManager: mockTaskExecutionManager,
			}

			// parse node execution
			spans := make([]*core.Span, 0)
			err := metricsManager.parseDynamicNodeExecution(context.TODO(), test.nodeExecution, &spans, -1)
			assert.Nil(t, err)

			// validate spans
			operationDurations, referenceCount := parseSpans(spans)
			assert.True(t, reflect.DeepEqual(test.operationDurations, operationDurations))
			assert.Equal(t, test.referenceCount, referenceCount)
		})
	}
}

func TestParseGateNodeExecution(t *testing.T) {
	tests := []struct {
		name               string
		nodeExecution      *admin.NodeExecution
		operationDurations map[string][]int64
	}{
		{
			"NotStarted",
			&admin.NodeExecution{
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: emptyTimestamp,
					Duration:  emptyDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 5),
				},
			},
			map[string][]int64{
				nodeSetup: []int64{5},
			},
		},
		{
			"Running",
			&admin.NodeExecution{
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: addTimestamp(baseTimestamp, 10),
					Duration:  emptyDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 15),
				},
			},
			map[string][]int64{
				nodeSetup: []int64{10},
				nodeIdle:  []int64{5},
			},
		},
		{
			"Completed",
			&admin.NodeExecution{
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: addTimestamp(baseTimestamp, 10),
					Duration:  baseDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 425),
				},
			},
			map[string][]int64{
				nodeSetup:    []int64{10},
				nodeIdle:     []int64{400},
				nodeTeardown: []int64{15},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// initialize mocks
			metricsManager := MetricsManager{}

			// parse node execution
			spans := make([]*core.Span, 0)
			metricsManager.parseGateNodeExecution(context.TODO(), test.nodeExecution, &spans)

			// validate spans
			operationDurations, _ := parseSpans(spans)
			assert.True(t, reflect.DeepEqual(test.operationDurations, operationDurations))
		})
	}
}

func TestParseLaunchPlanNodeExecution(t *testing.T) {
	tests := []struct {
		name               string
		nodeExecution      *admin.NodeExecution
		execution          *admin.Execution
		operationDurations map[string][]int64
		referenceCount     int
	}{
		{
			"NotStarted",
			&admin.NodeExecution{
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: emptyTimestamp,
					Duration:  emptyDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 5),
				},
			},
			nil,
			map[string][]int64{
				nodeSetup: []int64{5},
			},
			0,
		},
		{
			"Running",
			&admin.NodeExecution{
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: emptyTimestamp,
					Duration:  emptyDuration,
					UpdatedAt: baseTimestamp,
					TargetMetadata: &admin.NodeExecutionClosure_WorkflowNodeMetadata{
						WorkflowNodeMetadata: &admin.WorkflowNodeMetadata{
							ExecutionId: &core.WorkflowExecutionIdentifier{},
						},
					},
				},
			},
			&admin.Execution{
				Closure: &admin.ExecutionClosure{
					CreatedAt: addTimestamp(baseTimestamp, 10),
					StartedAt: addTimestamp(baseTimestamp, 15),
					Duration:  emptyDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 15),
				},
			},
			map[string][]int64{
				nodeSetup: []int64{10},
			},
			1,
		},
		{
			"Completed",
			&admin.NodeExecution{
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: emptyTimestamp,
					Duration:  emptyDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 440),
					TargetMetadata: &admin.NodeExecutionClosure_WorkflowNodeMetadata{
						WorkflowNodeMetadata: &admin.WorkflowNodeMetadata{
							ExecutionId: &core.WorkflowExecutionIdentifier{},
						},
					},
				},
			},
			&admin.Execution{
				Closure: &admin.ExecutionClosure{
					CreatedAt: addTimestamp(baseTimestamp, 10),
					StartedAt: addTimestamp(baseTimestamp, 15),
					Duration:  emptyDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 425),
				},
			},
			map[string][]int64{
				nodeSetup:    []int64{10},
				nodeTeardown: []int64{15},
			},
			1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// initialize mocks
			mockExecutionManager := getMockExecutionManager(test.execution)
			mockNodeExecutionManager := getMockNodeExecutionManager(
				[]*admin.NodeExecution{
					&admin.NodeExecution{
						Metadata: &admin.NodeExecutionMetaData{
							SpecNodeId: "start-node",
						},
						Closure: &admin.NodeExecutionClosure{
							CreatedAt: addTimestamp(baseTimestamp, 10),
							StartedAt: emptyTimestamp,
							Duration:  emptyDuration,
							UpdatedAt: addTimestamp(baseTimestamp, 10),
						},
					},
					&admin.NodeExecution{
						Metadata: &admin.NodeExecutionMetaData{
							SpecNodeId: "foo",
						},
						Closure: &admin.NodeExecutionClosure{
							CreatedAt: addTimestamp(baseTimestamp, 15),
							StartedAt: addTimestamp(baseTimestamp, 20),
							Duration:  baseDuration,
							UpdatedAt: addTimestamp(baseTimestamp, 435),
						},
					},
				}, nil)
			mockTaskExecutionManager := getMockTaskExecutionManager([]*admin.TaskExecution{})
			mockWorkflowManager := getMockWorkflowManager(
				&admin.Workflow{
					Closure: &admin.WorkflowClosure{
						CompiledWorkflow: &core.CompiledWorkflowClosure{
							Primary: &core.CompiledWorkflow{
								Connections: &core.ConnectionSet{
									Upstream: map[string]*core.ConnectionSet_IdList{
										"foo": &core.ConnectionSet_IdList{
											Ids: []string{"start-node"},
										},
										"end-node": &core.ConnectionSet_IdList{
											Ids: []string{"foo"},
										},
									},
								},
								Template: &core.WorkflowTemplate{
									Nodes: []*core.Node{
										&core.Node{
											Id:     "foo",
											Target: &core.Node_TaskNode{},
										},
									},
								},
							},
						},
					},
				})
			metricsManager := MetricsManager{
				executionManager:     mockExecutionManager,
				nodeExecutionManager: mockNodeExecutionManager,
				taskExecutionManager: mockTaskExecutionManager,
				workflowManager:      mockWorkflowManager,
			}

			// parse node execution
			spans := make([]*core.Span, 0)
			err := metricsManager.parseLaunchPlanNodeExecution(context.TODO(), test.nodeExecution, &spans, -1)
			assert.Nil(t, err)

			// validate spans
			operationDurations, referenceCount := parseSpans(spans)
			assert.True(t, reflect.DeepEqual(test.operationDurations, operationDurations))
			assert.Equal(t, test.referenceCount, referenceCount)
		})
	}
}

func TestParseSubworkflowNodeExecution(t *testing.T) {
	tests := []struct {
		name               string
		nodeExecution      *admin.NodeExecution
		nodeExecutions     []*admin.NodeExecution
		operationDurations map[string][]int64
		referenceCount     int
	}{
		{
			"NotStarted",
			&admin.NodeExecution{
				Id: &core.NodeExecutionIdentifier{
					ExecutionId: &core.WorkflowExecutionIdentifier{},
				},
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: emptyTimestamp,
					Duration:  emptyDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 5),
				},
			},
			nil,
			map[string][]int64{
				nodeSetup: []int64{5},
			},
			0,
		},
		{
			"Running",
			&admin.NodeExecution{
				Id: &core.NodeExecutionIdentifier{
					ExecutionId: &core.WorkflowExecutionIdentifier{},
				},
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: emptyTimestamp,
					Duration:  emptyDuration,
					UpdatedAt: baseTimestamp,
				},
			},
			[]*admin.NodeExecution{
				&admin.NodeExecution{
					Metadata: &admin.NodeExecutionMetaData{
						SpecNodeId: "start-node",
					},
					Closure: &admin.NodeExecutionClosure{
						CreatedAt: addTimestamp(baseTimestamp, 10),
						StartedAt: emptyTimestamp,
						Duration:  emptyDuration,
						UpdatedAt: addTimestamp(baseTimestamp, 10),
					},
				},
				&admin.NodeExecution{
					Metadata: &admin.NodeExecutionMetaData{
						SpecNodeId: "foo",
					},
					Closure: &admin.NodeExecutionClosure{
						CreatedAt: addTimestamp(baseTimestamp, 15),
						StartedAt: addTimestamp(baseTimestamp, 20),
						Duration:  baseDuration,
						UpdatedAt: addTimestamp(baseTimestamp, 435),
					},
				},
			},
			map[string][]int64{
				nodeSetup: []int64{10},
			},
			1,
		},
		{
			"Completed",
			&admin.NodeExecution{
				Id: &core.NodeExecutionIdentifier{
					ExecutionId: &core.WorkflowExecutionIdentifier{},
				},
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: emptyTimestamp,
					Duration:  emptyDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 455),
				},
			},
			[]*admin.NodeExecution{
				&admin.NodeExecution{
					Metadata: &admin.NodeExecutionMetaData{
						SpecNodeId: "start-node",
					},
					Closure: &admin.NodeExecutionClosure{
						CreatedAt: addTimestamp(baseTimestamp, 10),
						StartedAt: emptyTimestamp,
						Duration:  emptyDuration,
						UpdatedAt: addTimestamp(baseTimestamp, 10),
					},
				},
				&admin.NodeExecution{
					Metadata: &admin.NodeExecutionMetaData{
						SpecNodeId: "foo",
					},
					Closure: &admin.NodeExecutionClosure{
						CreatedAt: addTimestamp(baseTimestamp, 15),
						StartedAt: addTimestamp(baseTimestamp, 20),
						Duration:  baseDuration,
						UpdatedAt: addTimestamp(baseTimestamp, 435),
					},
				},
			},
			map[string][]int64{
				nodeSetup:    []int64{10},
				nodeTeardown: []int64{20},
			},
			1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// initialize mocks
			mockNodeExecutionManager := getMockNodeExecutionManager(test.nodeExecutions, nil)
			mockTaskExecutionManager := getMockTaskExecutionManager([]*admin.TaskExecution{})
			mockWorkflowManager := getMockWorkflowManager(
				&admin.Workflow{
					Closure: &admin.WorkflowClosure{
						CompiledWorkflow: &core.CompiledWorkflowClosure{
							Primary: &core.CompiledWorkflow{
								Connections: &core.ConnectionSet{
									Upstream: map[string]*core.ConnectionSet_IdList{
										"foo": &core.ConnectionSet_IdList{
											Ids: []string{"start-node"},
										},
										"end-node": &core.ConnectionSet_IdList{
											Ids: []string{"foo"},
										},
									},
								},
								Template: &core.WorkflowTemplate{
									Nodes: []*core.Node{
										&core.Node{
											Id:     "foo",
											Target: &core.Node_TaskNode{},
										},
									},
								},
							},
						},
					},
				})
			metricsManager := MetricsManager{
				nodeExecutionManager: mockNodeExecutionManager,
				taskExecutionManager: mockTaskExecutionManager,
				workflowManager:      mockWorkflowManager,
			}

			// parse node execution
			spans := make([]*core.Span, 0)
			err := metricsManager.parseSubworkflowNodeExecution(context.TODO(), test.nodeExecution, &core.Identifier{}, &spans, -1)
			assert.Nil(t, err)

			// validate spans
			operationDurations, referenceCount := parseSpans(spans)
			assert.True(t, reflect.DeepEqual(test.operationDurations, operationDurations))
			assert.Equal(t, test.referenceCount, referenceCount)
		})
	}
}

func TestParseTaskExecution(t *testing.T) {
	tests := []struct {
		name               string
		taskExecution      *admin.TaskExecution
		operationDurations map[string][]int64
	}{
		{
			"NotStarted",
			&admin.TaskExecution{
				Closure: &admin.TaskExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: emptyTimestamp,
					Duration:  emptyDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 5),
				},
			},
			map[string][]int64{
				taskSetup: []int64{5},
			},
		},
		{
			"Running",
			&admin.TaskExecution{
				Closure: &admin.TaskExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: addTimestamp(baseTimestamp, 5),
					Duration:  emptyDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 605),
				},
			},
			map[string][]int64{
				taskSetup:   []int64{5},
				taskRuntime: []int64{600},
			},
		},
		{
			"Completed",
			&admin.TaskExecution{
				Closure: &admin.TaskExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: addTimestamp(baseTimestamp, 5),
					Duration:  baseDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 415),
				},
			},
			map[string][]int64{
				taskSetup:    []int64{5},
				taskRuntime:  []int64{400},
				taskTeardown: []int64{10},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// parse task execution
			span := parseTaskExecution(test.taskExecution)
			_, ok := span.Id.(*core.Span_TaskId)
			assert.True(t, ok)

			// validate spans
			operationDurations, referenceCount := parseSpans(span.Spans)
			assert.True(t, reflect.DeepEqual(test.operationDurations, operationDurations))
			assert.Equal(t, 0, referenceCount)
		})
	}
}

func TestParseTaskExecutions(t *testing.T) {
	tests := []struct {
		name               string
		taskExecutions     []*admin.TaskExecution
		operationDurations map[string][]int64
		referenceCount     int
	}{
		{
			"SingleAttempt",
			[]*admin.TaskExecution{
				&admin.TaskExecution{
					Closure: &admin.TaskExecutionClosure{
						CreatedAt: baseTimestamp,
						StartedAt: addTimestamp(baseTimestamp, 5),
						Duration:  baseDuration,
						UpdatedAt: addTimestamp(baseTimestamp, 415),
					},
				},
			},
			map[string][]int64{},
			1,
		},
		{
			"MultipleAttempts",
			[]*admin.TaskExecution{
				&admin.TaskExecution{
					Closure: &admin.TaskExecutionClosure{
						CreatedAt: baseTimestamp,
						StartedAt: addTimestamp(baseTimestamp, 5),
						Duration:  baseDuration,
						UpdatedAt: addTimestamp(baseTimestamp, 605),
					},
				},
				&admin.TaskExecution{
					Closure: &admin.TaskExecutionClosure{
						CreatedAt: addTimestamp(baseTimestamp, 625),
						StartedAt: addTimestamp(baseTimestamp, 630),
						Duration:  emptyDuration,
						UpdatedAt: addTimestamp(baseTimestamp, 630),
					},
				},
			},
			map[string][]int64{
				nodeReset: []int64{20},
			},
			2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// parse task executions
			spans := make([]*core.Span, 0)
			parseTaskExecutions(test.taskExecutions, &spans, -1)

			// validate spans
			operationDurations, referenceCount := parseSpans(spans)
			assert.True(t, reflect.DeepEqual(test.operationDurations, operationDurations))
			assert.Equal(t, test.referenceCount, referenceCount)
		})
	}
}

func TestParseTaskNodeExecution(t *testing.T) {
	tests := []struct {
		name               string
		nodeExecution      *admin.NodeExecution
		taskExecutions     []*admin.TaskExecution
		operationDurations map[string][]int64
		referenceCount     int
	}{
		{
			"NotStarted",
			&admin.NodeExecution{
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: emptyTimestamp,
					Duration:  emptyDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 5),
				},
			},
			nil,
			map[string][]int64{
				nodeSetup: []int64{5},
			},
			0,
		},
		{
			"Running",
			&admin.NodeExecution{
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: emptyTimestamp,
					Duration:  emptyDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 10),
				},
			},
			[]*admin.TaskExecution{
				&admin.TaskExecution{
					Closure: &admin.TaskExecutionClosure{
						CreatedAt: addTimestamp(baseTimestamp, 10),
						StartedAt: addTimestamp(baseTimestamp, 15),
						Duration:  baseDuration,
						UpdatedAt: addTimestamp(baseTimestamp, 420),
					},
				},
			},
			map[string][]int64{
				nodeSetup: []int64{10},
			},
			1,
		},
		{
			"Completed",
			&admin.NodeExecution{
				Closure: &admin.NodeExecutionClosure{
					CreatedAt: baseTimestamp,
					StartedAt: emptyTimestamp,
					Duration:  emptyDuration,
					UpdatedAt: addTimestamp(baseTimestamp, 435),
				},
			},
			[]*admin.TaskExecution{
				&admin.TaskExecution{
					Closure: &admin.TaskExecutionClosure{
						CreatedAt: addTimestamp(baseTimestamp, 10),
						StartedAt: addTimestamp(baseTimestamp, 15),
						Duration:  baseDuration,
						UpdatedAt: addTimestamp(baseTimestamp, 420),
					},
				},
			},
			map[string][]int64{
				nodeSetup:    []int64{10},
				nodeTeardown: []int64{15},
			},
			1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// initialize mocks
			mockTaskExecutionManager := getMockTaskExecutionManager(test.taskExecutions)
			metricsManager := MetricsManager{
				taskExecutionManager: mockTaskExecutionManager,
			}

			// parse node execution
			spans := make([]*core.Span, 0)
			err := metricsManager.parseTaskNodeExecution(context.TODO(), test.nodeExecution, &spans, -1)
			assert.Nil(t, err)

			// validate spans
			operationDurations, referenceCount := parseSpans(spans)
			assert.True(t, reflect.DeepEqual(test.operationDurations, operationDurations))
			assert.Equal(t, test.referenceCount, referenceCount)
		})
	}
}
