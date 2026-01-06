package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorCollection_ErrorOrDefault(t *testing.T) {
	errs := ErrorCollection{}
	assert.Error(t, errs)
	assert.NoError(t, errs.ErrorOrDefault())
}

func TestErrorCollection_Append(t *testing.T) {
	errs := ErrorCollection{}
	errs.Append(nil)
	errs.Append(fmt.Errorf("this is an actual error"))
	assert.Error(t, errs.ErrorOrDefault())
	assert.Len(t, errs, 1)
	assert.Len(t, errs.ErrorOrDefault(), 1)
}

func TestErrorCollection_Error(t *testing.T) {
	errs := ErrorCollection{}
	errs.Append(nil)
	errs.Append(fmt.Errorf("this is an actual error"))
	assert.Error(t, errs.ErrorOrDefault())
	assert.Contains(t, errs.ErrorOrDefault().Error(), "this is an actual error")
}
