package commandutils

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCase struct {
	Input  io.Reader `json:"input"`
	Output bool      `json:"output"`
}

func TestAskForConfirmation(t *testing.T) {
	tests := []TestCase{
		{
			Input:  strings.NewReader("yes"),
			Output: true,
		},
		{
			Input:  strings.NewReader("y"),
			Output: true,
		},
		{
			Input:  strings.NewReader("no"),
			Output: false,
		},
		{
			Input:  strings.NewReader("n"),
			Output: false,
		},
		{
			Input:  strings.NewReader("No"),
			Output: false,
		},
		{
			Input:  strings.NewReader("Yes"),
			Output: true,
		},
		{
			Input:  strings.NewReader(""),
			Output: false,
		},
	}
	for _, test := range tests {
		answer := AskForConfirmation("Testing for yes", test.Input)
		assert.Equal(t, test.Output, answer)
	}
}
