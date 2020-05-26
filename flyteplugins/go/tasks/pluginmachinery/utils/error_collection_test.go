package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorCollection(t *testing.T) {
	ec := ErrorCollection{}

	assert.Empty(t, ec.Error())

	ec.Errors = append(ec.Errors, fmt.Errorf("error1"))
	assert.NotEmpty(t, ec.Error())

	ec.Errors = append(ec.Errors, fmt.Errorf("error2"))
	assert.NotEmpty(t, ec.Error())

	assert.Equal(t, "0: error1\r\n1: error2\r\n", ec.Error())
}
