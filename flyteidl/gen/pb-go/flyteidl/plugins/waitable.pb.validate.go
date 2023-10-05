// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: flyteidl/plugins/waitable.proto

package plugins

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang/protobuf/ptypes"

	core "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = ptypes.DynamicAny{}

	_ = core.WorkflowExecution_Phase(0)
)

// define the regex for a UUID once up-front
var _waitable_uuidPattern = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

// Validate checks the field values on Waitable with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *Waitable) Validate() error {
	if m == nil {
		return nil
	}

	if v, ok := interface{}(m.GetWfExecId()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return WaitableValidationError{
				field:  "WfExecId",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	// no validation rules for Phase

	// no validation rules for WorkflowId

	return nil
}

// WaitableValidationError is the validation error returned by
// Waitable.Validate if the designated constraints aren't met.
type WaitableValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e WaitableValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e WaitableValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e WaitableValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e WaitableValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e WaitableValidationError) ErrorName() string { return "WaitableValidationError" }

// Error satisfies the builtin error interface
func (e WaitableValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sWaitable.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = WaitableValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = WaitableValidationError{}
