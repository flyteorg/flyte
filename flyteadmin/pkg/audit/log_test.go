package audit

import (
	"context"
	"testing"
	"time"

	"github.com/lyft/flyteadmin/pkg/common"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

func TestLogBuilderLog(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, common.PrincipalContextKey, "prince")
	tokenIssuedAt := time.Date(2020, time.January, 5, 10, 15, 0, 0, time.UTC)
	requestedAt := time.Date(2020, time.January, 5, 10, 30, 0, 0, time.UTC)
	sentAt := time.Date(2020, time.January, 5, 10, 31, 0, 0, time.UTC)
	err := errors.NewFlyteAdminError(codes.AlreadyExists, "womp womp")
	ctx = context.WithValue(ctx, common.AuditFieldsContextKey, AuthenticatedClientMeta{
		ClientIds:     []string{"12345"},
		TokenIssuedAt: tokenIssuedAt,
		ClientIP:      "192.0.2.1:25",
	})
	builder := NewLogBuilder().WithAuthenticatedCtx(ctx).WithRequest(
		"my_method", map[string]string{
			"my": "params",
		}, ReadWrite, requestedAt).WithResponse(sentAt, err)
	assert.EqualValues(t, "Recording request: [{\"Principal\":{\"Subject\":\"prince\",\"ClientID\":\"12345\","+
		"\"TokenIssuedAt\":\"2020-01-05T10:15:00Z\"},\"Client\":{\"ClientIP\":\"192.0.2.1:25\"},\"Request\":"+
		"{\"Method\":\"my_method\",\"Parameters\":{\"my\":\"params\"},\"Mode\":1,\"ReceivedAt\":"+
		"\"2020-01-05T10:30:00Z\"},\"Response\":{\"ResponseCode\":\"AlreadyExists\",\"SentAt\":"+
		"\"2020-01-05T10:31:00Z\"}}]", builder.(*logBuilder).formatLogString(context.TODO()))
}
