package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
)

// The implementation of Emailer needs to be passed to the implementation of Processor
// in order for emails to be sent.
type Emailer interface {
	SendEmail(ctx context.Context, email admin.EmailMessage) error
}
