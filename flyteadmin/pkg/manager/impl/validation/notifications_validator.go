package validation

import (
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

// TODO: maybe add more sophisticated email validation.
func validateRecipientsEmail(recipients []string) error {
	if len(recipients) == 0 {
		return shared.GetMissingArgumentError("recipients")
	}
	for _, recipient := range recipients {
		if len(recipient) == 0 {
			return shared.GetMissingArgumentError("recipient")
		}
	}
	return nil
}

func validateNotifications(notifications []*admin.Notification) error {
	for _, notif := range notifications {
		switch {
		case notif.GetEmail() != nil:
			if err := validateRecipientsEmail(notif.GetEmail().GetRecipientsEmail()); err != nil {
				return err
			}
		case notif.GetSlack() != nil:
			if err := validateRecipientsEmail(notif.GetSlack().GetRecipientsEmail()); err != nil {
				return err
			}
		case notif.GetPagerDuty() != nil:
			if err := validateRecipientsEmail(notif.GetPagerDuty().GetRecipientsEmail()); err != nil {
				return err
			}
		default:
			return shared.GetInvalidArgumentError("notification type")
		}

		for _, phase := range notif.GetPhases() {
			if !common.IsExecutionTerminal(phase) {
				return shared.GetInvalidArgumentError("phase")
			}
		}
	}

	return nil
}
