package config

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeepCopyConfig(t *testing.T) {
	type pair struct {
		first  interface{}
		second interface{}
	}

	type fakeConfig struct {
		I   int
		S   string
		Ptr *fakeConfig
	}

	testCases := []pair{
		{3, 3},
		{"word", "word"},
		{fakeConfig{I: 4, S: "four", Ptr: &fakeConfig{I: 5, S: "five"}}, fakeConfig{I: 4, S: "four", Ptr: &fakeConfig{I: 5, S: "five"}}},
		{&fakeConfig{I: 4, S: "four", Ptr: &fakeConfig{I: 5, S: "five"}}, &fakeConfig{I: 4, S: "four", Ptr: &fakeConfig{I: 5, S: "five"}}},
	}

	for i, testCase := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			input, expected := testCase.first, testCase.second
			actual, err := DeepCopyConfig(input)
			assert.NoError(t, err)
			assert.Equal(t, reflect.TypeOf(expected).String(), reflect.TypeOf(actual).String())
			assert.Equal(t, expected, actual)
		})
	}
}
