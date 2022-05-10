//go:build integration
// +build integration

// Shared test code values.
package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/stretchr/testify/assert"
)

const project, domain, name = "project", "domain", "execution name"

var entityProjects = []string{"admintests"}
var entityDomains = []string{"development", "production"}
var entityNames = []string{"name_a", "name_b", "name_c"}
var entityVersions = []string{"123", "456", "789"}

func addHTTPRequestHeaders(req *http.Request) {
	req.Header.Add("Accept", "application/octet-stream")
	req.Header.Add("Content-Type", "application/json")
}

func insertTasksForTests(t *testing.T, client service.AdminServiceClient) {
	ctx := context.Background()
	for _, project := range entityProjects {
		for _, domain := range entityDomains {
			for _, name := range entityNames {
				for _, version := range entityVersions {
					req := admin.TaskCreateRequest{
						Id: &core.Identifier{
							ResourceType: core.ResourceType_TASK,
							Project:      project,
							Domain:       domain,
							Name:         name,
							Version:      version,
						},
						Spec: testutils.GetValidTaskRequest().Spec,
					}

					_, err := client.CreateTask(ctx, &req)
					assert.Nil(t, err)
				}
			}
		}
	}
}

func insertWorkflowsForTests(t *testing.T, client service.AdminServiceClient) {
	ctx := context.Background()
	for _, project := range entityProjects {
		for _, domain := range entityDomains {
			for _, name := range entityNames {
				for _, version := range entityVersions {
					identifier := core.Identifier{
						ResourceType: core.ResourceType_WORKFLOW,
						Project:      project,
						Domain:       domain,
						Name:         name,
						Version:      version,
					}
					req := admin.WorkflowCreateRequest{
						Id: &identifier,
						Spec: &admin.WorkflowSpec{
							Template: &core.WorkflowTemplate{
								Id: &identifier,
								Interface: &core.TypedInterface{
									Inputs: &core.VariableMap{
										Variables: map[string]*core.Variable{
											"foo": {
												Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
											},
										},
									},
								},
								Nodes: []*core.Node{
									{
										Id: "I'm a node",
										Target: &core.Node_TaskNode{
											TaskNode: &core.TaskNode{
												Reference: &core.TaskNode_ReferenceId{
													ReferenceId: &core.Identifier{
														ResourceType: core.ResourceType_TASK,
														Project:      project,
														Domain:       domain,
														Name:         name,
														Version:      version,
													},
												},
											},
										},
									},
								},
							},
						},
					}

					_, err := client.CreateWorkflow(ctx, &req)
					assert.Nil(t, err, "Failed to create workflow test data with err %v", err)
				}
			}
		}
	}
}

func insertLaunchPlansForTests(t *testing.T, client service.AdminServiceClient) {
	ctx := context.Background()
	for _, project := range entityProjects {
		for _, domain := range entityDomains {
			for _, name := range entityNames {
				for _, version := range entityVersions {
					workflowIdentifier := core.Identifier{
						ResourceType: core.ResourceType_WORKFLOW,
						Project:      project,
						Domain:       domain,
						Name:         name,
						Version:      version,
					}
					req := admin.LaunchPlanCreateRequest{
						Id: &core.Identifier{
							ResourceType: core.ResourceType_LAUNCH_PLAN,
							Project:      workflowIdentifier.Project,
							Domain:       workflowIdentifier.Domain,
							Name:         workflowIdentifier.Name,
							Version:      workflowIdentifier.Version,
						},
						Spec: &admin.LaunchPlanSpec{
							WorkflowId: &workflowIdentifier,
							DefaultInputs: &core.ParameterMap{
								Parameters: map[string]*core.Parameter{
									"foo": {
										Var: &core.Variable{
											Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
										},
										Behavior: &core.Parameter_Default{
											Default: coreutils.MustMakeLiteral("foo-value"),
										},
									},
								},
							},
						},
					}

					_, err := client.CreateLaunchPlan(ctx, &req)
					assert.Nil(t, err, "Failed to create launch plan test data with err %v", err)
				}
			}
		}
	}
}
