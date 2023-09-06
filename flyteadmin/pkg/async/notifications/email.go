package notifications

import (
	"fmt"

	"strings"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

type GetTemplateValue func(admin.WorkflowExecutionEventRequest, *admin.Execution) string

const executionError = " The execution failed with error: [%s]."

const substitutionParam = "{{ %s }}"
const substitutionParamNoSpaces = "{{%s}}"
const project = "project"
const domain = "domain"
const name = "name"
const phase = "phase"
const errorPlaceholder = "error"
const workflowProject = "workflow.project"
const workflowDomain = "workflow.domain"
const workflowName = "workflow.name"
const workflowVersion = "workflow.version"
const launchPlanProject = "launch_plan.project"
const launchPlanDomain = "launch_plan.domain"
const launchPlanName = "launch_plan.name"
const launchPlanVersion = "launch_plan.version"
const replaceAllInstances = -1

func getProject(_ admin.WorkflowExecutionEventRequest, exec *admin.Execution) string {
	return exec.Id.Project
}

func getDomain(_ admin.WorkflowExecutionEventRequest, exec *admin.Execution) string {
	return exec.Id.Domain
}

func getName(_ admin.WorkflowExecutionEventRequest, exec *admin.Execution) string {
	return exec.Id.Name
}

func getPhase(request admin.WorkflowExecutionEventRequest, _ *admin.Execution) string {
	return strings.ToLower(request.Event.Phase.String())
}

func getError(request admin.WorkflowExecutionEventRequest, _ *admin.Execution) string {
	if request.Event.GetError() != nil {
		return fmt.Sprintf(executionError, request.Event.GetError().Message)
	}
	return ""
}

func getWorkflowProject(_ admin.WorkflowExecutionEventRequest, exec *admin.Execution) string {
	return exec.Closure.WorkflowId.Project
}

func getWorkflowDomain(_ admin.WorkflowExecutionEventRequest, exec *admin.Execution) string {
	return exec.Closure.WorkflowId.Domain
}

func getWorkflowName(_ admin.WorkflowExecutionEventRequest, exec *admin.Execution) string {
	return exec.Closure.WorkflowId.Name
}

func getWorkflowVersion(_ admin.WorkflowExecutionEventRequest, exec *admin.Execution) string {
	return exec.Closure.WorkflowId.Version
}

func getLaunchPlanProject(_ admin.WorkflowExecutionEventRequest, exec *admin.Execution) string {
	return exec.Spec.LaunchPlan.Project
}

func getLaunchPlanDomain(_ admin.WorkflowExecutionEventRequest, exec *admin.Execution) string {
	return exec.Spec.LaunchPlan.Domain
}

func getLaunchPlanName(_ admin.WorkflowExecutionEventRequest, exec *admin.Execution) string {
	return exec.Spec.LaunchPlan.Name
}

func getLaunchPlanVersion(_ admin.WorkflowExecutionEventRequest, exec *admin.Execution) string {
	return exec.Spec.LaunchPlan.Version
}

var getTemplateValueFuncs = map[string]GetTemplateValue{
	project:           getProject,
	domain:            getDomain,
	name:              getName,
	phase:             getPhase,
	errorPlaceholder:  getError,
	workflowProject:   getWorkflowProject,
	workflowDomain:    getWorkflowDomain,
	workflowName:      getWorkflowName,
	workflowVersion:   getWorkflowVersion,
	launchPlanProject: getLaunchPlanProject,
	launchPlanDomain:  getLaunchPlanDomain,
	launchPlanName:    getLaunchPlanName,
	launchPlanVersion: getLaunchPlanVersion,
}

func substituteEmailParameters(message string, request admin.WorkflowExecutionEventRequest, execution *admin.Execution) string {
	for template, function := range getTemplateValueFuncs {
		message = strings.Replace(message, fmt.Sprintf(substitutionParam, template), function(request, execution), replaceAllInstances)
		message = strings.Replace(message, fmt.Sprintf(substitutionParamNoSpaces, template), function(request, execution), replaceAllInstances)
	}
	return message
}

// Converts a terminal execution event and existing execution model to an admin.EmailMessage proto, substituting parameters
// in customizable email fields set in the flyteadmin application notifications config.
func ToEmailMessageFromWorkflowExecutionEvent(
	config runtimeInterfaces.NotificationsConfig,
	emailNotification admin.EmailNotification,
	request admin.WorkflowExecutionEventRequest,
	execution *admin.Execution) *admin.EmailMessage {

	return &admin.EmailMessage{
		SubjectLine:     substituteEmailParameters(config.NotificationsEmailerConfig.Subject, request, execution),
		SenderEmail:     config.NotificationsEmailerConfig.Sender,
		RecipientsEmail: emailNotification.GetRecipientsEmail(),
		Body:            substituteEmailParameters(config.NotificationsEmailerConfig.Body, request, execution),
	}
}
