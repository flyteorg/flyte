package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetExecutionName(t *testing.T) {
	randString := GetExecutionName(time.Now().UnixNano())
	assert.Len(t, randString, ExecutionIDLength)
	assert.Contains(t, AllowedExecutionIDStartChars, rune(randString[0]))
	for i := 1; i < len(randString); i++ {
		assert.Contains(t, AllowedExecutionIDChars, rune(randString[i]))
	}
}
