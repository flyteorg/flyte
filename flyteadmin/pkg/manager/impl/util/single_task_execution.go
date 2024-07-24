package util

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	repositoryInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

var defaultRetryStrategy = core.RetryStrategy{
	Retries: 3,
}

const systemNamePrefix = ".flytegen.%s"

const noInputNodeID = ""

func generateNodeNameFromTask(taskName string) (string, error) {
	nodeNameBuilder := strings.Builder{}
	for _, i := range taskName {
		if i == '-' || unicode.IsLetter(i) || unicode.IsNumber(i) {
			nodeNameBuilder.WriteRune(i)
		}
	}
	id, err := encoding.FixedLengthUniqueID(nodeNameBuilder.String(), utils.MaxUniqueIDLength)
	if err != nil {
		return "", err
	}
	return id, nil
}

func generateWorkflowNameFromTask(taskName string) string {
	return fmt.Sprintf(systemNamePrefix, taskName)
}

func generateBindings(outputs core.VariableMap, nodeID string) []*core.Binding {
	bindings := make([]*core.Binding, 0, len(outputs.Variables))
	for key := range outputs.Variables {
		binding := &core.Binding{
			Var: key,
			Binding: &core.BindingData{
				Value: &core.BindingData_Promise{
					Promise: &core.OutputReference{
						NodeId: nodeID,
						Var:    key,
					},
				},
			},
		}

		bindings = append(bindings, binding)
	}
	return bindings
}

func generateWorkflowNameFromNode(node *core.Node) (string, error) {
	switch node.Target.(type) {
	case *core.Node_TaskNode:
		// TODO need node_id + is_downstream to generate unique workflow name if running downstream nodes
		taskNodeWrapper := node.Target.(*core.Node_TaskNode)
		return generateWorkflowNameFromTask(taskNodeWrapper.TaskNode.Reference.(*core.TaskNode_ReferenceId).ReferenceId.Name), nil
	case *core.Node_ArrayNode:
		arrayNodeWrapper := node.Target.(*core.Node_ArrayNode)
		return generateWorkflowNameFromTask(arrayNodeWrapper.ArrayNode.Node.Id), nil
	}
	return "", errors.NewFlyteAdminErrorf(codes.InvalidArgument, "Only Task and Array Nodes are supported for relaunching")
}

func createAndUpdateWorkflowModel(ctx context.Context, db repositoryInterfaces.Repository,
	workflowManager interfaces.WorkflowInterface, namedEntityManager interfaces.NamedEntityInterface,
	workflowIdentifier core.Identifier, workflowSpec admin.WorkflowSpec) (models.Workflow, error) {
	_, err := workflowManager.CreateWorkflow(ctx, admin.WorkflowCreateRequest{
		Id:   &workflowIdentifier,
		Spec: &workflowSpec,
	})
	if err != nil {
		// In the case of race conditions, if the workflow already exists we can safely ignore the corresponding
		// error.
		if ferr, ok := err.(errors.FlyteAdminError); !ok || ferr.Code() != codes.AlreadyExists {
			return models.Workflow{}, err
		}
	}
	// Now, set the newly created skeleton workflow to 'SYSTEM_GENERATED'.
	_, err = namedEntityManager.UpdateNamedEntity(ctx, admin.NamedEntityUpdateRequest{
		ResourceType: core.ResourceType_WORKFLOW,
		Id: &admin.NamedEntityIdentifier{
			Org:     workflowIdentifier.Org,
			Project: workflowIdentifier.Project,
			Domain:  workflowIdentifier.Domain,
			Name:    workflowIdentifier.Name,
		},
		Metadata: &admin.NamedEntityMetadata{State: admin.NamedEntityState_SYSTEM_GENERATED},
	})
	if err != nil {
		logger.Warningf(ctx, "Failed to set skeleton workflow state to system-generated: %v", err)
		return models.Workflow{}, err
	}
	workflowModel, err := db.WorkflowRepo().Get(ctx, repositoryInterfaces.Identifier{
		Org:     workflowIdentifier.Org,
		Project: workflowIdentifier.Project,
		Domain:  workflowIdentifier.Domain,
		Name:    workflowIdentifier.Name,
		Version: workflowIdentifier.Version,
	})
	if err != nil {
		// This is unexpected - at this point we've successfully just created the skeleton workflow.
		logger.Warningf(ctx, "Failed to fetch newly created workflow model from db store: %v", err)
		return models.Workflow{}, err
	}

	return workflowModel, err
}

