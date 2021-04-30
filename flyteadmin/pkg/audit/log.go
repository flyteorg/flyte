package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flytestdlib/logger"
	"google.golang.org/grpc/codes"
)

type LogBuilder interface {
	WithAuthenticatedCtx(ctx context.Context) LogBuilder
	WithRequest(method string, parameters map[string]string, mode AccessMode, requestedAt time.Time) LogBuilder
	WithResponse(sentAt time.Time, err error) LogBuilder
	Log(ctx context.Context)
}

type logBuilder struct {
	auditLog Message
	readOnly bool
}

func (b *logBuilder) WithAuthenticatedCtx(ctx context.Context) LogBuilder {
	clientMeta := ctx.Value(common.AuditFieldsContextKey)
	switch m := clientMeta.(type) {
	case AuthenticatedClientMeta:
		b.auditLog.Principal = Principal{
			Subject:       m.Subject,
			TokenIssuedAt: m.TokenIssuedAt,
		}
		if len(m.ClientIds) > 0 {
			b.auditLog.Principal.ClientID = m.ClientIds[0]
		}
		b.auditLog.Client = Client{
			ClientIP: m.ClientIP,
		}
	default:
		logger.Warningf(ctx, "Failed to parse authenticated client metadata when creating audit log")
	}
	return b
}

// TODO: Also look into passing down HTTP verb
func (b *logBuilder) WithRequest(method string, parameters map[string]string, mode AccessMode,
	requestedAt time.Time) LogBuilder {
	b.auditLog.Request = Request{
		Method:     method,
		Parameters: parameters,
		Mode:       mode,
		ReceivedAt: requestedAt,
	}
	return b
}

func (b *logBuilder) WithResponse(sentAt time.Time, err error) LogBuilder {
	responseCode := codes.OK.String()
	if err != nil {
		switch err := err.(type) {
		case errors.FlyteAdminError:
			responseCode = err.(errors.FlyteAdminError).Code().String()
		default:
			responseCode = codes.Internal.String()
		}
	}
	b.auditLog.Response = Response{
		ResponseCode: responseCode,
		SentAt:       sentAt,
	}
	return b
}

func (b *logBuilder) formatLogString(ctx context.Context) string {
	auditLog, err := json.Marshal(&b.auditLog)
	if err != nil {
		logger.Warningf(ctx, "Failed to marshal audit log to protobuf with err: %v", err)
	}
	return fmt.Sprintf("Recording request: [%s]", auditLog)
}

func (b *logBuilder) Log(ctx context.Context) {
	if b.readOnly {
		logger.Warningf(ctx, "Attempting to record audit log for request: [%+v] more than once. Aborting.", b.auditLog.Request)
	}
	defer func() {
		b.readOnly = true
	}()
	logger.Info(ctx, b.formatLogString(ctx))
}

func NewLogBuilder() LogBuilder {
	return &logBuilder{}
}
