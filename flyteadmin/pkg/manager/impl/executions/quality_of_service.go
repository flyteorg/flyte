package executions

import (
	"context"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
)

type QualityOfServiceSpec struct {
	QueuingBudget time.Duration
}

type GetQualityOfServiceInput struct {
	Workflow               *admin.Workflow
	LaunchPlan             *admin.LaunchPlan
	ExecutionCreateRequest *admin.ExecutionCreateRequest
}

type QualityOfServiceAllocator interface {
	GetQualityOfService(ctx context.Context, input GetQualityOfServiceInput) (QualityOfServiceSpec, error)
}

type qualityOfServiceAllocator struct {
	config          runtimeInterfaces.Configuration
	resourceManager interfaces.ResourceInterface
}

func (q qualityOfServiceAllocator) getQualityOfServiceFromDb(ctx context.Context, workflowIdentifier *core.Identifier) (
	*core.QualityOfService, error) {
	resource, err := q.resourceManager.GetResource(ctx, interfaces.ResourceRequest{
		Project:      workflowIdentifier.Project,
		Domain:       workflowIdentifier.Domain,
		Workflow:     workflowIdentifier.Name,
		ResourceType: admin.MatchableResource_QUALITY_OF_SERVICE_SPECIFICATION,
	})
	if err != nil {
		if _, ok := err.(errors.FlyteAdminError); !ok || err.(errors.FlyteAdminError).Code() != codes.NotFound {
			logger.Warningf(ctx,
				"Failed to fetch override values when assigning quality of service values for [%+v] with err: %v",
				workflowIdentifier, err)
			return nil, err
		}
		logger.Debugf(ctx, "No quality of service specified as an overridable matching attribute in db")
		return nil, nil
	}

	if resource != nil && resource.Attributes != nil && resource.Attributes.GetQualityOfService() != nil &&
		resource.Attributes.GetQualityOfService() != nil {
		// Use custom override value in database rather than from registered entities or the admin application config.
		return resource.Attributes.GetQualityOfService(), nil
	}
	logger.Warningf(ctx, "Empty quality of service specified as an overridable matching attribute in db")
	return nil, nil
}

