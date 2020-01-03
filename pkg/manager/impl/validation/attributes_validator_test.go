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

func TestValidateProjectAttributesUpdateRequest(t *testing.T) {
	_, err := ValidateProjectAttributesUpdateRequest(admin.ProjectAttributesUpdateRequest{})
	assert.Equal(t, "missing attributes", err.Error())

	_, err = ValidateProjectAttributesUpdateRequest(admin.ProjectAttributesUpdateRequest{
		Attributes: &admin.ProjectAttributes{}})
	assert.Equal(t, "missing project", err.Error())

	matchableResource, err := ValidateProjectAttributesUpdateRequest(admin.ProjectAttributesUpdateRequest{
		Attributes: &admin.ProjectAttributes{
			Project: "project",
			MatchingAttributes: &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_TaskResourceAttributes{
					TaskResourceAttributes: &admin.TaskResourceAttributes{
						Defaults: &admin.TaskResourceSpec{
							Cpu: "1",
						},
					},
				},
			},
		}})
	assert.Equal(t, admin.MatchableResource_TASK_RESOURCE, matchableResource)
	assert.Nil(t, err)
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
