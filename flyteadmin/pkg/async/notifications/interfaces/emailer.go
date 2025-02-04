package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
)

//go:generate mockery-v2 --name=Emailer --output=../mocks --case=underscore --with-expecter

// The implementation of Emailer needs to be passed to the implementation of Processor
// in order for emails to be sent.
type Emailer interface {
	SendEmail(ctx context.Context, email *admin.EmailMessage) error
}
