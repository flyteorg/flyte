package transformers

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

// Transforms a WorkflowCreateRequest to a workflow model
func CreateWorkflowModel(request *admin.WorkflowCreateRequest, remoteClosureIdentifier string,
	digest []byte) (models.Workflow, error) {
	var typedInterface []byte
	if request.Spec != nil && request.Spec.Template != nil && request.Spec.Template.Interface != nil {
		serializedTypedInterface, err := proto.Marshal(request.Spec.Template.Interface)
		if err != nil {
			return models.Workflow{}, errors.NewFlyteAdminError(codes.Internal, "Failed to serialize workflow spec")
		}
		typedInterface = serializedTypedInterface
	}
	return models.Workflow{
		WorkflowKey: models.WorkflowKey{
			Project: request.Id.Project,
			Domain:  request.Id.Domain,
			Name:    request.Id.Name,
			Version: request.Id.Version,
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
	createdAt, err := ptypes.TimestampProto(workflowModel.CreatedAt)
	if err != nil {
		return admin.Workflow{}, errors.NewFlyteAdminErrorf(codes.Internal, "failed to read created at timestamp")
	}

	var workflowInterface core.TypedInterface
	if len(workflowModel.TypedInterface) > 0 {
		err = proto.Unmarshal(workflowModel.TypedInterface, &workflowInterface)
		if err != nil {
			return admin.Workflow{}, errors.NewFlyteAdminErrorf(codes.Internal, fmt.Sprintf("failed to unmarshal workflow %v interface. Error message: %v", workflowModel.ID, err.Error()))
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
