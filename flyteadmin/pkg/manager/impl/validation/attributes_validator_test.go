package validation

import (
	"testing"

	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestValidateMatchingAttributes(t *testing.T) {
	testCases := []struct {
		attributes                *admin.MatchingAttributes
		identifier                string
		expectedMatchableResource admin.MatchableResource
		expectedErr               error
	}{
		{
			nil,
			"foo",
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
			"foo",
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
			"foo",
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
			"foo",
			admin.MatchableResource_EXECUTION_QUEUE,
			nil,
		},
	}
	for _, tc := range testCases {
		matchableResource, err := validateMatchingAttributes(tc.attributes, tc.identifier)
		assert.Equal(t, tc.expectedMatchableResource, matchableResource)
		assert.EqualValues(t, tc.expectedErr, err)
	}
}

func TestValidateProjectDomainAttributesUpdateRequest(t *testing.T) {
	_, err := ValidateProjectDomainAttributesUpdateRequest(admin.ProjectDomainAttributesUpdateRequest{})
	assert.Equal(t, "missing attributes", err.Error())

	_, err = ValidateProjectDomainAttributesUpdateRequest(admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{}})
	assert.Equal(t, "missing project", err.Error())

	_, err = ValidateProjectDomainAttributesUpdateRequest(admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{
			Project: "project",
		}})
	assert.Equal(t, "missing domain", err.Error())

	matchableResource, err := ValidateProjectDomainAttributesUpdateRequest(admin.ProjectDomainAttributesUpdateRequest{
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
	err := ValidateProjectDomainAttributesGetRequest(admin.ProjectDomainAttributesGetRequest{})
	assert.Equal(t, "missing project", err.Error())

	err = ValidateProjectDomainAttributesGetRequest(admin.ProjectDomainAttributesGetRequest{
		Project: "project",
	})
	assert.Equal(t, "missing domain", err.Error())

	assert.Nil(t, ValidateProjectDomainAttributesGetRequest(admin.ProjectDomainAttributesGetRequest{
		Project: "project",
		Domain:  "domain",
	}))
}

func TestValidateProjectDomainAttributesDeleteRequest(t *testing.T) {
	err := ValidateProjectDomainAttributesDeleteRequest(admin.ProjectDomainAttributesDeleteRequest{})
	assert.Equal(t, "missing project", err.Error())

	err = ValidateProjectDomainAttributesDeleteRequest(admin.ProjectDomainAttributesDeleteRequest{
		Project: "project",
	})
	assert.Equal(t, "missing domain", err.Error())

	assert.Nil(t, ValidateProjectDomainAttributesDeleteRequest(admin.ProjectDomainAttributesDeleteRequest{
		Project: "project",
		Domain:  "domain",
	}))
}

func TestValidateWorkflowAttributesUpdateRequest(t *testing.T) {
	_, err := ValidateWorkflowAttributesUpdateRequest(admin.WorkflowAttributesUpdateRequest{})
	assert.Equal(t, "missing attributes", err.Error())

	_, err = ValidateWorkflowAttributesUpdateRequest(admin.WorkflowAttributesUpdateRequest{
		Attributes: &admin.WorkflowAttributes{}})
	assert.Equal(t, "missing project", err.Error())

	_, err = ValidateWorkflowAttributesUpdateRequest(admin.WorkflowAttributesUpdateRequest{
		Attributes: &admin.WorkflowAttributes{
			Project: "project",
		}})
	assert.Equal(t, "missing domain", err.Error())

	_, err = ValidateWorkflowAttributesUpdateRequest(admin.WorkflowAttributesUpdateRequest{
		Attributes: &admin.WorkflowAttributes{
			Project: "project",
			Domain:  "domain",
		}})
	assert.Equal(t, "missing name", err.Error())

	matchableResource, err := ValidateWorkflowAttributesUpdateRequest(admin.WorkflowAttributesUpdateRequest{
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
	err := ValidateWorkflowAttributesGetRequest(admin.WorkflowAttributesGetRequest{})
	assert.Equal(t, "missing project", err.Error())

	err = ValidateWorkflowAttributesGetRequest(admin.WorkflowAttributesGetRequest{
		Project: "project",
	})
	assert.Equal(t, "missing domain", err.Error())

	err = ValidateWorkflowAttributesGetRequest(admin.WorkflowAttributesGetRequest{
		Project: "project",
		Domain:  "domain",
	})
	assert.Equal(t, "missing name", err.Error())

	assert.Nil(t, ValidateWorkflowAttributesGetRequest(admin.WorkflowAttributesGetRequest{
		Project:  "project",
		Domain:   "domain",
		Workflow: "workflow",
	}))
}

func TestValidateWorkflowAttributesDeleteRequest(t *testing.T) {
	err := ValidateWorkflowAttributesDeleteRequest(admin.WorkflowAttributesDeleteRequest{})
	assert.Equal(t, "missing project", err.Error())

	err = ValidateWorkflowAttributesDeleteRequest(admin.WorkflowAttributesDeleteRequest{
		Project: "project",
	})
	assert.Equal(t, "missing domain", err.Error())

	err = ValidateWorkflowAttributesDeleteRequest(admin.WorkflowAttributesDeleteRequest{
		Project: "project",
		Domain:  "domain",
	})
	assert.Equal(t, "missing name", err.Error())

	assert.Nil(t, ValidateWorkflowAttributesDeleteRequest(admin.WorkflowAttributesDeleteRequest{
		Project:  "project",
		Domain:   "domain",
		Workflow: "workflow",
	}))
}
