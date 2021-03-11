package executions

import (
	"context"
	"testing"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	managerMocks "github.com/flyteorg/flyteadmin/pkg/manager/mocks"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	runtimeIFaceMocks "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces/mocks"
	"github.com/flyteorg/flyteadmin/pkg/runtime/mocks"
	runtimeMocks "github.com/flyteorg/flyteadmin/pkg/runtime/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
)

var workflowIdentifier = &core.Identifier{
	ResourceType: core.ResourceType_WORKFLOW,
	Project:      "project",
	Domain:       "development",
	Name:         "worky",
}

func getQualityOfServiceWithDuration(duration time.Duration) *core.QualityOfService {
	return &core.QualityOfService{
		Designation: &core.QualityOfService_Spec{
			Spec: &core.QualityOfServiceSpec{
				QueueingBudget: ptypes.DurationProto(duration),
			},
		},
	}
}

func getMockConfig() runtimeInterfaces.Configuration {
	mockConfig := mocks.NewMockConfigurationProvider(nil, nil, nil, nil, nil, nil)
	provider := &runtimeIFaceMocks.QualityOfServiceConfiguration{}
	provider.OnGetTierExecutionValues().Return(map[core.QualityOfService_Tier]core.QualityOfServiceSpec{
		core.QualityOfService_HIGH: {
			QueueingBudget: ptypes.DurationProto(10 * time.Minute),
		},
		core.QualityOfService_MEDIUM: {
			QueueingBudget: ptypes.DurationProto(20 * time.Minute),
		},
		core.QualityOfService_LOW: {
			QueueingBudget: ptypes.DurationProto(30 * time.Minute),
		},
	})

	provider.OnGetDefaultTiers().Return(map[string]core.QualityOfService_Tier{
		"production":  core.QualityOfService_HIGH,
		"development": core.QualityOfService_LOW,
	})

	mockConfig.(*runtimeMocks.MockConfigurationProvider).AddQualityOfServiceConfiguration(provider)
	return mockConfig
}

func addGetResourceFunc(t *testing.T, resourceManager interfaces.ResourceInterface) {
	resourceManager.(*managerMocks.MockResourceManager).GetResourceFunc = func(ctx context.Context,
		request interfaces.ResourceRequest) (*interfaces.ResourceResponse, error) {
		assert.EqualValues(t, request, interfaces.ResourceRequest{
			Project:      workflowIdentifier.Project,
			Domain:       workflowIdentifier.Domain,
			Workflow:     workflowIdentifier.Name,
			ResourceType: admin.MatchableResource_QUALITY_OF_SERVICE_SPECIFICATION,
		})
		return &interfaces.ResourceResponse{
			Attributes: &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_QualityOfService{
					QualityOfService: getQualityOfServiceWithDuration(5 * time.Minute),
				},
			},
		}, nil
	}
}

func getWorkflowWithQosSpec(qualityOfService *core.QualityOfService) *admin.Workflow {
	return &admin.Workflow{
		Id: workflowIdentifier,
		Closure: &admin.WorkflowClosure{
			CompiledWorkflow: &core.CompiledWorkflowClosure{
				Primary: &core.CompiledWorkflow{
					Template: &core.WorkflowTemplate{
						Metadata: &core.WorkflowMetadata{
							QualityOfService: qualityOfService,
						},
					},
				},
			},
		},
	}
}

func TestGetQualityOfService_ExecutionCreateRequest(t *testing.T) {
	resourceManager := managerMocks.MockResourceManager{}
	addGetResourceFunc(t, &resourceManager)

	allocator := NewQualityOfServiceAllocator(getMockConfig(), &resourceManager)
	spec, err := allocator.GetQualityOfService(context.Background(), GetQualityOfServiceInput{
		Workflow: getWorkflowWithQosSpec(getQualityOfServiceWithDuration(4 * time.Minute)),
		LaunchPlan: &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{
				QualityOfService: getQualityOfServiceWithDuration(2 * time.Minute),
			},
		},
		ExecutionCreateRequest: &admin.ExecutionCreateRequest{
			Domain: "production",
			Spec: &admin.ExecutionSpec{
				QualityOfService: getQualityOfServiceWithDuration(3 * time.Minute),
			},
		},
	})
	assert.Nil(t, err)
	assert.EqualValues(t, spec.QueuingBudget, 3*time.Minute)
}

