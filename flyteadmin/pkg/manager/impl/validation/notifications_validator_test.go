package validation

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var recipients = []string{"foo@example.com"}
var emptyRecipients = make([]string, 0)

func assertInvalidArgument(t *testing.T, err error) {
	s, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, s.Code())
}

func TestValidateRecipientsEmail(t *testing.T) {
	t.Run("valid emails", func(t *testing.T) {
		assert.NoError(t, validateRecipientsEmail(recipients))
	})
	t.Run("invalid recipients", func(t *testing.T) {
		err := validateRecipientsEmail(nil)
		assertInvalidArgument(t, err)
	})
	t.Run("invalid recipient", func(t *testing.T) {
		err := validateRecipientsEmail(emptyRecipients)
		assertInvalidArgument(t, err)
	})
}

func TestValidateNotifications(t *testing.T) {
	phases := []core.WorkflowExecution_Phase{
		core.WorkflowExecution_FAILED,
	}
	t.Run("email type", func(t *testing.T) {
		err := validateNotifications([]*admin.Notification{
			{
				Type: &admin.Notification_Email{
					Email: &admin.EmailNotification{
						RecipientsEmail: recipients,
					},
				},
				Phases: phases,
			},
		})
		assert.NoError(t, err)
	})
	t.Run("email type - invalid", func(t *testing.T) {
		err := validateNotifications([]*admin.Notification{
			{
				Type: &admin.Notification_Email{
					Email: &admin.EmailNotification{
						RecipientsEmail: emptyRecipients,
					},
				},
				Phases: phases,
			},
		})
		assertInvalidArgument(t, err)
	})
	t.Run("slack type", func(t *testing.T) {
		err := validateNotifications([]*admin.Notification{
			{
				Type: &admin.Notification_Slack{
					Slack: &admin.SlackNotification{
						RecipientsEmail: recipients,
					},
				},
				Phases: phases,
			},
		})
		assert.NoError(t, err)
	})
	t.Run("slack type - invalid", func(t *testing.T) {
		err := validateNotifications([]*admin.Notification{
			{
				Type: &admin.Notification_Slack{
					Slack: &admin.SlackNotification{
						RecipientsEmail: emptyRecipients,
					},
				},
				Phases: phases,
			},
		})
		assertInvalidArgument(t, err)
	})
	t.Run("pagerduty type", func(t *testing.T) {
		err := validateNotifications([]*admin.Notification{
			{
				Type: &admin.Notification_PagerDuty{
					PagerDuty: &admin.PagerDutyNotification{
						RecipientsEmail: recipients,
					},
				},
				Phases: phases,
			},
		})
		assert.NoError(t, err)
	})
	t.Run("pagerduty type - invalid", func(t *testing.T) {
		err := validateNotifications([]*admin.Notification{
			{
				Type: &admin.Notification_PagerDuty{
					PagerDuty: &admin.PagerDutyNotification{
						RecipientsEmail: emptyRecipients,
					},
				},
				Phases: phases,
			},
		})
		assertInvalidArgument(t, err)
	})
	t.Run("invalid recipients", func(t *testing.T) {
		err := validateNotifications([]*admin.Notification{
			{
				Type: &admin.Notification_PagerDuty{
					PagerDuty: &admin.PagerDutyNotification{},
				},
				Phases: phases,
			},
		})
		assertInvalidArgument(t, err)
	})
	t.Run("invalid phases", func(t *testing.T) {
		err := validateNotifications([]*admin.Notification{
			{
				Type: &admin.Notification_PagerDuty{
					PagerDuty: &admin.PagerDutyNotification{
						RecipientsEmail: recipients,
					},
				},
				Phases: []core.WorkflowExecution_Phase{
					core.WorkflowExecution_QUEUED,
				},
			},
		})
		assertInvalidArgument(t, err)
	})
}