func CreateOrGetWorkflowModel(
	ctx context.Context, request admin.ExecutionCreateRequest, db repositoryInterfaces.Repository,
	workflowManager interfaces.WorkflowInterface, namedEntityManager interfaces.NamedEntityInterface, taskIdentifier *core.Identifier,
	task *admin.Task) (*models.Workflow, error) {
	workflowIdentifier := core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Org:          taskIdentifier.Org,
		Project:      taskIdentifier.Project,
		Domain:       taskIdentifier.Domain,
		Name:         generateWorkflowNameFromTask(taskIdentifier.Name),
		Version:      taskIdentifier.Version,
	}
	workflowModel, err := db.WorkflowRepo().Get(ctx, repositoryInterfaces.Identifier{
		Org:     workflowIdentifier.Org,
		Project: workflowIdentifier.Project,
		Domain:  workflowIdentifier.Domain,
		Name:    workflowIdentifier.Name,
		Version: workflowIdentifier.Version,
	})

	var retryStrategy *core.RetryStrategy
	if task.GetClosure().GetCompiledTask().GetTemplate().GetMetadata().GetRetries() != nil {
		retryStrategy = task.GetClosure().GetCompiledTask().GetTemplate().GetMetadata().GetRetries()
	} else {
		retryStrategy = &defaultRetryStrategy
	}

	if err != nil {
		if ferr, ok := err.(errors.FlyteAdminError); !ok || ferr.Code() != codes.NotFound {
			return nil, err
		}
		// If we got this far, there is no existing workflow. Create a skeleton one now.
		nodeID, err := generateNodeNameFromTask(taskIdentifier.Name)
		if err != nil {
			return nil, err
		}
		workflowSpec := admin.WorkflowSpec{
			Template: &core.WorkflowTemplate{
				Id:        &workflowIdentifier,
				Interface: task.Closure.CompiledTask.Template.Interface,
				Nodes: []*core.Node{
					{
						Id: nodeID,
						Metadata: &core.NodeMetadata{
							Name:    nodeID,
							Retries: retryStrategy,
						},
						Inputs: generateBindings(*task.Closure.CompiledTask.Template.Interface.Inputs, noInputNodeID),
						Target: &core.Node_TaskNode{
							TaskNode: &core.TaskNode{
								Reference: &core.TaskNode_ReferenceId{
									ReferenceId: taskIdentifier,
								},
							},
						},
					},
				},

				Outputs: generateBindings(*task.Closure.CompiledTask.Template.Interface.Outputs, nodeID),
			},
		}

		workflowModel, err = createAndUpdateWorkflowModel(ctx, db, workflowManager, namedEntityManager, workflowIdentifier, workflowSpec)
		if err != nil {
			return nil, err
		}
	}

	return &workflowModel, nil
}

func CreateOrGetLaunchPlan(ctx context.Context,
	db repositoryInterfaces.Repository, config runtimeInterfaces.Configuration, namedEntityManager interfaces.NamedEntityInterface, identifier *core.Identifier,
	workflowInterface *core.TypedInterface, workflowID uint, authRole *admin.AuthRole, securityContext *core.SecurityContext) (*admin.LaunchPlan, error) {
	var launchPlan *admin.LaunchPlan
	var err error
	launchPlanIdentifier := core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Name:         identifier.Name,
		Org:          identifier.Org,
		Project:      identifier.Project,
		Domain:       identifier.Domain,
		Version:      identifier.Version,
	}
	if identifier.GetResourceType() == core.ResourceType_TASK {
		launchPlanIdentifier.Name = generateWorkflowNameFromTask(identifier.Name)
	}
	launchPlan, err = GetLaunchPlan(ctx, db, launchPlanIdentifier)
	if err != nil {
		if ferr, ok := err.(errors.FlyteAdminError); !ok || ferr.Code() != codes.NotFound {
			return nil, err
		}

		// Create launch plan.
		generatedCreateLaunchPlanReq := admin.LaunchPlanCreateRequest{
			Id: &launchPlanIdentifier,
			Spec: &admin.LaunchPlanSpec{
				WorkflowId: &core.Identifier{
					ResourceType: core.ResourceType_WORKFLOW,
					Org:          identifier.Org,
					Project:      identifier.Project,
					Domain:       identifier.Domain,
					Name:         identifier.Name,
					Version:      identifier.Version,
				},
				EntityMetadata:  &admin.LaunchPlanMetadata{},
				DefaultInputs:   &core.ParameterMap{},
				FixedInputs:     &core.LiteralMap{},
				Labels:          &admin.Labels{},
				Annotations:     &admin.Annotations{},
				AuthRole:        authRole,
				SecurityContext: securityContext,
			},
		}
		if err := validation.ValidateLaunchPlan(ctx, generatedCreateLaunchPlanReq, db, config.ApplicationConfiguration(), workflowInterface); err != nil {
			logger.Debugf(ctx, "could not create launch plan: %+v, request failed validation with err: %v", identifier, err)
			return nil, err
		}
		transformedLaunchPlan := transformers.CreateLaunchPlan(generatedCreateLaunchPlanReq, workflowInterface.Outputs)
		launchPlan = &transformedLaunchPlan
		launchPlanDigest, err := GetLaunchPlanDigest(ctx, launchPlan)
		if err != nil {
			logger.Errorf(ctx, "failed to compute launch plan digest for [%+v] with err: %v", launchPlan.Id, err)
			return nil, err
		}
		launchPlanModel, err :=
			transformers.CreateLaunchPlanModel(*launchPlan, workflowID, launchPlanDigest, admin.LaunchPlanState_INACTIVE)
		if err != nil {
			logger.Errorf(ctx,
				"Failed to transform launch plan model [%+v], and workflow outputs [%+v] with err: %v",
				identifier, workflowInterface.Outputs, err)
			return nil, err
		}
		err = db.LaunchPlanRepo().Create(ctx, launchPlanModel)
		if err != nil {
			logger.Errorf(ctx, "Failed to save launch plan model [%+v] with err: %v", launchPlanIdentifier, err)
			return nil, err
		}
		_, err = namedEntityManager.UpdateNamedEntity(ctx, admin.NamedEntityUpdateRequest{
			ResourceType: core.ResourceType_LAUNCH_PLAN,
			Id: &admin.NamedEntityIdentifier{
				Project: launchPlan.GetId().GetProject(),
				Domain:  launchPlan.GetId().GetDomain(),
				Name:    launchPlan.GetId().GetName(),
			},
			Metadata: &admin.NamedEntityMetadata{State: admin.NamedEntityState_SYSTEM_GENERATED},
		})
		if err != nil {
			logger.Warningf(ctx, "Failed to set launch plan state to system-generated: %v", err)
			return nil, err
		}
	}

	return launchPlan, nil
}