func TestGetQualityOfService_LaunchPlan(t *testing.T) {
	resourceManager := managerMocks.MockResourceManager{}
	addGetResourceFunc(t, &resourceManager)

	allocator := NewQualityOfServiceAllocator(getMockConfig(), &resourceManager)
	spec, err := allocator.GetQualityOfService(context.Background(), GetQualityOfServiceInput{
		Workflow: getWorkflowWithQosSpec(getQualityOfServiceWithDuration(4 * time.Minute)),
		LaunchPlan: &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{
				QualityOfService: getQualityOfServiceWithDuration(2 * time.Minute),
			},
		},
		ExecutionCreateRequest: &admin.ExecutionCreateRequest{
			Domain: "production",
			Spec:   &admin.ExecutionSpec{},
		},
	})
	assert.Nil(t, err)
	assert.EqualValues(t, spec.QueuingBudget, 2*time.Minute)
}

func TestGetQualityOfService_Workflow(t *testing.T) {
	resourceManager := managerMocks.MockResourceManager{}
	addGetResourceFunc(t, &resourceManager)

	allocator := NewQualityOfServiceAllocator(getMockConfig(), &resourceManager)
	spec, err := allocator.GetQualityOfService(context.Background(), GetQualityOfServiceInput{
		Workflow: getWorkflowWithQosSpec(getQualityOfServiceWithDuration(4 * time.Minute)),
		LaunchPlan: &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{},
		},
		ExecutionCreateRequest: &admin.ExecutionCreateRequest{
			Domain: "production",
			Spec:   &admin.ExecutionSpec{},
		},
	})
	assert.Nil(t, err)
	assert.EqualValues(t, spec.QueuingBudget, 4*time.Minute)
}

func TestGetQualityOfService_MatchableResource(t *testing.T) {
	resourceManager := managerMocks.MockResourceManager{}
	addGetResourceFunc(t, &resourceManager)

	allocator := NewQualityOfServiceAllocator(getMockConfig(), &resourceManager)
	spec, err := allocator.GetQualityOfService(context.Background(), GetQualityOfServiceInput{
		Workflow: getWorkflowWithQosSpec(nil),
		LaunchPlan: &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{},
		},
		ExecutionCreateRequest: &admin.ExecutionCreateRequest{
			Domain: "production",
			Spec:   &admin.ExecutionSpec{},
		},
	})
	assert.Nil(t, err)
	assert.EqualValues(t, spec.QueuingBudget, 5*time.Minute)
}

func TestGetQualityOfService_ConfigValues(t *testing.T) {
	resourceManager := managerMocks.MockResourceManager{}

	allocator := NewQualityOfServiceAllocator(getMockConfig(), &resourceManager)
	spec, err := allocator.GetQualityOfService(context.Background(), GetQualityOfServiceInput{
		Workflow: getWorkflowWithQosSpec(nil),
		LaunchPlan: &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{},
		},
		ExecutionCreateRequest: &admin.ExecutionCreateRequest{
			Domain: "production",
			Spec:   &admin.ExecutionSpec{},
		},
	})
	assert.Nil(t, err)
	assert.EqualValues(t, spec.QueuingBudget, 10*time.Minute)
}

func TestGetQualityOfService_NoDefault(t *testing.T) {
	resourceManager := managerMocks.MockResourceManager{}

	allocator := NewQualityOfServiceAllocator(getMockConfig(), &resourceManager)
	spec, err := allocator.GetQualityOfService(context.Background(), GetQualityOfServiceInput{
		Workflow: getWorkflowWithQosSpec(nil),
		LaunchPlan: &admin.LaunchPlan{
			Spec: &admin.LaunchPlanSpec{},
		},
		ExecutionCreateRequest: &admin.ExecutionCreateRequest{
			Domain: "staging", // Nothing configured to match in the application config.
			Spec:   &admin.ExecutionSpec{},
		},
	})
	assert.Nil(t, err)
	assert.EqualValues(t, spec.QueuingBudget.Seconds(), 0)
}
