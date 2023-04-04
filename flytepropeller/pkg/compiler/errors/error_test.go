package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func addError(errs CompileErrors) {
	errs.Collect(NewValueRequiredErr("node", "param"))
}

func TestCompileErrors_Collect(t *testing.T) {
	errs := NewCompileErrors()
	assert.False(t, errs.HasErrors())
	addError(errs)
	assert.True(t, errs.HasErrors())
}

func TestCompileErrors_NewScope(t *testing.T) {
	errs := NewCompileErrors()
	addError(errs.NewScope().NewScope())
	assert.True(t, errs.HasErrors())
	assert.Equal(t, 1, errs.ErrorCount())
}

func TestCompileErrors_Errors(t *testing.T) {
	errs := NewCompileErrors()
	addError(errs.NewScope().NewScope())
	addError(errs.NewScope().NewScope())
	assert.True(t, errs.HasErrors())
	assert.Equal(t, 2, errs.ErrorCount())

	set := errs.Errors()
	assert.Equal(t, 1, len(*set))
}

func TestCompileErrors_Error(t *testing.T) {
	errs := NewCompileErrors()
	addError(errs.NewScope().NewScope())
	addError(errs.NewScope().NewScope())
	assert.NotEqual(t, "", errs.Error())
}
