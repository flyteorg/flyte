package error

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/protoadapt"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/utils/set"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

var grpcToConnectCodes = map[codes.Code]connect.Code{
	codes.Canceled:           connect.CodeCanceled,
	codes.Unknown:            connect.CodeUnknown,
	codes.InvalidArgument:    connect.CodeInvalidArgument,
	codes.DeadlineExceeded:   connect.CodeDeadlineExceeded,
	codes.NotFound:           connect.CodeNotFound,
	codes.AlreadyExists:      connect.CodeAlreadyExists,
	codes.PermissionDenied:   connect.CodePermissionDenied,
	codes.ResourceExhausted:  connect.CodeResourceExhausted,
	codes.FailedPrecondition: connect.CodeFailedPrecondition,
	codes.Aborted:            connect.CodeAborted,
	codes.OutOfRange:         connect.CodeOutOfRange,
	codes.Unimplemented:      connect.CodeUnimplemented,
	codes.Internal:           connect.CodeInternal,
	codes.Unavailable:        connect.CodeUnavailable,
	codes.DataLoss:           connect.CodeDataLoss,
	codes.Unauthenticated:    connect.CodeUnauthenticated,
}

var connectToGrpcCodes = map[connect.Code]codes.Code{
	connect.CodeCanceled:           codes.Canceled,
	connect.CodeUnknown:            codes.Unknown,
	connect.CodeInvalidArgument:    codes.InvalidArgument,
	connect.CodeDeadlineExceeded:   codes.DeadlineExceeded,
	connect.CodeNotFound:           codes.NotFound,
	connect.CodeAlreadyExists:      codes.AlreadyExists,
	connect.CodePermissionDenied:   codes.PermissionDenied,
	connect.CodeResourceExhausted:  codes.ResourceExhausted,
	connect.CodeFailedPrecondition: codes.FailedPrecondition,
	connect.CodeAborted:            codes.Aborted,
	connect.CodeOutOfRange:         codes.OutOfRange,
	connect.CodeUnimplemented:      codes.Unimplemented,
	connect.CodeInternal:           codes.Internal,
	connect.CodeUnavailable:        codes.Unavailable,
	connect.CodeDataLoss:           codes.DataLoss,
	connect.CodeUnauthenticated:    codes.Unauthenticated,
}

var httpToConnectCodes = map[int32]connect.Code{
	499:                            connect.CodeCanceled, // Client closed request
	http.StatusBadRequest:          connect.CodeInvalidArgument,
	http.StatusConflict:            connect.CodeAlreadyExists,
	http.StatusForbidden:           connect.CodePermissionDenied,
	http.StatusGatewayTimeout:      connect.CodeDeadlineExceeded,
	http.StatusInternalServerError: connect.CodeUnknown,
	http.StatusNotFound:            connect.CodeNotFound,
	http.StatusNotImplemented:      connect.CodeUnimplemented,
	http.StatusServiceUnavailable:  connect.CodeUnavailable,
	http.StatusTooManyRequests:     connect.CodeResourceExhausted,
	http.StatusUnauthorized:        connect.CodeUnauthenticated,
}

type grpcCode interface {
	Code() codes.Code
}

// ToConnectError converts an error with a grpc status code to a connect error with the corresponding status code.
// Otherwise return the original error, which the framework will convert to a default connect error downstream.
// https://connectrpc.com/docs/go/errors#working-with-errors
func ToConnectError(err error) error {
	if err == nil {
		return nil
	}

	ctx := context.Background()
	var grpcCodeErr grpcCode
	var apiStatus apiErrors.APIStatus
	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		return err
	} else if st, statusOK := status.FromError(err); statusOK {
		code, codeOK := grpcToConnectCodes[st.Code()]
		if codeOK {
			connectErr = connect.NewError(code, err)
			for _, detail := range st.Details() {
				detailProto, ok := detail.(protoadapt.MessageV1)
				if !ok {
					logger.Errorf(ctx, "status detail [%v] does not implement proto.Message", detail)
					continue
				}
				errDetail, convertErr := connect.NewErrorDetail(protoadapt.MessageV2Of(detailProto))
				if convertErr != nil {
					logger.Errorf(ctx, "failed to convert status detail to v2 proto: %v", convertErr)
					continue
				}
				connectErr.AddDetail(errDetail)
			}
			return connectErr
		}
	} else if errors.As(err, &grpcCodeErr) {
		if code, ok := grpcToConnectCodes[grpcCodeErr.Code()]; ok {
			return connect.NewError(code, err)
		}
	} else if errors.As(err, &apiStatus) {
		v1Status := apiStatus.Status()
		if code, codeOK := httpToConnectCodes[v1Status.Code]; codeOK {
			connectErr = connect.NewError(code, err)
			if v1Status.Details == nil {
				logger.Errorf(ctx, "v1Status is nil")
				return connectErr
			}

			errDetail, convertErr := connect.NewErrorDetail(protoadapt.MessageV2Of(v1Status.Details))
			if convertErr != nil {
				logger.Errorf(ctx, "failed to convert status detail to v2 proto: %v", convertErr)
			} else {
				connectErr.AddDetail(errDetail)
			}

			return connectErr
		}
	}
	// Use the default conversion downstream
	return err
}

