package errors

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	invalidArgFormat   = "invalid value for %s, value:[%s]"
	missingFieldFormat = "missing %s"
	notFound           = "missing entity of type %s with identifier %v"
)

type CacheServiceError interface {
	Error() string
	Code() codes.Code
	GRPCStatus() *status.Status
	WithDetails(details proto.Message) (CacheServiceError, error)
	String() string
}

type cacheServiceErrorImpl struct {
	status *status.Status
}

func (e *cacheServiceErrorImpl) Error() string {
	return e.status.Message()
}

func (e *cacheServiceErrorImpl) Code() codes.Code {
	return e.status.Code()
}

func (e *cacheServiceErrorImpl) GRPCStatus() *status.Status {
	return e.status
}

func (e *cacheServiceErrorImpl) String() string {
	return fmt.Sprintf("status: %v", e.status)
}

func NewCacheServiceError(code codes.Code, message string) error {
	return &cacheServiceErrorImpl{
		status: status.New(code, message),
	}
}

func NewCacheServiceErrorf(code codes.Code, format string, a ...interface{}) error {
	return NewCacheServiceError(code, fmt.Sprintf(format, a...))
}

func NewNotFoundError(entityType string, key string) error {
	return NewCacheServiceErrorf(codes.NotFound, notFound, entityType, key)
}

func NewMissingArgumentError(field string) error {
	return NewCacheServiceErrorf(codes.InvalidArgument, fmt.Sprintf(missingFieldFormat, field))
}

func NewInvalidArgumentError(field string, value string) error {
	return NewCacheServiceErrorf(codes.InvalidArgument, fmt.Sprintf(invalidArgFormat, field, value))
}
