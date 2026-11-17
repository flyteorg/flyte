package service

// ErrorKind identifies a service failure for errors.Is.
type ErrorKind string

const (
	ErrFailedPrecondition ErrorKind = "failed precondition"
	ErrAlreadyExists      ErrorKind = "already exists"
	ErrInvalidArgument    ErrorKind = "invalid argument"
	ErrNotFound           ErrorKind = "not found"
	ErrInternal           ErrorKind = "internal"
	ErrUnimplemented      ErrorKind = "unimplemented"
)

func (k ErrorKind) Error() string          { return string(k) }
func (k ErrorKind) wrap(cause error) error { return &serviceError{kind: k, cause: cause} }

type serviceError struct {
	kind  ErrorKind
	cause error
}

func (e *serviceError) Error() string   { return e.cause.Error() }
func (e *serviceError) Unwrap() []error { return []error{e.kind, e.cause} }

// RunErrorKind and the ErrRun constants are the run API names for service errors.
type RunErrorKind = ErrorKind

const (
	ErrRunInvalidArgument = ErrInvalidArgument
	ErrRunNotFound        = ErrNotFound
	ErrRunInternal        = ErrInternal
	ErrRunUnimplemented   = ErrUnimplemented
)
