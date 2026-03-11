//go:build !console

package console

import (
	"context"

	"github.com/flyteorg/flyte/v2/app"
)

func Setup(_ context.Context, _ *app.SetupContext) error {
	return nil
}
