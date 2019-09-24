package notifications

import (
	"fmt"

	"strings"

	"github.com/lyft/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
)

const executionError = " The execution failed with error: [%s]."

const substitutionParam = "{{ %s }}"
const project = "project"
const domain = "domain"
const name = "name"
const phase = "phase"
const errorPlaceholder = "error"
const replaceAllInstances = -1

func substituteEmailParameters(message string, request admin.WorkflowExecutionEventRequest, execution models.Execution) string {
	response := strings.Replace(message, fmt.Sprintf(substitutionParam, project), execution.Project, replaceAllInstances)
	response = strings.Replace(response, fmt.Sprintf(substitutionParam, domain), execution.Domain, replaceAllInstances)
	response = strings.Replace(response, fmt.Sprintf(substitutionParam, name), execution.Name, replaceAllInstances)
	response = strings.Replace(response, fmt.Sprintf(substitutionParam, phase),
		strings.ToLower(request.Event.Phase.String()), replaceAllInstances)
	if request.Event.GetError() != nil {
		response = strings.Replace(response, fmt.Sprintf(substitutionParam, errorPlaceholder),
			fmt.Sprintf(executionError, request.Event.GetError().Message), replaceAllInstances)
	} else {
		// Replace the optional error placeholder with an empty string.
		response = strings.Replace(response, fmt.Sprintf(substitutionParam, errorPlaceholder), "", replaceAllInstances)
	}

	return response
}

// Converts a terminal execution event and existing execution model to an admin.EmailMessage proto, substituting parameters
// in customizable email fields set in the flyteadmin application notifications config.
func ToEmailMessageFromWorkflowExecutionEvent(
	config runtimeInterfaces.NotificationsConfig,
	emailNotification admin.EmailNotification,
	request admin.WorkflowExecutionEventRequest,
	execution models.Execution) *admin.EmailMessage {

	return &admin.EmailMessage{
		SubjectLine:     substituteEmailParameters(config.NotificationsEmailerConfig.Subject, request, execution),
		SenderEmail:     config.NotificationsEmailerConfig.Sender,
		RecipientsEmail: emailNotification.GetRecipientsEmail(),
		Body:            substituteEmailParameters(config.NotificationsEmailerConfig.Body, request, execution),
	}
}
