package errors

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DataCatalogError interface {
	Error() string
	Code() codes.Code
	GRPCStatus() *status.Status
	String() string
}

type dataCatalogErrorImpl struct {
	status *status.Status
}

func (e *dataCatalogErrorImpl) Error() string {
	return e.status.Message()
}

func (e *dataCatalogErrorImpl) Code() codes.Code {
	return e.status.Code()
}

func (e *dataCatalogErrorImpl) GRPCStatus() *status.Status {
	return e.status
}

func (e *dataCatalogErrorImpl) String() string {
	return fmt.Sprintf("status: %v", e.status)
}

func NewDataCatalogError(code codes.Code, message string) error {
	return &dataCatalogErrorImpl{
		status: status.New(code, message),
	}
}

func NewDataCatalogErrorf(code codes.Code, format string, a ...interface{}) error {
	return NewDataCatalogError(code, fmt.Sprintf(format, a...))
}

func NewCollectedErrors(code codes.Code, errors []error) error {
	errorCollection := make([]string, len(errors))
	for idx, err := range errors {
		errorCollection[idx] = err.Error()
	}

	return NewDataCatalogError(code, strings.Join(errorCollection, ", "))
}

func IsAlreadyExistsError(err error) bool {
	dcErr, ok := err.(DataCatalogError)
	return ok && dcErr.GRPCStatus().Code() == codes.AlreadyExists
}

func IsDoesNotExistError(err error) bool {
	dcErr, ok := err.(DataCatalogError)
	return ok && dcErr.GRPCStatus().Code() == codes.NotFound
}
