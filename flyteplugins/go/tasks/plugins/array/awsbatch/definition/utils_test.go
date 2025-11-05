package definition

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetJobDefinitionSafeName(t *testing.T) {
	testCases := [][]string{
		{"ThisIsA9_Valid--Name", "ThisIsA9_Valid--Name"},
		{"bad.name.with.special?chars", "bad-name-with-special-chars"},
		{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
	}

	for _, testCase := range testCases {
		t.Run(testCase[1], func(t *testing.T) {
			actual := GetJobDefinitionSafeName(testCase[0])
			assert.Equal(t, testCase[1], actual)
		})
	}

}
