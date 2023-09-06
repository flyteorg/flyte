package transformers

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/common/testutils"

	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"

	"github.com/stretchr/testify/assert"
)

const resourceProject = "project"
const resourceDomain = "domain"
const resourceWorkflow = "workflow"

var matchingClusterResourceAttributes = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_ClusterResourceAttributes{
		ClusterResourceAttributes: &admin.ClusterResourceAttributes{
			Attributes: map[string]string{
				"foo": "bar",
			},
		},
	},
}

var workflowAttributes = admin.WorkflowAttributes{
	Project:            resourceProject,
	Domain:             resourceDomain,
	Workflow:           resourceWorkflow,
	MatchingAttributes: matchingClusterResourceAttributes,
}

var marshalledClusterResourceAttributes, _ = proto.Marshal(matchingClusterResourceAttributes)

var matchingExecutionQueueAttributes = &admin.MatchingAttributes{
	Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
		ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
			Tags: []string{
				"foo",
			},
		},
	},
}

var projectDomainAttributes = admin.ProjectDomainAttributes{
	Project:            resourceProject,
	Domain:             resourceDomain,
	MatchingAttributes: matchingExecutionQueueAttributes,
}

var marshalledExecutionQueueAttributes, _ = proto.Marshal(matchingExecutionQueueAttributes)

func TestToProjectDomainAttributesModel(t *testing.T) {

	model, err := ProjectDomainAttributesToResourceModel(projectDomainAttributes, admin.MatchableResource_EXECUTION_QUEUE)
	assert.Nil(t, err)
	assert.EqualValues(t, models.Resource{
		Project:      resourceProject,
		Domain:       resourceDomain,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE.String(),
		Priority:     models.ResourcePriorityProjectDomainLevel,
		Attributes:   marshalledExecutionQueueAttributes,
	}, model)
}

func TestMergeUpdateProjectDomainAttributes(t *testing.T) {
	t.Run("plugin override", func(t *testing.T) {
		existingWorkflowAttributes, _ := proto.Marshal(testutils.GetPluginOverridesAttributes(map[string][]string{
			"python": {"plugin_a"},
			"hive":   {"plugin_b"},
		}))

		existingModel := models.Resource{
			ID:           1,
			Project:      resourceProject,
			Domain:       resourceDomain,
			Workflow:     resourceWorkflow,
			ResourceType: "PLUGIN_OVERRIDE",
			Attributes:   existingWorkflowAttributes,
		}
		mergeUpdatedModel, err := MergeUpdatePluginAttributes(context.Background(), existingModel,
			admin.MatchableResource_PLUGIN_OVERRIDE, &repoInterfaces.ResourceID{},
			testutils.GetPluginOverridesAttributes(map[string][]string{
				"sidecar": {"plugin_c"},
				"hive":    {"plugin_d"},
			}),
		)
		assert.NoError(t, err)
		var updatedAttributes admin.MatchingAttributes
		err = proto.Unmarshal(mergeUpdatedModel.Attributes, &updatedAttributes)
		assert.NoError(t, err)
		var sawPythonTask, sawSidecarTask, sawHiveTask bool
		for _, override := range updatedAttributes.GetPluginOverrides().GetOverrides() {
			if override.TaskType == "python" {
				sawPythonTask = true
				assert.EqualValues(t, []string{"plugin_a"}, override.PluginId)
			} else if override.TaskType == "sidecar" {
				sawSidecarTask = true
				assert.EqualValues(t, []string{"plugin_c"}, override.PluginId)
			} else if override.TaskType == "hive" {
				sawHiveTask = true
				assert.EqualValues(t, []string{"plugin_d"}, override.PluginId)
			}
		}
		assert.True(t, sawPythonTask, "Missing python task from finalized attributes")
		assert.True(t, sawSidecarTask, "Missing sidecar task from finalized attributes")
		assert.True(t, sawHiveTask, "Missing hive task from finalized attributes")
	})
	t.Run("unsupported resource type", func(t *testing.T) {
		existingModel := models.Resource{
			ID:           1,
			Project:      resourceProject,
			Domain:       resourceDomain,
			Workflow:     resourceWorkflow,
			ResourceType: "PLUGIN_OVERRIDE",
		}
		_, err := MergeUpdatePluginAttributes(context.Background(), existingModel,
			admin.MatchableResource_TASK_RESOURCE, &repoInterfaces.ResourceID{}, &admin.MatchingAttributes{})
		assert.Error(t, err, "unsupported resource type")
	})
}

func TestFromProjectDomainAttributesModel(t *testing.T) {
	model := models.Resource{
		Project:      resourceProject,
		Domain:       resourceDomain,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes:   marshalledExecutionQueueAttributes,
	}
	unmarshalledAttributes, err := FromResourceModelToProjectDomainAttributes(model)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&projectDomainAttributes, &unmarshalledAttributes))
}

func TestFromProjectDomainAttributesModel_InvalidResourceAttributes(t *testing.T) {
	model := models.Resource{
		Project:      resourceProject,
		Domain:       resourceDomain,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes:   []byte("i'm invalid!"),
	}
	_, err := FromResourceModelToProjectDomainAttributes(model)
	assert.NotNil(t, err)
	assert.Equal(t, codes.Internal, err.(errors.FlyteAdminError).Code())
}

