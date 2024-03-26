package errors

import (
	"fmt"
	"testing"

	extErrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestErrorf(t *testing.T) {
	msg := "msg"
	err := Errorf(IllegalStateError, "w1", "Message [%v]", msg)
	assert.NotNil(t, err)
	e := err.(*WorkflowError)
	assert.Equal(t, IllegalStateError, e.Code)
	assert.Equal(t, "w1", e.Workflow)
	assert.Equal(t, fmt.Sprintf("Message [%v]", msg), e.Message)
	assert.Equal(t, err, extErrors.Cause(e))
	assert.Equal(t, "Workflow[w1] failed. IllegalStateError: Message [msg]", err.Error())
}

func TestErrorfWithCause(t *testing.T) {
	cause := extErrors.Errorf("Some Error")
	msg := "msg"
	err := Wrapf(IllegalStateError, "w1", cause, "Message [%v]", msg)
	assert.NotNil(t, err)
	e := err.(*WorkflowErrorWithCause)
	assert.Equal(t, IllegalStateError, e.Code)
	assert.Equal(t, "w1", e.Workflow)
	assert.Equal(t, fmt.Sprintf("Message [%v]", msg), e.Message)
	assert.Equal(t, cause, extErrors.Cause(e))
	assert.Equal(t, "Workflow[w1] failed. IllegalStateError: Message [msg], caused by: Some Error", err.Error())
}

func TestMatches(t *testing.T) {
	err := Errorf(IllegalStateError, "w1", "Message ")
	assert.True(t, Matches(err, IllegalStateError))
	assert.False(t, Matches(err, BadSpecificationError))

	cause := extErrors.Errorf("Some Error")
	err = Wrapf(IllegalStateError, "w1", cause, "Message ")
	assert.True(t, Matches(err, IllegalStateError))
	assert.False(t, Matches(err, BadSpecificationError))

	assert.False(t, Matches(cause, IllegalStateError))
	assert.False(t, Matches(cause, BadSpecificationError))
}
