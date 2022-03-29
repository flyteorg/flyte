package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const AllowedExecutionIDStartCharStr = "abcdefghijklmnopqrstuvwxyz"
const AllowedExecutionIDStr = "abcdefghijklmnopqrstuvwxyz1234567890"

var AllowedExecutionIDStartChars = []rune(AllowedExecutionIDStartCharStr)
var AllowedExecutionIDChars = []rune(AllowedExecutionIDStr)

func TestGetExecutionName(t *testing.T) {
	randString := GetExecutionName(time.Now().UnixNano())
	assert.Len(t, randString, ExecutionIDLength)
	assert.Contains(t, AllowedExecutionIDStartChars, rune(randString[0]))
	for i := 1; i < len(randString); i++ {
		assert.Contains(t, AllowedExecutionIDChars, rune(randString[i]))
	}
}
