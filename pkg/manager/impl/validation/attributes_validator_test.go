package validation

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

var attributesApplicationConfigProvider = testutils.GetApplicationConfigWithDefaultDomains()

func TestValidateMatchingAttributes(t *testing.T) {
	testCases := []struct {
		attributes                *admin.MatchingAttributes
		expectedMatchableResource admin.MatchableResource
		expectedErr               error
	}{
		{
			nil,
			defaultMatchableResource,
			errors.NewFlyteAdminErrorf(codes.InvalidArgument, "missing matching_attributes"),
		},
		{
			&admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_TaskResourceAttributes{
					TaskResourceAttributes: &admin.TaskResourceAttributes{
						Defaults: &admin.TaskResourceSpec{
							Cpu: "1",
						},
					},
				},
			},
			admin.MatchableResource_TASK_RESOURCE,
			nil,
		},
		{
			&admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_ClusterResourceAttributes{
					ClusterResourceAttributes: &admin.ClusterResourceAttributes{
						Attributes: map[string]string{
							"bar": "baz",
						},
					},
				},
			},
			admin.MatchableResource_CLUSTER_RESOURCE,
			nil,
		},
		{
			&admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
					ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
						Tags: []string{"bar", "baz"},
					},
				},
			},
			admin.MatchableResource_EXECUTION_QUEUE,
			nil,
		},
		{
			&admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_PluginOverrides{
					PluginOverrides: &admin.PluginOverrides{
						Overrides: []*admin.PluginOverride{
							{
								TaskType: "python",
								PluginId: []string{"foo"},
							},
						},
					},
				},
			},
			admin.MatchableResource_PLUGIN_OVERRIDE,
			nil,
		},
		{
			&admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
					WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
						MaxParallelism: 100,
					},
				},
			},
			admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG,
			nil,
		},
		{
			&admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_ClusterAssignment{
					ClusterAssignment: &admin.ClusterAssignment{
						ClusterPoolName: "gpu",
					},
				},
			},
			admin.MatchableResource_CLUSTER_ASSIGNMENT,
			nil,
		},
	}
	for _, tc := range testCases {
		matchableResource, err := validateMatchingAttributes(tc.attributes, "foo")
		assert.Equal(t, tc.expectedMatchableResource, matchableResource)
		assert.EqualValues(t, tc.expectedErr, err)
	}
}

func TestValidateProjectDomainAttributesUpdateRequest(t *testing.T) {
	_, err := ValidateProjectDomainAttributesUpdateRequest(context.Background(),
		testutils.GetRepoWithDefaultProject(), attributesApplicationConfigProvider,
		admin.ProjectDomainAttributesUpdateRequest{})
	assert.Equal(t, "missing attributes", err.Error())

	_, err = ValidateProjectDomainAttributesUpdateRequest(context.Background(),
		testutils.GetRepoWithDefaultProject(), attributesApplicationConfigProvider,
		admin.ProjectDomainAttributesUpdateRequest{
			Attributes: &admin.ProjectDomainAttributes{}})
	assert.Equal(t, "domain [] is unrecognized by system", err.Error())

	_, err = ValidateProjectDomainAttributesUpdateRequest(context.Background(),
		testutils.GetRepoWithDefaultProjectAndErr(shared.GetMissingArgumentError("project")), attributesApplicationConfigProvider,
		admin.ProjectDomainAttributesUpdateRequest{
			Attributes: &admin.ProjectDomainAttributes{
				Domain: "development",
			}})
	assert.Equal(t,
		"failed to validate that project [] and domain [development] are registered, err: [missing project]",
		err.Error())

	matchableResource, err := ValidateProjectDomainAttributesUpdateRequest(context.Background(),
		testutils.GetRepoWithDefaultProject(), attributesApplicationConfigProvider,
		admin.ProjectDomainAttributesUpdateRequest{
			Attributes: &admin.ProjectDomainAttributes{
				Project: "project",
				Domain:  "domain",
				MatchingAttributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_ClusterResourceAttributes{
						ClusterResourceAttributes: &admin.ClusterResourceAttributes{
							Attributes: map[string]string{
								"bar": "baz",
							},
						},
					},
				},
			}})
	assert.Equal(t, admin.MatchableResource_CLUSTER_RESOURCE, matchableResource)
	assert.Nil(t, err)
}

