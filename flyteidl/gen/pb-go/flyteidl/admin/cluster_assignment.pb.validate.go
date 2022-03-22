// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: flyteidl/admin/cluster_assignment.proto

package admin

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
)

// define the regex for a UUID once up-front
var _cluster_assignment_uuidPattern = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

// Validate checks the field values on ClusterAssignment with the rules defined
// in the proto definition for this message. If any rules are violated, an
// error is returned.
func (m *ClusterAssignment) Validate() error {
	if m == nil {
		return nil
	}

	if v, ok := interface{}(m.GetAffinity()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return ClusterAssignmentValidationError{
				field:  "Affinity",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	return nil
}

// ClusterAssignmentValidationError is the validation error returned by
// ClusterAssignment.Validate if the designated constraints aren't met.
type ClusterAssignmentValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ClusterAssignmentValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ClusterAssignmentValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ClusterAssignmentValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ClusterAssignmentValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ClusterAssignmentValidationError) ErrorName() string {
	return "ClusterAssignmentValidationError"
}

// Error satisfies the builtin error interface
func (e ClusterAssignmentValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sClusterAssignment.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ClusterAssignmentValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ClusterAssignmentValidationError{}

// Validate checks the field values on Affinity with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *Affinity) Validate() error {
	if m == nil {
		return nil
	}

	for idx, item := range m.GetSelectors() {
		_, _ = idx, item

		if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return AffinityValidationError{
					field:  fmt.Sprintf("Selectors[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	return nil
}

// AffinityValidationError is the validation error returned by
// Affinity.Validate if the designated constraints aren't met.
type AffinityValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e AffinityValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e AffinityValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e AffinityValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e AffinityValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e AffinityValidationError) ErrorName() string { return "AffinityValidationError" }

// Error satisfies the builtin error interface
func (e AffinityValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sAffinity.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = AffinityValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = AffinityValidationError{}

// Validate checks the field values on Selector with the rules defined in the
// proto definition for this message. If any rules are violated, an error is returned.
func (m *Selector) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Key

	// no validation rules for Operator

	return nil
}

// SelectorValidationError is the validation error returned by
// Selector.Validate if the designated constraints aren't met.
type SelectorValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e SelectorValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e SelectorValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e SelectorValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e SelectorValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e SelectorValidationError) ErrorName() string { return "SelectorValidationError" }

// Error satisfies the builtin error interface
func (e SelectorValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sSelector.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = SelectorValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = SelectorValidationError{}
