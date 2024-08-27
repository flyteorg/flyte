package yunikorn

import (
	"testing"
	"gotest.tools/assert"
)

func TestNewPlugin(t *testing.T) {
	tests := []struct{
		input string
		expect *Plugin
	}{
		{
			input: "",
			expect: &Plugin{
				Parameters:  "",
			},
		},
		{
			input: "placeholderTimeoutInSeconds=30 gangSchedulingStyle=Hard",
			expect: &Plugin{
				Parameters:  "placeholderTimeoutInSeconds=30 gangSchedulingStyle=Hard",
			},
		},
	}
	t.Run("New Yunikorn plugin", func(t *testing.T) {
		got := NewPlugin(t.input)
		assert.NotNil(t, got)
		assert.Equal(t, t.input, got.Parameters)
	})
}

func TestProcess(t *testing.T) {
	tests := []struct{
		input interface{}
		expect error
	}{
		{input: 1, expect: nil},
		{input: "test", expect: nil},
	}
	t.Run("Yunikorn plugin process any type", func(t *testing.T) {
		got := NewPlugin(t.input)
		assert.NotNil(t, got)
		assert.Equal(t, t.input, got.Parameters)
	})
}