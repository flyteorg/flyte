package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorf(t *testing.T) {
	e := Errorf("Code1", "msg")
	assert.NotNil(t, e)
	assert.Equal(t, "[Code1] msg", e.Error())
}

func TestWrapf(t *testing.T) {
	e := Wrapf("Code1", fmt.Errorf("test error"), "msg")
	assert.NotNil(t, e)
	assert.Equal(t, "[Code1] msg, caused by: test error", e.Error())
}

func TestGetErrorCode(t *testing.T) {
	e := Errorf("Code1", "msg")
	assert.NotNil(t, e)
	code, found := GetErrorCode(e)
	assert.True(t, found)
	assert.Equal(t, "Code1", code)
}

func TestIsCausedBy(t *testing.T) {
	e := Errorf("Code1", "msg")
	assert.NotNil(t, e)

	e = Wrapf("Code2", e, "msg")
	assert.True(t, IsCausedBy(e, "Code1"))
	assert.True(t, IsCausedBy(e, "Code2"))

	e = fmt.Errorf("new err caused by: %w", e)
	assert.True(t, IsCausedBy(e, "Code1"))

	e = fmt.Errorf("not sharing code err")
	assert.False(t, IsCausedBy(e, "Code1"))
}

func TestIsCausedByError(t *testing.T) {
	eRoot := Errorf("Code1", "msg")
	assert.NotNil(t, eRoot)

	e1 := Wrapf("Code2", eRoot, "msg")
	assert.True(t, IsCausedByError(e1, eRoot))

	e2 := Wrapf("Code3", e1, "msg")
	assert.True(t, IsCausedByError(e2, eRoot))
	assert.True(t, IsCausedByError(e2, e1))

	e3 := fmt.Errorf("default errors. caused by: %w", e2)
	assert.True(t, IsCausedByError(e3, eRoot))
	assert.True(t, IsCausedByError(e3, e1))
}

func TestErrorsIs(t *testing.T) {
	eRoot := Errorf("Code1", "msg")
	assert.True(t, errors.Is(eRoot, Errorf("Code1", "different msg")))

	e1 := Wrapf("Code2", eRoot, "Wrapped error")
	assert.True(t, errors.Is(e1, Errorf("Code1", "different msg")))
}

func TestErrorsUnwrap(t *testing.T) {
	eRoot := Errorf("Code1", "msg")
	e1 := Wrapf("Code2", eRoot, "Wrapped error")
	assert.True(t, errors.Is(e1, Errorf("Code1", "different msg")))

	newErr := &err{}
	assert.True(t, errors.As(e1, &newErr))
	assert.Equal(t, "Code1", newErr.Code())

	assert.True(t, errors.Is(errors.Unwrap(e1), Errorf("Code1", "different msg")))
}
