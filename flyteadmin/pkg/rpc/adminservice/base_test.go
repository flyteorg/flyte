package adminservice

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

func Test_interceptPanic(t *testing.T) {
	m := AdminService{
		Metrics: InitMetrics(promutils.NewTestScope()),
	}

	ctx := context.Background()

	// Mute logs to avoid .Fatal() (called in interceptPanic) causing the process to close
	assert.NoError(t, logger.SetConfig(&logger.Config{Mute: true}))

	func() {
		defer func() {
			if err := recover(); err != nil {
				assert.Fail(t, "Unexpected error", err)
			}
		}()

		a := func() {
			defer m.interceptPanic(ctx, proto.Message(nil))

			var x *int
			*x = 10
		}

		a()
	}()
}
