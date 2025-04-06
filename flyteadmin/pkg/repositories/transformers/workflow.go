package transformers

import (
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

// Transforms a WorkflowCreateRequest to a workflow model
func CreateWorkflowModel(request *admin.WorkflowCreateRequest, remoteClosureIdentifier string,
	digest []byte) (models.Workflow, error) {
	var typedInterface []byte
	if request.GetSpec() != nil && request.GetSpec().GetTemplate() != nil && request.GetSpec().GetTemplate().GetInterface() != nil {
		serializedTypedInterface, err := proto.Marshal(request.GetSpec().GetTemplate().GetInterface())
		if err != nil {
			return models.Workflow{}, errors.NewFlyteAdminError(codes.Internal, "Failed to serialize workflow spec")
		}
		typedInterface = serializedTypedInterface
	}
	return models.Workflow{
		WorkflowKey: models.WorkflowKey{
			Project: request.GetId().GetProject(),
			Domain:  request.GetId().GetDomain(),
			Name:    request.GetId().GetName(),
			Version: request.GetId().GetVersion(),
		},
		TypedInterface:          typedInterface,
		RemoteClosureIdentifier: remoteClosureIdentifier,
		Digest:                  digest,
	}, nil
}

func FromWorkflowModel(workflowModel models.Workflow) (admin.Workflow, error) {
	id := core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      workflowModel.WorkflowKey.Project,
		Domain:       workflowModel.WorkflowKey.Domain,
		Name:         workflowModel.WorkflowKey.Name,
		Version:      workflowModel.WorkflowKey.Version,
	}
	createdAt := timestamppb.New(workflowModel.CreatedAt)

	var workflowInterface core.TypedInterface
	if len(workflowModel.TypedInterface) > 0 {
		err := proto.Unmarshal(workflowModel.TypedInterface, &workflowInterface)
		if err != nil {
			return admin.Workflow{}, errors.NewFlyteAdminErrorf(codes.Internal, fmt.Sprintf("failed to unmarshal workflow %v interface. Error message: %v", workflowModel.ID, err.Error())) //nolint
		}
	}

	// Because the spec if offloaded, it is not populated in the model returned here.
	return admin.Workflow{
		Id: &id,
		Closure: &admin.WorkflowClosure{
			CreatedAt: createdAt,
			CompiledWorkflow: &core.CompiledWorkflowClosure{
				Primary: &core.CompiledWorkflow{
					Template: &core.WorkflowTemplate{
						Interface: &workflowInterface,
					},
				},
			},
		},
		ShortDescription: workflowModel.ShortDescription,
	}, nil
}

func FromWorkflowModels(workflowModels []models.Workflow) ([]*admin.Workflow, error) {
	workflows := make([]*admin.Workflow, len(workflowModels))
	for idx, workflowModel := range workflowModels {
		workflow, err := FromWorkflowModel(workflowModel)
		if err != nil {
			return nil, err
		}
		workflows[idx] = &workflow
	}
	return workflows, nil
}

func FromWorkflowModelsToIdentifiers(workflowModels []models.Workflow) []*admin.NamedEntityIdentifier {
	ids := make([]*admin.NamedEntityIdentifier, len(workflowModels))
	for i, wf := range workflowModels {
		ids[i] = &admin.NamedEntityIdentifier{
			Project: wf.Project,
			Domain:  wf.Domain,
			Name:    wf.Name,
		}
	}

	return ids
}
