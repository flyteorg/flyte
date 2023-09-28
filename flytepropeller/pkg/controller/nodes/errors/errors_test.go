package errors

import (
	"fmt"
	"testing"

	extErrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestErrorf(t *testing.T) {
	msg := "msg"
	err := Errorf(IllegalStateError, "n1", "Message [%v]", msg)
	assert.NotNil(t, err)
	e := err.(*NodeError)
	assert.Equal(t, IllegalStateError, e.ErrCode)
	assert.Equal(t, "n1", e.Node)
	assert.Equal(t, fmt.Sprintf("Message [%v]", msg), e.Message)
	assert.Equal(t, err, extErrors.Cause(e))
	assert.Equal(t, "failed at Node[n1]. IllegalStateError: Message [msg]", err.Error())
}

func TestErrorfWithCause(t *testing.T) {
	cause := extErrors.Errorf("Some Error")
	msg := "msg"
	err := Wrapf(IllegalStateError, "n1", cause, "Message [%v]", msg)
	assert.NotNil(t, err)
	e := err.(*NodeErrorWithCause)
	nodeErr := e.NodeError.(*NodeError)
	assert.Equal(t, IllegalStateError, nodeErr.ErrCode)
	assert.Equal(t, "n1", nodeErr.Node)
	assert.Equal(t, fmt.Sprintf("Message [%v]", msg), nodeErr.Message)
	assert.Equal(t, cause, extErrors.Cause(e))
	assert.Equal(t, "failed at Node[n1]. IllegalStateError: Message [msg], caused by: Some Error", err.Error())
}

func TestMatches(t *testing.T) {
	err := Errorf(IllegalStateError, "n1", "Message ")
	assert.True(t, Matches(err, IllegalStateError))
	assert.False(t, Matches(err, BadSpecificationError))

	cause := extErrors.Errorf("Some Error")
	err = Wrapf(IllegalStateError, "n1", cause, "Message ")
	assert.True(t, Matches(err, IllegalStateError))
	assert.False(t, Matches(err, BadSpecificationError))

	assert.False(t, Matches(cause, IllegalStateError))
	assert.False(t, Matches(cause, BadSpecificationError))
}
