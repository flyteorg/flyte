package notifications

import (
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/stretchr/testify/assert"
)

func TestSubstituteEmailParameters(t *testing.T) {
	message := "{{ unused }}. {{project }} and {{ domain }} and {{ name }} ended up in {{ phase }}.{{ error }}"
	request := admin.WorkflowExecutionEventRequest{
		Event: &event.WorkflowExecutionEvent{
			Phase: core.WorkflowExecution_SUCCEEDED,
		},
	}
	model := models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "proj",
			Domain:  "prod",
			Name:    "e124",
		},
	}
	assert.Equal(t, "{{ unused }}. {{project }} and prod and e124 ended up in succeeded.",
		substituteEmailParameters(message, request, model))
	request.Event.OutputResult = &event.WorkflowExecutionEvent_Error{
		Error: &core.ExecutionError{
			Message: "uh-oh",
		},
	}
	assert.Equal(t, "{{ unused }}. {{project }} and prod and e124 ended up in succeeded. The execution failed with error: [uh-oh].",
		substituteEmailParameters(message, request, model))
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
	model := models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "proj",
			Domain:  "prod",
			Name:    "e124",
		},
	}
	emailMessage := ToEmailMessageFromWorkflowExecutionEvent(notificationsConfig, emailNotification, request, model)
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
