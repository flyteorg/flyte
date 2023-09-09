package util

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	repositoryInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"google.golang.org/grpc/codes"
)

const maxNodeIDLength = 63

var defaultRetryStrategy = core.RetryStrategy{
	Retries: 3,
}

const systemNamePrefix = ".flytegen.%s"

const noInputNodeID = ""

func generateNodeNameFromTask(taskName string) string {
	if len(taskName) >= maxNodeIDLength {
		taskName = taskName[len(taskName)-maxNodeIDLength:]
	}
	nodeNameBuilder := strings.Builder{}
	for _, i := range taskName {
		if i == '-' || unicode.IsLetter(i) || unicode.IsNumber(i) {
			nodeNameBuilder.WriteRune(i)
		}
	}
	return nodeNameBuilder.String()
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

func CreateOrGetWorkflowModel(
	ctx context.Context, request admin.ExecutionCreateRequest, db repositoryInterfaces.Repository,
	workflowManager interfaces.WorkflowInterface, namedEntityManager interfaces.NamedEntityInterface, taskIdentifier *core.Identifier,
	task *admin.Task) (*models.Workflow, error) {
	workflowIdentifier := core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      taskIdentifier.Project,
		Domain:       taskIdentifier.Domain,
		Name:         generateWorkflowNameFromTask(taskIdentifier.Name),
		Version:      taskIdentifier.Version,
	}
	workflowModel, err := db.WorkflowRepo().Get(ctx, repositoryInterfaces.Identifier{
		Project: workflowIdentifier.Project,
		Domain:  workflowIdentifier.Domain,
		Name:    workflowIdentifier.Name,
		Version: workflowIdentifier.Version,
	})

	if err != nil {
		if ferr, ok := err.(errors.FlyteAdminError); !ok || ferr.Code() != codes.NotFound {
			return nil, err
		}
		// If we got this far, there is no existing workflow. Create a skeleton one now.
		workflowSpec := admin.WorkflowSpec{
			Template: &core.WorkflowTemplate{
				Id:        &workflowIdentifier,
				Interface: task.Closure.CompiledTask.Template.Interface,
				Nodes: []*core.Node{
					{
						Id: generateNodeNameFromTask(taskIdentifier.Name),
						Metadata: &core.NodeMetadata{
							Name:    generateNodeNameFromTask(taskIdentifier.Name),
							Retries: &defaultRetryStrategy,
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

				Outputs: generateBindings(*task.Closure.CompiledTask.Template.Interface.Outputs, generateNodeNameFromTask(taskIdentifier.Name)),
			},
		}

		_, err = workflowManager.CreateWorkflow(ctx, admin.WorkflowCreateRequest{
			Id:   &workflowIdentifier,
			Spec: &workflowSpec,
		})
		if err != nil {
			// In the case of race conditions, if the workflow already exists we can safely ignore the corresponding
			// error.
			if ferr, ok := err.(errors.FlyteAdminError); !ok || ferr.Code() != codes.AlreadyExists {
				return nil, err
			}
		}
		// Now, set the newly created skeleton workflow to 'SYSTEM_GENERATED'.
		_, err = namedEntityManager.UpdateNamedEntity(ctx, admin.NamedEntityUpdateRequest{
			ResourceType: core.ResourceType_WORKFLOW,
			Id: &admin.NamedEntityIdentifier{
				Project: workflowIdentifier.Project,
				Domain:  workflowIdentifier.Domain,
				Name:    workflowIdentifier.Name,
			},
			Metadata: &admin.NamedEntityMetadata{State: admin.NamedEntityState_SYSTEM_GENERATED},
		})
		if err != nil {
			logger.Warningf(ctx, "Failed to set skeleton workflow state to system-generated: %v", err)
			return nil, err
		}
		workflowModel, err = db.WorkflowRepo().Get(ctx, repositoryInterfaces.Identifier{
			Project: workflowIdentifier.Project,
			Domain:  workflowIdentifier.Domain,
			Name:    workflowIdentifier.Name,
			Version: workflowIdentifier.Version,
		})
		if err != nil {
			// This is unexpected - at this point we've successfully just created the skeleton workflow.
			logger.Warningf(ctx, "Failed to fetch newly created workflow model from db store: %v", err)
			return nil, err
		}
	}

	return &workflowModel, nil
}

func CreateOrGetLaunchPlan(ctx context.Context,
	db repositoryInterfaces.Repository, config runtimeInterfaces.Configuration, taskIdentifier *core.Identifier,
	workflowInterface *core.TypedInterface, workflowID uint, spec *admin.ExecutionSpec) (*admin.LaunchPlan, error) {
	var launchPlan *admin.LaunchPlan
	var err error
	launchPlanIdentifier := core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      taskIdentifier.Project,
		Domain:       taskIdentifier.Domain,
		Name:         generateWorkflowNameFromTask(taskIdentifier.Name),
		Version:      taskIdentifier.Version,
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
					Project:      taskIdentifier.Project,
					Domain:       taskIdentifier.Domain,
					Name:         taskIdentifier.Name,
					Version:      taskIdentifier.Version,
				},
				EntityMetadata:  &admin.LaunchPlanMetadata{},
				DefaultInputs:   &core.ParameterMap{},
				FixedInputs:     &core.LiteralMap{},
				Labels:          &admin.Labels{},
				Annotations:     &admin.Annotations{},
				AuthRole:        spec.AuthRole,
				SecurityContext: spec.SecurityContext,
			},
		}
		if err := validation.ValidateLaunchPlan(ctx, generatedCreateLaunchPlanReq, db, config.ApplicationConfiguration(), workflowInterface); err != nil {
			logger.Debugf(ctx, "could not create launch plan: %+v, request failed validation with err: %v", taskIdentifier, err)
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
				taskIdentifier, workflowInterface.Outputs, err)
			return nil, err
		}
		err = db.LaunchPlanRepo().Create(ctx, launchPlanModel)
		if err != nil {
			logger.Errorf(ctx, "Failed to save launch plan model [%+v] with err: %v", launchPlanIdentifier, err)
			return nil, err
		}
	}

	return launchPlan, nil
}
