package config

import (
	"testing"

	"github.com/flyteorg/flytectl/pkg/printer"
	"github.com/stretchr/testify/assert"
)

func TestOutputFormat(t *testing.T) {
	c := &Config{
		Output: "json",
	}
	result, err := c.OutputFormat()
	assert.Nil(t, err)
	assert.Equal(t, printer.OutputFormat(1), result)
}

func TestInvalidOutputFormat(t *testing.T) {
	c := &Config{
		Output: "test",
	}
	var result printer.OutputFormat
	defer func() {
		if r := recover(); r != nil {
			assert.Equal(t, printer.OutputFormat(0), result)
			assert.NotNil(t, r)
		}
	}()
	result = c.MustOutputFormat()

}
