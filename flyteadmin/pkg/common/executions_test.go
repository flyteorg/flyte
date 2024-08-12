package common

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const AllowedExecutionIDAlphabetStr = "abcdefghijklmnopqrstuvwxyz"
const AllowedExecutionIDAlphanumericStr = "abcdefghijklmnopqrstuvwxyz1234567890"
const AllowedExecutionIDHumanHashStr = "abcdefghijklmnopqrstuvwxyz-"

var AllowedExecutionIDAlphabets = []rune(AllowedExecutionIDAlphabetStr)
var AllowedExecutionIDAlphanumerics = []rune(AllowedExecutionIDAlphanumericStr)
var AllowedExecutionIDHumanHashChars = []rune(AllowedExecutionIDHumanHashStr)

func TestGetExecutionName(t *testing.T) {
	randString, err := GetExecutionName(time.Now().UnixNano(), false)
	if err != nil {
		t.Errorf("unexpected error generating execution name with random string style: %v", err)
	}
	assert.Len(t, randString, ExecutionIDLength)
	assert.Contains(t, AllowedExecutionIDAlphabets, rune(randString[0]))
	for i := 1; i < len(randString); i++ {
		assert.Contains(t, AllowedExecutionIDAlphanumerics, rune(randString[i]))
	}
}

func TestGetExecutionName_HumanHash(t *testing.T) {
	randString, err := GetExecutionName(time.Now().UnixNano(), true)
	if err != nil {
		t.Errorf("unexpected error generating execution name with humanhash style: %v", err)
	}
	assert.LessOrEqual(t, len(randString), ExecutionIDLengthLimit)
	for i := 0; i < len(randString); i++ {
		assert.Contains(t, AllowedExecutionIDHumanHashChars, rune(randString[i]))
	}
	hyphenCount := strings.Count(randString, "-")
	assert.Equal(t, 2, hyphenCount, "HumanHash should contain exactly two hyphens")
	words := strings.Split(randString, "-")
	assert.Equal(t, 3, len(words), "HumanHash should be split into exactly three words")
}
