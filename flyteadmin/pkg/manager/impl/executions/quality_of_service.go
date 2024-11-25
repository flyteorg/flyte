package executions

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
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
		Project:      workflowIdentifier.GetProject(),
		Domain:       workflowIdentifier.GetDomain(),
		Workflow:     workflowIdentifier.GetName(),
		ResourceType: admin.MatchableResource_QUALITY_OF_SERVICE_SPECIFICATION,
	})
	if err != nil {
		if !errors.IsDoesNotExistError(err) {
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
	workflowIdentifier := input.Workflow.GetId()

	var qualityOfServiceTier core.QualityOfService_Tier
	if input.ExecutionCreateRequest.GetSpec().GetQualityOfService() != nil {
		if input.ExecutionCreateRequest.GetSpec().GetQualityOfService().GetSpec() != nil {
			logger.Debugf(ctx, "Determining quality of service from execution spec for [%s/%s/%s]",
				input.ExecutionCreateRequest.GetProject(), input.ExecutionCreateRequest.GetDomain(),
				input.ExecutionCreateRequest.GetName())
			duration, err := ptypes.Duration(input.ExecutionCreateRequest.GetSpec().GetQualityOfService().GetSpec().GetQueueingBudget())
			if err != nil {
				return QualityOfServiceSpec{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"Invalid custom quality of service set in create execution request [%s/%s/%s], failed to parse duration [%v] with: %v",
					input.ExecutionCreateRequest.GetProject(), input.ExecutionCreateRequest.GetDomain(),
					input.ExecutionCreateRequest.GetName(),
					input.ExecutionCreateRequest.GetSpec().GetQualityOfService().GetSpec().GetQueueingBudget(), err)
			}
			return QualityOfServiceSpec{
				QueuingBudget: duration,
			}, nil
		}
		qualityOfServiceTier = input.ExecutionCreateRequest.GetSpec().GetQualityOfService().GetTier()
	} else if input.LaunchPlan.GetSpec().GetQualityOfService() != nil {
		if input.LaunchPlan.GetSpec().GetQualityOfService().GetSpec() != nil {
			logger.Debugf(ctx, "Determining quality of service from launch plan spec for [%s/%s/%s]",
				input.ExecutionCreateRequest.GetProject(), input.ExecutionCreateRequest.GetDomain(),
				input.ExecutionCreateRequest.GetName())
			duration, err := ptypes.Duration(input.LaunchPlan.GetSpec().GetQualityOfService().GetSpec().GetQueueingBudget())
			if err != nil {
				return QualityOfServiceSpec{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"Invalid custom quality of service set in launch plan [%v], failed to parse duration [%v] with: %v",
					input.LaunchPlan.GetId(),
					input.ExecutionCreateRequest.GetSpec().GetQualityOfService().GetSpec().GetQueueingBudget(), err)
			}
			return QualityOfServiceSpec{
				QueuingBudget: duration,
			}, nil
		}
		qualityOfServiceTier = input.LaunchPlan.GetSpec().GetQualityOfService().GetTier()
	} else if input.Workflow.GetClosure().GetCompiledWorkflow().GetPrimary().GetTemplate().GetMetadata() != nil &&
		input.Workflow.GetClosure().GetCompiledWorkflow().GetPrimary().GetTemplate().GetMetadata().GetQualityOfService() != nil {
		logger.Debugf(ctx, "Determining quality of service from workflow spec for [%s/%s/%s]",
			input.ExecutionCreateRequest.GetProject(), input.ExecutionCreateRequest.GetDomain(),
			input.ExecutionCreateRequest.GetName())
		if input.Workflow.GetClosure().GetCompiledWorkflow().GetPrimary().GetTemplate().GetMetadata().GetQualityOfService().GetSpec() != nil {
			duration, err := ptypes.Duration(input.Workflow.GetClosure().GetCompiledWorkflow().GetPrimary().GetTemplate().GetMetadata().GetQualityOfService().GetSpec().GetQueueingBudget())
			if err != nil {
				return QualityOfServiceSpec{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"Invalid custom quality of service set in workflow [%v], failed to parse duration [%v] with: %v",
					workflowIdentifier,
					input.ExecutionCreateRequest.GetSpec().GetQualityOfService().GetSpec().GetQueueingBudget(), err)
			}
			return QualityOfServiceSpec{
				QueuingBudget: duration,
			}, nil
		}
		qualityOfServiceTier = input.Workflow.GetClosure().GetCompiledWorkflow().GetPrimary().GetTemplate().GetMetadata().GetQualityOfService().GetTier()
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
				input.ExecutionCreateRequest.GetProject(), input.ExecutionCreateRequest.GetDomain(),
				input.ExecutionCreateRequest.GetName())
			duration, err := ptypes.Duration(qualityOfService.GetSpec().GetQueueingBudget())
			if err != nil {
				return QualityOfServiceSpec{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
					"Invalid custom quality of service set in overridable matching attributes for [%v],"+
						"failed to parse duration [%v] with: %v", workflowIdentifier,
					input.ExecutionCreateRequest.GetSpec().GetQualityOfService().GetSpec().GetQueueingBudget(), err)
			}
			return QualityOfServiceSpec{
				QueuingBudget: duration,
			}, nil
		} else if qualityOfService != nil && qualityOfService.GetTier() != core.QualityOfService_UNDEFINED {
			logger.Debugf(ctx, "Determining quality of service tier from database override for [%s/%s/%s]",
				input.ExecutionCreateRequest.GetProject(), input.ExecutionCreateRequest.GetDomain(),
				input.ExecutionCreateRequest.GetName())
			qualityOfServiceTier = input.Workflow.GetClosure().GetCompiledWorkflow().GetPrimary().GetTemplate().GetMetadata().GetQualityOfService().GetTier()
		}
	}

	// If we've come all this way and at no layer is an overridable configuration for the quality of service tier
	// set, use the default values from the admin application config.
	if qualityOfServiceTier == core.QualityOfService_UNDEFINED {
		logger.Debugf(ctx, "Determining quality of service tier from application config override for [%s/%s/%s]",
			input.ExecutionCreateRequest.GetProject(), input.ExecutionCreateRequest.GetDomain(),
			input.ExecutionCreateRequest.GetName())
		var ok bool
		qualityOfServiceTier, ok = q.config.QualityOfServiceConfiguration().GetDefaultTiers()[input.ExecutionCreateRequest.GetDomain()]
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
		input.ExecutionCreateRequest.GetProject(), input.ExecutionCreateRequest.GetDomain(),
		input.ExecutionCreateRequest.GetName(), qualityOfServiceTier)
	// Config values should always be vetted so there's no need to check the error from conversion.
	duration, _ := ptypes.Duration(executionValues.GetQueueingBudget())

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
