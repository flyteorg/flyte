package encoding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixedLengthUniqueID(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		maxLength   int
		output      string
		expectError bool
	}{
		{"smallerThanMax", "x", 5, "x", false},
		{"veryLowLimit", "xx", 1, "flfryc2i", true},
		{"highLimit", "xxxxxx", 5, "fufiti6i", true},
		{"higherLimit", "xxxxx", 10, "xxxxx", false},
		{"largeID", "xxxxxxxxxxx", 10, "fggddjly", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			i, err := FixedLengthUniqueID(test.input, test.maxLength)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, i, test.output)
		})
	}
}

func TestFixedLengthUniqueIDForParts(t *testing.T) {
	tests := []struct {
		name        string
		parts       []string
		maxLength   int
		output      string
		expectError bool
	}{
		{"smallerThanMax", []string{"x", "y", "z"}, 10, "x-y-z", false},
		{"veryLowLimit", []string{"x", "y"}, 1, "fz2jizji", true},
		{"fittingID", []string{"x"}, 2, "x", false},
		{"highLimit", []string{"x", "y", "z"}, 4, "fxzsoqrq", true},
		{"largeID", []string{"x", "y", "z", "m", "n"}, 8, "fsigbmty", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			i, err := FixedLengthUniqueIDForParts(test.maxLength, test.parts...)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, i, test.output)
		})
	}
}