func TestToWorkflowAttributesModel(t *testing.T) {
	model, err := WorkflowAttributesToResourceModel(workflowAttributes, admin.MatchableResource_EXECUTION_QUEUE)
	assert.Nil(t, err)
	assert.EqualValues(t, models.Resource{
		Project:      resourceProject,
		Domain:       resourceDomain,
		Workflow:     resourceWorkflow,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE.String(),
		Priority:     models.ResourcePriorityWorkflowLevel,
		Attributes:   marshalledClusterResourceAttributes,
	}, model)
}

func TestMergeUpdateWorkflowAttributes(t *testing.T) {
	t.Run("plugin override", func(t *testing.T) {
		existingWorkflowAttributes, _ := proto.Marshal(testutils.GetPluginOverridesAttributes(map[string][]string{
			"python": {"plugin_a"},
			"hive":   {"plugin_b"},
		}))

		existingModel := models.Resource{
			ID:           1,
			Project:      resourceProject,
			Domain:       resourceDomain,
			Workflow:     resourceWorkflow,
			ResourceType: "PLUGIN_OVERRIDE",
			Attributes:   existingWorkflowAttributes,
		}
		mergeUpdatedModel, err := MergeUpdateWorkflowAttributes(context.Background(), existingModel,
			admin.MatchableResource_PLUGIN_OVERRIDE, &repoInterfaces.ResourceID{}, &admin.WorkflowAttributes{
				Project:  resourceProject,
				Domain:   resourceDomain,
				Workflow: resourceWorkflow,
				MatchingAttributes: testutils.GetPluginOverridesAttributes(map[string][]string{
					"sidecar": {"plugin_c"},
					"hive":    {"plugin_d"},
				}),
			})
		assert.NoError(t, err)
		var updatedAttributes admin.MatchingAttributes
		err = proto.Unmarshal(mergeUpdatedModel.Attributes, &updatedAttributes)
		assert.NoError(t, err)
		var sawPythonTask, sawSidecarTask, sawHiveTask bool
		for _, override := range updatedAttributes.GetPluginOverrides().GetOverrides() {
			if override.TaskType == "python" {
				sawPythonTask = true
				assert.EqualValues(t, []string{"plugin_a"}, override.PluginId)
			} else if override.TaskType == "sidecar" {
				sawSidecarTask = true
				assert.EqualValues(t, []string{"plugin_c"}, override.PluginId)
			} else if override.TaskType == "hive" {
				sawHiveTask = true
				assert.EqualValues(t, []string{"plugin_d"}, override.PluginId)
			}
		}
		assert.True(t, sawPythonTask, "Missing python task from finalized attributes")
		assert.True(t, sawSidecarTask, "Missing sidecar task from finalized attributes")
		assert.True(t, sawHiveTask, "Missing hive task from finalized attributes")
	})
	t.Run("unsupported resource type", func(t *testing.T) {
		existingModel := models.Resource{
			ID:           1,
			Project:      resourceProject,
			Domain:       resourceDomain,
			Workflow:     resourceWorkflow,
			ResourceType: "TASK_RESOURCE",
		}
		_, err := MergeUpdateWorkflowAttributes(context.Background(), existingModel,
			admin.MatchableResource_TASK_RESOURCE, &repoInterfaces.ResourceID{}, &admin.WorkflowAttributes{})
		assert.Error(t, err, "unsupported resource type")
	})
}

func TestFromWorkflowAttributesModel(t *testing.T) {
	model := models.Resource{
		Project:      resourceProject,
		Domain:       resourceDomain,
		Workflow:     "workflow",
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes:   marshalledClusterResourceAttributes,
	}
	unmarshalledAttributes, err := FromResourceModelToWorkflowAttributes(model)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&workflowAttributes, &unmarshalledAttributes))
}

func TestFromWorkflowAttributesModel_InvalidResourceAttributes(t *testing.T) {
	model := models.Resource{
		Project:      resourceProject,
		Domain:       resourceDomain,
		Workflow:     "workflow",
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes:   []byte("i'm invalid!"),
	}
	_, err := FromResourceModelToWorkflowAttributes(model)
	assert.NotNil(t, err)
	assert.Equal(t, codes.Internal, err.(errors.FlyteAdminError).Code())
}

func TestProjectAttributesToResourceModel(t *testing.T) {
	pa := admin.ProjectAttributes{
		Project:            resourceProject,
		MatchingAttributes: matchingClusterResourceAttributes,
	}
	rm, err := ProjectAttributesToResourceModel(pa, admin.MatchableResource_CLUSTER_RESOURCE)

	assert.NoError(t, err)
	assert.EqualValues(t, models.Resource{
		Project:      resourceProject,
		Domain:       "",
		ResourceType: admin.MatchableResource_CLUSTER_RESOURCE.String(),
		Priority:     models.ResourcePriorityProjectLevel,
		Attributes:   marshalledClusterResourceAttributes,
	}, rm)
}