// ToConnectErrorf formats according to the format specifier, returns the string
// as a value that satisfies error, and converts it to a connect error.
func ToConnectErrorf(format string, args ...any) error {
	return ToConnectError(fmt.Errorf(format, args...))
}

func FromConnectError(err error) error {
	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		return status.Error(codes.Code(connectErr.Code()), connectErr.Error())
	}
	return err
}

func ConnectErrorCode(err error) connect.Code {
	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		return connectErr.Code()
	}
	if st, ok := status.FromError(err); ok {
		if code, ok := grpcToConnectCodes[st.Code()]; ok {
			return code
		}
		return connect.CodeUnknown
	}
	var grpcCodeErr grpcCode
	if errors.As(err, &grpcCodeErr) {
		if code, ok := grpcToConnectCodes[grpcCodeErr.Code()]; ok {
			return code
		}
		return connect.CodeUnknown
	}
	var apiStatus apiErrors.APIStatus
	if errors.As(err, &apiStatus) {
		if code, ok := httpToConnectCodes[apiStatus.Status().Code]; ok {
			return code
		}
		return connect.CodeUnknown
	}
	if errors.Is(err, context.Canceled) {
		return connect.CodeCanceled
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return connect.CodeDeadlineExceeded
	}
	return connect.CodeUnknown
}

func GrpcErrorCode(err error) codes.Code {
	if err == nil {
		return codes.OK
	}
	connectCode := ConnectErrorCode(err)
	if grpcCode, ok := connectToGrpcCodes[connectCode]; ok {
		return grpcCode
	}
	return codes.Unknown
}

type ServerError interface {
	Error() string
	Code() codes.Code
	WithDetails(details proto.Message) (ServerError, error)
}

type serverError struct {
	status *status.Status
}

func (e *serverError) Error() string {
	return e.status.Message()
}

func (e *serverError) Code() codes.Code {
	return e.status.Code()
}

func (e *serverError) WithDetails(details proto.Message) (ServerError, error) {
	s, err := e.status.WithDetails(details)
	if err != nil {
		return nil, err
	}
	return NewServerErrorFromStatus(s), nil
}

func (e *serverError) GRPCStatus() *status.Status {
	return e.status
}

func NewServerError(code codes.Code, message string) ServerError {
	return &serverError{
		status: status.New(code, message),
	}
}

func NewServerErrorf(code codes.Code, format string, a ...interface{}) ServerError {
	return NewServerError(code, fmt.Sprintf(format, a...))
}

func NewServerErrorFromStatus(status *status.Status) ServerError {
	return &serverError{
		status: status,
	}
}

func NewCollectedErr(code codes.Code, errors []error) ServerError {
	errMsg := fmt.Sprintf("Collected Errors: %v\n", len(errors))
	for idx, e := range errors {
		errMsg += fmt.Sprintf("\tError %d: %s\n", idx, e.Error())
	}

	return NewServerError(code, errMsg)
}

func IsGrpcErrorWithCode(code codes.Code, err error) bool {
	s, ok := status.FromError(err)
	return ok && s.Code() == code
}

func IsAlreadyExists(err error) bool {
	return IsGrpcErrorWithCode(codes.AlreadyExists, err)
}

func IsNotFound(err error) bool {
	return IsGrpcErrorWithCode(codes.NotFound, err)
}

func GetGrpcErrorCodeToPropagate(observed error, propagateErrors set.Set[codes.Code]) codes.Code {
	s, ok := status.FromError(observed)
	if ok && propagateErrors.HasAny(s.Code()) {
		return s.Code()
	}
	return codes.Internal
}

func GrpcPropagatableError(err error, propagateErrors set.Set[codes.Code], operation string) error {
	s, ok := status.FromError(err)
	if ok && propagateErrors.HasAny(s.Code()) {
		return status.Errorf(s.Code(), "unable to %s due to: %s", operation, s.Message())
	}
	return status.Errorf(codes.Internal, "unable to %s", operation)
}

func IsConnectErrorWithCode(code connect.Code, err error) bool {
	var ce *connect.Error
	return errors.As(err, &ce) && ce.Code() == code
}

func IsUnavailable(err error) bool {
	if err == nil {
		return false
	}
	var s interface{ GRPCStatus() *status.Status }
	if errors.As(err, &s) {
		if s.GRPCStatus().Code() == codes.Unavailable {
			return true
		}
	}

	if connect.CodeOf(err) == connect.CodeUnavailable {
		return true
	}

	errMsg := err.Error()
	// catch all for envoy or service mesh errors
	return strings.Contains(errMsg, "service unavailable") ||
		strings.Contains(errMsg, "unavailable:")
}
