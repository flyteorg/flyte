package notifications

import (
	"fmt"
	"testing"

	"strings"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

const executionProjectValue = "proj"
const executionDomainValue = "prod"
const executionNameValue = "e124"
const launchPlanProjectValue = "lp_proj"
const launchPlanDomainValue = "lp_domain"
const launchPlanNameValue = "lp_name"
const launchPlanVersionValue = "lp_version"
const workflowProjectValue = "wf_proj"
const workflowDomainValue = "wf_domain"
const workflowNameValue = "wf_name"
const workflowVersionValue = "wf_version"

var workflowExecution = &admin.Execution{
	Id: &core.WorkflowExecutionIdentifier{
		Project: executionProjectValue,
		Domain:  executionDomainValue,
		Name:    executionNameValue,
	},
	Spec: &admin.ExecutionSpec{
		LaunchPlan: &core.Identifier{
			Project: launchPlanProjectValue,
			Domain:  launchPlanDomainValue,
			Name:    launchPlanNameValue,
			Version: launchPlanVersionValue,
		},
	},
	Closure: &admin.ExecutionClosure{
		WorkflowId: &core.Identifier{
			Project: workflowProjectValue,
			Domain:  workflowDomainValue,
			Name:    workflowNameValue,
			Version: workflowVersionValue,
		},
		Phase: core.WorkflowExecution_SUCCEEDED,
	},
}

func TestSubstituteEmailParameters(t *testing.T) {
	message := "{{ unused }}. {{project }} and {{ domain }} and {{ name }} ended up in {{ phase }}.{{ error }}"
	request := admin.WorkflowExecutionEventRequest{
		Event: &event.WorkflowExecutionEvent{
			Phase: core.WorkflowExecution_SUCCEEDED,
		},
	}
	assert.Equal(t, "{{ unused }}. {{project }} and prod and e124 ended up in succeeded.",
		substituteEmailParameters(message, request, workflowExecution))
	request.Event.OutputResult = &event.WorkflowExecutionEvent_Error{
		Error: &core.ExecutionError{
			Message: "uh-oh",
		},
	}
	assert.Equal(t, "{{ unused }}. {{project }} and prod and e124 ended up in succeeded. The execution failed with error: [uh-oh].",
		substituteEmailParameters(message, request, workflowExecution))
}

func TestSubstituteAllTemplates(t *testing.T) {
	templateVars := map[string]string{
		fmt.Sprintf(substitutionParam, project):           executionProjectValue,
		fmt.Sprintf(substitutionParam, domain):            executionDomainValue,
		fmt.Sprintf(substitutionParam, name):              executionNameValue,
		fmt.Sprintf(substitutionParam, launchPlanProject): launchPlanProjectValue,
		fmt.Sprintf(substitutionParam, launchPlanDomain):  launchPlanDomainValue,
		fmt.Sprintf(substitutionParam, launchPlanName):    launchPlanNameValue,
		fmt.Sprintf(substitutionParam, launchPlanVersion): launchPlanVersionValue,
		fmt.Sprintf(substitutionParam, workflowProject):   workflowProjectValue,
		fmt.Sprintf(substitutionParam, workflowDomain):    workflowDomainValue,
		fmt.Sprintf(substitutionParam, workflowName):      workflowNameValue,
		fmt.Sprintf(substitutionParam, workflowVersion):   workflowVersionValue,
		fmt.Sprintf(substitutionParam, phase):             strings.ToLower(core.WorkflowExecution_SUCCEEDED.String()),
	}
	var messageTemplate, desiredResult []string
	for template, result := range templateVars {
		messageTemplate = append(messageTemplate, template)
		desiredResult = append(desiredResult, result)
	}
	request := admin.WorkflowExecutionEventRequest{
		Event: &event.WorkflowExecutionEvent{
			Phase: core.WorkflowExecution_SUCCEEDED,
		},
	}
	assert.Equal(t, strings.Join(desiredResult, ","),
		substituteEmailParameters(strings.Join(messageTemplate, ","), request, workflowExecution))
}

func TestSubstituteAllTemplatesNoSpaces(t *testing.T) {
	templateVars := map[string]string{
		fmt.Sprintf(substitutionParamNoSpaces, project):           executionProjectValue,
		fmt.Sprintf(substitutionParamNoSpaces, domain):            executionDomainValue,
		fmt.Sprintf(substitutionParamNoSpaces, name):              executionNameValue,
		fmt.Sprintf(substitutionParamNoSpaces, launchPlanProject): launchPlanProjectValue,
		fmt.Sprintf(substitutionParamNoSpaces, launchPlanDomain):  launchPlanDomainValue,
		fmt.Sprintf(substitutionParamNoSpaces, launchPlanName):    launchPlanNameValue,
		fmt.Sprintf(substitutionParamNoSpaces, launchPlanVersion): launchPlanVersionValue,
		fmt.Sprintf(substitutionParamNoSpaces, workflowProject):   workflowProjectValue,
		fmt.Sprintf(substitutionParamNoSpaces, workflowDomain):    workflowDomainValue,
		fmt.Sprintf(substitutionParamNoSpaces, workflowName):      workflowNameValue,
		fmt.Sprintf(substitutionParamNoSpaces, workflowVersion):   workflowVersionValue,
		fmt.Sprintf(substitutionParamNoSpaces, phase):             strings.ToLower(core.WorkflowExecution_SUCCEEDED.String()),
	}
	var messageTemplate, desiredResult []string
	for template, result := range templateVars {
		messageTemplate = append(messageTemplate, template)
		desiredResult = append(desiredResult, result)
	}
	request := admin.WorkflowExecutionEventRequest{
		Event: &event.WorkflowExecutionEvent{
			Phase: core.WorkflowExecution_SUCCEEDED,
		},
	}
	assert.Equal(t, strings.Join(desiredResult, ","),
		substituteEmailParameters(strings.Join(messageTemplate, ","), request, workflowExecution))
}

func TestToEmailMessageFromWorkflowExecutionEvent(t *testing.T) {
	notificationsConfig := runtimeInterfaces.NotificationsConfig{
		NotificationsEmailerConfig: runtimeInterfaces.NotificationsEmailerConfig{
			Body: "Execution \"{{ name }}\" has succeeded in \"{{ domain }}\". View details at " +
				"<a href=\"https://example.com/executions/{{ project }}/{{ domain }}/{{ name }}\">" +
				"https://example.com/executions/{{ project }}/{{ domain }}/{{ name }}</a>.",
			Sender:  "no-reply@example.com",
			Subject: "Notice: Execution \"{{ name }}\" has succeeded in \"{{ domain }}\".",
		},
	}
	emailNotification := admin.EmailNotification{
		RecipientsEmail: []string{
			"a@example.com", "b@example.org",
		},
	}
	request := admin.WorkflowExecutionEventRequest{
		Event: &event.WorkflowExecutionEvent{
			Phase: core.WorkflowExecution_ABORTED,
		},
	}
	emailMessage := ToEmailMessageFromWorkflowExecutionEvent(notificationsConfig, emailNotification, request, workflowExecution)
	assert.True(t, proto.Equal(emailMessage, &admin.EmailMessage{
		RecipientsEmail: []string{
			"a@example.com", "b@example.org",
		},
		SenderEmail: "no-reply@example.com",
		SubjectLine: "Notice: Execution \"e124\" has succeeded in \"prod\".",
		Body: "Execution \"e124\" has succeeded in \"prod\". View details at " +
			"<a href=\"https://example.com/executions/proj/prod/e124\">" +
			"https://example.com/executions/proj/prod/e124</a>.",
	}), fmt.Sprintf("%+v", emailMessage))
}
