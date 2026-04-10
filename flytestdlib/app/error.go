package app

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