func TestValidateProjectDomainAttributesGetRequest(t *testing.T) {
	err := ValidateProjectDomainAttributesGetRequest(context.Background(),
		testutils.GetRepoWithDefaultProject(), attributesApplicationConfigProvider,
		admin.ProjectDomainAttributesGetRequest{})
	assert.Equal(t, "domain [] is unrecognized by system", err.Error())

	err = ValidateProjectDomainAttributesGetRequest(context.Background(),
		testutils.GetRepoWithDefaultProjectAndErr(shared.GetMissingArgumentError("project")),
		attributesApplicationConfigProvider, admin.ProjectDomainAttributesGetRequest{
			Domain: "development",
		})
	assert.Equal(t, "failed to validate that project [] and domain [development] are registered, err: [missing project]",
		err.Error())

	assert.Nil(t, ValidateProjectDomainAttributesGetRequest(context.Background(),
		testutils.GetRepoWithDefaultProject(), attributesApplicationConfigProvider,
		admin.ProjectDomainAttributesGetRequest{
			Project: "project",
			Domain:  "domain",
		}))
}

func TestValidateProjectDomainAttributesDeleteRequest(t *testing.T) {
	err := ValidateProjectDomainAttributesDeleteRequest(context.Background(),
		testutils.GetRepoWithDefaultProject(), attributesApplicationConfigProvider,
		admin.ProjectDomainAttributesDeleteRequest{})
	assert.Equal(t, "domain [] is unrecognized by system", err.Error())

	err = ValidateProjectDomainAttributesDeleteRequest(context.Background(),
		testutils.GetRepoWithDefaultProjectAndErr(shared.GetMissingArgumentError("project")),
		attributesApplicationConfigProvider, admin.ProjectDomainAttributesDeleteRequest{
			Domain: "development",
		})
	assert.Equal(t,
		"failed to validate that project [] and domain [development] are registered, err: [missing project]",
		err.Error())

	assert.Nil(t, ValidateProjectDomainAttributesDeleteRequest(context.Background(),
		testutils.GetRepoWithDefaultProject(), attributesApplicationConfigProvider,
		admin.ProjectDomainAttributesDeleteRequest{
			Project: "project",
			Domain:  "domain",
		}))
}

func TestValidateWorkflowAttributesUpdateRequest(t *testing.T) {
	_, err := ValidateWorkflowAttributesUpdateRequest(context.Background(),
		testutils.GetRepoWithDefaultProject(), attributesApplicationConfigProvider,
		admin.WorkflowAttributesUpdateRequest{})
	assert.Equal(t, "missing attributes", err.Error())

	_, err = ValidateWorkflowAttributesUpdateRequest(context.Background(),
		testutils.GetRepoWithDefaultProject(), attributesApplicationConfigProvider,
		admin.WorkflowAttributesUpdateRequest{
			Attributes: &admin.WorkflowAttributes{}})
	assert.Equal(t, "domain [] is unrecognized by system", err.Error())

	_, err = ValidateWorkflowAttributesUpdateRequest(context.Background(),
		testutils.GetRepoWithDefaultProjectAndErr(shared.GetMissingArgumentError("project")),
		attributesApplicationConfigProvider, admin.WorkflowAttributesUpdateRequest{
			Attributes: &admin.WorkflowAttributes{
				Domain: "development",
			}})
	assert.Equal(t,
		"failed to validate that project [] and domain [development] are registered, err: [missing project]",
		err.Error())

	_, err = ValidateWorkflowAttributesUpdateRequest(context.Background(),
		testutils.GetRepoWithDefaultProject(), attributesApplicationConfigProvider,
		admin.WorkflowAttributesUpdateRequest{
			Attributes: &admin.WorkflowAttributes{
				Project: "project",
				Domain:  "domain",
			}})
	assert.Equal(t, "missing name", err.Error())

	matchableResource, err := ValidateWorkflowAttributesUpdateRequest(context.Background(),
		testutils.GetRepoWithDefaultProject(), attributesApplicationConfigProvider,
		admin.WorkflowAttributesUpdateRequest{
			Attributes: &admin.WorkflowAttributes{
				Project:  "project",
				Domain:   "domain",
				Workflow: "workflow",
				MatchingAttributes: &admin.MatchingAttributes{
					Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
						ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
							Tags: []string{"bar", "baz"},
						},
					},
				},
			}})
	assert.Equal(t, admin.MatchableResource_EXECUTION_QUEUE, matchableResource)
	assert.Nil(t, err)
}

