package validation

import (
	"context"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	runtime "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

const numSystemNodes = 2 // A workflow graph always has a start and end node injected by the platform.

func ValidateWorkflow(
	ctx context.Context, request *admin.WorkflowCreateRequest, db repositoryInterfaces.Repository,
	config runtime.ApplicationConfiguration) error {
	if err := ValidateIdentifier(request.Id, common.Workflow); err != nil {
		return err
	}
	if err := ValidateProjectAndDomain(ctx, db, config, request.Id.Project, request.Id.Domain); err != nil {
		return err
	}
	if request.Spec == nil || request.Spec.Template == nil {
		return shared.GetMissingArgumentError(shared.Spec)
	}
	return nil
}

func ValidateCompiledWorkflow(identifier *core.Identifier, workflow *admin.WorkflowClosure, config runtime.RegistrationValidationConfiguration) error {
	if len(config.GetWorkflowSizeLimit()) > 0 {
		workflowSizeLimit := resource.MustParse(config.GetWorkflowSizeLimit())
		workflowSizeValue := resource.NewQuantity(int64(proto.Size(workflow)), resource.DecimalExponent)
		if workflowSizeLimit.Cmp(*workflowSizeValue) <= -1 {
			return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
				"Workflow closure size exceeds max limit [%v]", config.GetWorkflowSizeLimit())
		}
	}
	if config.GetWorkflowNodeLimit() == 0 {
		// Treat this is unset. There is no limit to compare against.
		return nil
	}
	if workflow.CompiledWorkflow == nil || workflow.CompiledWorkflow.Primary == nil ||
		workflow.CompiledWorkflow.Primary.Template == nil || workflow.CompiledWorkflow.Primary.Template.Nodes == nil {
		logger.Warningf(context.Background(), "workflow [%+v] did not have any primary nodes", identifier)
		return nil
	}
	numUserNodes := len(workflow.CompiledWorkflow.Primary.Template.Nodes) - numSystemNodes
	if numUserNodes > config.GetWorkflowNodeLimit() {
		return errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"number of nodes in workflow [%+v] exceeds limit (%v > %v)", identifier,
			numUserNodes, config.GetWorkflowNodeLimit())
	}
	return nil
}
