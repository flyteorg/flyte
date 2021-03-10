package compiler

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestGetRequirements(t *testing.T) {
	g := &core.WorkflowTemplate{
		Nodes: []*core.Node{
			{
				Target: &core.Node_TaskNode{
					TaskNode: &core.TaskNode{
						Reference: &core.TaskNode_ReferenceId{
							ReferenceId: &core.Identifier{Name: "Task_1"},
						},
					},
				},
			},
			{
				Target: &core.Node_TaskNode{
					TaskNode: &core.TaskNode{
						Reference: &core.TaskNode_ReferenceId{
							ReferenceId: &core.Identifier{Name: "Task_2"},
						},
					},
				},
			},
			{
				Target: &core.Node_WorkflowNode{
					WorkflowNode: &core.WorkflowNode{
						Reference: &core.WorkflowNode_LaunchplanRef{
							LaunchplanRef: &core.Identifier{Name: "Graph_1"},
						},
					},
				},
			},
			{
				Target: &core.Node_BranchNode{
					BranchNode: &core.BranchNode{
						IfElse: &core.IfElseBlock{
							Case: &core.IfBlock{
								ThenNode: &core.Node{
									Target: &core.Node_WorkflowNode{
										WorkflowNode: &core.WorkflowNode{
											Reference: &core.WorkflowNode_LaunchplanRef{
												LaunchplanRef: &core.Identifier{Name: "Graph_1"},
											},
										},
									},
								},
							},
							Other: []*core.IfBlock{
								{
									ThenNode: &core.Node{
										Target: &core.Node_TaskNode{
											TaskNode: &core.TaskNode{
												Reference: &core.TaskNode_ReferenceId{
													ReferenceId: &core.Identifier{Name: "Task_3"},
												},
											},
										},
									},
								},
								{
									ThenNode: &core.Node{
										Target: &core.Node_BranchNode{
											BranchNode: &core.BranchNode{
												IfElse: &core.IfElseBlock{
													Case: &core.IfBlock{
														ThenNode: &core.Node{
															Target: &core.Node_WorkflowNode{
																WorkflowNode: &core.WorkflowNode{
																	Reference: &core.WorkflowNode_LaunchplanRef{
																		LaunchplanRef: &core.Identifier{Name: "Graph_2"},
																	},
																},
															},
														},
													},
													Other: []*core.IfBlock{
														{
															ThenNode: &core.Node{
																Target: &core.Node_TaskNode{
																	TaskNode: &core.TaskNode{
																		Reference: &core.TaskNode_ReferenceId{
																			ReferenceId: &core.Identifier{Name: "Task_4"},
																		},
																	},
																},
															},
														},
														{
															ThenNode: &core.Node{
																Target: &core.Node_TaskNode{
																	TaskNode: &core.TaskNode{
																		Reference: &core.TaskNode_ReferenceId{
																			ReferenceId: &core.Identifier{Name: "Task_5"},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	subWorkflows := make([]*core.WorkflowTemplate, 0)
	reqs, err := GetRequirements(g, subWorkflows)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(reqs.GetRequiredTaskIds()))
	assert.Equal(t, 2, len(reqs.GetRequiredLaunchPlanIds()))
}