func TestValidateWorkflowAttributesGetRequest(t *testing.T) {
	err := ValidateWorkflowAttributesGetRequest(context.Background(),
		testutils.GetRepoWithDefaultProject(), attributesApplicationConfigProvider,
		admin.WorkflowAttributesGetRequest{})
	assert.Equal(t, "domain [] is unrecognized by system", err.Error())

	err = ValidateWorkflowAttributesGetRequest(context.Background(),
		testutils.GetRepoWithDefaultProjectAndErr(shared.GetMissingArgumentError("project")),
		attributesApplicationConfigProvider, admin.WorkflowAttributesGetRequest{
			Domain: "development",
		})
	assert.Equal(t,
		"failed to validate that project [] and domain [development] are registered, err: [missing project]",
		err.Error())

	err = ValidateWorkflowAttributesGetRequest(context.Background(),
		testutils.GetRepoWithDefaultProject(), attributesApplicationConfigProvider,
		admin.WorkflowAttributesGetRequest{
			Project: "project",
			Domain:  "domain",
		})
	assert.Equal(t, "missing name", err.Error())

	assert.Nil(t, ValidateWorkflowAttributesGetRequest(context.Background(),
		testutils.GetRepoWithDefaultProject(), attributesApplicationConfigProvider,
		admin.WorkflowAttributesGetRequest{
			Project:  "project",
			Domain:   "domain",
			Workflow: "workflow",
		}))
}

func TestValidateWorkflowAttributesDeleteRequest(t *testing.T) {
	err := ValidateWorkflowAttributesDeleteRequest(context.Background(),
		testutils.GetRepoWithDefaultProject(), attributesApplicationConfigProvider,
		admin.WorkflowAttributesDeleteRequest{})
	assert.Equal(t, "domain [] is unrecognized by system", err.Error())

	err = ValidateWorkflowAttributesDeleteRequest(context.Background(),
		testutils.GetRepoWithDefaultProjectAndErr(shared.GetMissingArgumentError("project")),
		attributesApplicationConfigProvider, admin.WorkflowAttributesDeleteRequest{
			Domain: "development",
		})
	assert.Equal(t,
		"failed to validate that project [] and domain [development] are registered, err: [missing project]",
		err.Error())

	err = ValidateWorkflowAttributesDeleteRequest(context.Background(),
		testutils.GetRepoWithDefaultProject(), attributesApplicationConfigProvider,
		admin.WorkflowAttributesDeleteRequest{
			Project: "project",
			Domain:  "domain",
		})
	assert.Equal(t, "missing name", err.Error())

	assert.Nil(t, ValidateWorkflowAttributesDeleteRequest(context.Background(),
		testutils.GetRepoWithDefaultProject(), attributesApplicationConfigProvider,
		admin.WorkflowAttributesDeleteRequest{
			Project:  "project",
			Domain:   "domain",
			Workflow: "workflow",
		}))
}

func TestValidateListAllMatchableAttributesRequest(t *testing.T) {
	err := ValidateListAllMatchableAttributesRequest(admin.ListMatchableAttributesRequest{
		ResourceType: 44,
	})
	assert.EqualError(t, err, "invalid value for resource_type")

	err = ValidateListAllMatchableAttributesRequest(admin.ListMatchableAttributesRequest{
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	})
	assert.Nil(t, err)
}