func CreateOrGetWorkflowFromNode(
	ctx context.Context, subNode *core.Node, db repositoryInterfaces.Repository,
	workflowManager interfaces.WorkflowInterface, namedEntityManager interfaces.NamedEntityInterface,
	identifier *core.Identifier, name string) (*models.Workflow, error) {

	if name == "" {
		var err error
		name, err = generateWorkflowNameFromNode(subNode)
		if err != nil {
			return nil, err
		}
	} else {
		name = fmt.Sprintf(systemNamePrefix, name)
	}

	workflowIdentifier := &core.Identifier{
		Name:         name,
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      identifier.Project,
		Domain:       identifier.Domain,
		Version:      identifier.Version,
		Org:          identifier.Org,
	}

	workflowModel, err := db.WorkflowRepo().Get(ctx, repositoryInterfaces.Identifier{
		Name:    workflowIdentifier.Name,
		Project: workflowIdentifier.Project,
		Domain:  workflowIdentifier.Domain,
		Version: workflowIdentifier.Version,
		Org:     workflowIdentifier.Org,
	})

	if err != nil {
		if ferr, ok := err.(errors.FlyteAdminError); !ok || ferr.Code() != codes.NotFound {
			return nil, err
		}

		var wfInterface *core.TypedInterface
		var wfOutputs []*core.Binding
		subNode.UpstreamNodeIds = []string{}
		switch subNode.Target.(type) {
		case *core.Node_TaskNode:
			taskNodeWrapper := subNode.Target.(*core.Node_TaskNode)
			taskNodeReference := taskNodeWrapper.TaskNode.Reference.(*core.TaskNode_ReferenceId)
			task, err := GetTask(ctx, db, *taskNodeReference.ReferenceId)
			if err != nil {
				return nil, err
			}
			wfInterface = task.Closure.CompiledTask.Template.Interface
			wfOutputs = generateBindings(*task.Closure.CompiledTask.Template.Interface.Outputs, subNode.GetId())
			subNode.Inputs = generateBindings(*task.Closure.CompiledTask.Template.Interface.Inputs, noInputNodeID)
		case *core.Node_ArrayNode:
			arrayNodeWrapper := subNode.Target.(*core.Node_ArrayNode)
			arrayNode := arrayNodeWrapper.ArrayNode.Node
			switch arrayNode.Target.(type) {
			case *core.Node_TaskNode:
				taskNodeWrapper := arrayNode.Target.(*core.Node_TaskNode)
				taskNodeReference := taskNodeWrapper.TaskNode.Reference.(*core.TaskNode_ReferenceId)
				task, err := GetTask(ctx, db, *taskNodeReference.ReferenceId)
				if err != nil {
					return nil, err
				}
				taskInterface := task.Closure.CompiledTask.Template.Interface
				wfInterface = taskInterface
				wfOutputs = generateBindings(*taskInterface.Outputs, subNode.GetId())
				subNode.Inputs = generateBindings(*task.Closure.CompiledTask.Template.Interface.Inputs, noInputNodeID)
			case *core.Node_WorkflowNode:
				// TODO - pvditt implement as part of mapping over launch plan PRs
				return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "Only Task and Workflow Nodes are supported for Array Nodes")
			default:
				return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "Only Task and Workflow Nodes are supported for Array Nodes")
			}
		default:
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "Only Task and Array Nodes are supported for relaunching")
		}

		workflowSpec := admin.WorkflowSpec{
			Template: &core.WorkflowTemplate{
				Id:        workflowIdentifier,
				Interface: wfInterface,
				Nodes:     []*core.Node{subNode},
				Outputs:   wfOutputs,
			},
		}
		workflowModel, err = createAndUpdateWorkflowModel(ctx, db, workflowManager, namedEntityManager, *workflowIdentifier, workflowSpec)
		if err != nil {
			return nil, err
		}
	}

	return &workflowModel, nil
}