/*
Users can specify the quality of service for an execution (in order of decreasing specificity)

  - At CreateExecution request time
  - In the LaunchPlan spec
  - In the Workflow spec
  - As an overridable MatchableResource (https://lyft.github.io/flyte/administrator/install/managing_customizable_resources.html)
    for the underlying workflow

System administrators can specify default QualityOfService specs
(https://github.com/flyteorg/flyteidl/blob/e9727afcedf8d4c30a1fc2eeac45593e426d9bb0/protos/flyteidl/core/execution.proto#L92)s
for different QualityOfService tiers. The execution domain determines the tier, which in turn determines the configured
QualityOfService spec to apply.

This method handles resolving the QualityOfService for an execution given the above rules.
*/
func (q qualityOfServiceAllocator) GetQualityOfService(ctx context.Context, input GetQualityOfServiceInput) (QualityOfServiceSpec, error) {
	workflowIdentifier := input.Workflow.Id

	var qualityOfServiceTier core.QualityOfService_Tier
	if input.ExecutionCreateRequest.Spec.QualityOfService != nil {
		if input.ExecutionCreateRequest.Spec.QualityOfService.GetSpec() != nil {
			logger.Debugf(ctx, "Determining quality of service from execution spec for [%s/%s/%s]",
				input.ExecutionCreateRequest.Project, input.ExecutionCreateRequest.Domain,
				input.ExecutionCreateRequest.Name)
			duration, err := ptypes.Duration(input.ExecutionCreateRequest.Spec.QualityOfService.GetSpec().QueueingBudget)
			if err != nil {
				return QualityOfServiceSpec{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"Invalid custom quality of service set in create execution request [%s/%s/%s], failed to parse duration [%v] with: %v",
					input.ExecutionCreateRequest.Project, input.ExecutionCreateRequest.Domain,
					input.ExecutionCreateRequest.Name,
					input.ExecutionCreateRequest.Spec.QualityOfService.GetSpec().QueueingBudget, err)
			}
			return QualityOfServiceSpec{
				QueuingBudget: duration,
			}, nil
		}
		qualityOfServiceTier = input.ExecutionCreateRequest.Spec.QualityOfService.GetTier()
	} else if input.LaunchPlan.Spec.QualityOfService != nil {
		if input.LaunchPlan.Spec.QualityOfService.GetSpec() != nil {
			logger.Debugf(ctx, "Determining quality of service from launch plan spec for [%s/%s/%s]",
				input.ExecutionCreateRequest.Project, input.ExecutionCreateRequest.Domain,
				input.ExecutionCreateRequest.Name)
			duration, err := ptypes.Duration(input.LaunchPlan.Spec.QualityOfService.GetSpec().QueueingBudget)
			if err != nil {
				return QualityOfServiceSpec{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"Invalid custom quality of service set in launch plan [%v], failed to parse duration [%v] with: %v",
					input.LaunchPlan.Id,
					input.ExecutionCreateRequest.Spec.QualityOfService.GetSpec().QueueingBudget, err)
			}
			return QualityOfServiceSpec{
				QueuingBudget: duration,
			}, nil
		}
		qualityOfServiceTier = input.LaunchPlan.Spec.QualityOfService.GetTier()
	} else if input.Workflow.Closure.CompiledWorkflow.Primary.Template.Metadata != nil &&
		input.Workflow.Closure.CompiledWorkflow.Primary.Template.Metadata.QualityOfService != nil {
		logger.Debugf(ctx, "Determining quality of service from workflow spec for [%s/%s/%s]",
			input.ExecutionCreateRequest.Project, input.ExecutionCreateRequest.Domain,
			input.ExecutionCreateRequest.Name)
		if input.Workflow.Closure.CompiledWorkflow.Primary.Template.Metadata.QualityOfService.GetSpec() != nil {
			duration, err := ptypes.Duration(input.Workflow.Closure.CompiledWorkflow.Primary.Template.Metadata.QualityOfService.
				GetSpec().QueueingBudget)
			if err != nil {
				return QualityOfServiceSpec{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"Invalid custom quality of service set in workflow [%v], failed to parse duration [%v] with: %v",
					workflowIdentifier,
					input.ExecutionCreateRequest.Spec.QualityOfService.GetSpec().QueueingBudget, err)
			}
			return QualityOfServiceSpec{
				QueuingBudget: duration,
			}, nil
		}
		qualityOfServiceTier = input.Workflow.Closure.CompiledWorkflow.Primary.Template.Metadata.QualityOfService.GetTier()
	}

	// If nothing in the hierarchy of registrable entities has set the quality of service,
	// see if an override exists in the matchable attributes resource table.
	if qualityOfServiceTier == core.QualityOfService_UNDEFINED {
		qualityOfService, err := q.getQualityOfServiceFromDb(ctx, workflowIdentifier)
		if err != nil {
			return QualityOfServiceSpec{}, err
		}
		if qualityOfService != nil && qualityOfService.GetSpec() != nil {
			logger.Debugf(ctx, "Determining quality of service from spec database override for [%s/%s/%s]",
				input.ExecutionCreateRequest.Project, input.ExecutionCreateRequest.Domain,
				input.ExecutionCreateRequest.Name)
			duration, err := ptypes.Duration(qualityOfService.GetSpec().QueueingBudget)
			if err != nil {
				return QualityOfServiceSpec{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"Invalid custom quality of service set in overridable matching attributes for [%v],"+
						"failed to parse duration [%v] with: %v", workflowIdentifier,
					input.ExecutionCreateRequest.Spec.QualityOfService.GetSpec().QueueingBudget, err)
			}
			return QualityOfServiceSpec{
				QueuingBudget: duration,
			}, nil
		} else if qualityOfService != nil && qualityOfService.GetTier() != core.QualityOfService_UNDEFINED {
			logger.Debugf(ctx, "Determining quality of service tier from database override for [%s/%s/%s]",
				input.ExecutionCreateRequest.Project, input.ExecutionCreateRequest.Domain,
				input.ExecutionCreateRequest.Name)
			qualityOfServiceTier = input.Workflow.Closure.CompiledWorkflow.Primary.Template.Metadata.QualityOfService.GetTier()
		}
	}

	// If we've come all this way and at no layer is an overridable configuration for the quality of service tier
	// set, use the default values from the admin application config.
	if qualityOfServiceTier == core.QualityOfService_UNDEFINED {
		logger.Debugf(ctx, "Determining quality of service tier from application config override for [%s/%s/%s]",
			input.ExecutionCreateRequest.Project, input.ExecutionCreateRequest.Domain,
			input.ExecutionCreateRequest.Name)
		var ok bool
		qualityOfServiceTier, ok = q.config.QualityOfServiceConfiguration().GetDefaultTiers()[input.ExecutionCreateRequest.Domain]
		if !ok {
			// No queueing budget to set when no default is specified
			return QualityOfServiceSpec{}, nil
		}
	}
	executionValues, ok := q.config.QualityOfServiceConfiguration().GetTierExecutionValues()[qualityOfServiceTier]
	if !ok {
		// No queueing budget to set when no default is specified
		return QualityOfServiceSpec{}, nil
	}
	logger.Debugf(ctx, "Determining quality of service spec from application config override for [%s/%s/%s] with tier [%v]",
		input.ExecutionCreateRequest.Project, input.ExecutionCreateRequest.Domain,
		input.ExecutionCreateRequest.Name, qualityOfServiceTier)
	// Config values should always be vetted so there's no need to check the error from conversion.
	duration, _ := ptypes.Duration(executionValues.QueueingBudget)

	return QualityOfServiceSpec{
		QueuingBudget: duration,
	}, nil
}

func NewQualityOfServiceAllocator(config runtimeInterfaces.Configuration, resourceManager interfaces.ResourceInterface) QualityOfServiceAllocator {
	return &qualityOfServiceAllocator{
		config:          config,
		resourceManager: resourceManager,
	}
}
