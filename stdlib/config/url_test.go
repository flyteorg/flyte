package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/flyteorg/flytestdlib/internal/utils"

	"github.com/stretchr/testify/assert"
)

func TestURL_MarshalJSON(t *testing.T) {
	validURLs := []string{
		"http://localhost:123",
		"http://localhost",
		"https://non-existent.com/path/to/something",
	}

	for i, validURL := range validURLs {
		t.Run(fmt.Sprintf("Valid %v", i), func(t *testing.T) {
			expected := URL{URL: utils.MustParseURL(validURL)}

			b, err := expected.MarshalJSON()
			assert.NoError(t, err)

			actual := URL{}
			err = actual.UnmarshalJSON(b)
			assert.NoError(t, err)

			assert.True(t, reflect.DeepEqual(expected, actual))
		})
	}
}

func TestURL_UnmarshalJSON(t *testing.T) {
	invalidValues := []interface{}{
		"%gh&%ij",
		123,
		true,
	}
	for i, invalidURL := range invalidValues {
		t.Run(fmt.Sprintf("Invalid %v", i), func(t *testing.T) {
			raw, err := json.Marshal(invalidURL)
			assert.NoError(t, err)

			actual := URL{}
			err = actual.UnmarshalJSON(raw)
			assert.Error(t, err)
		})
	}

	t.Run("Invalid json", func(t *testing.T) {
		actual := URL{}
		err := actual.UnmarshalJSON([]byte{})
		assert.Error(t, err)
	})
}
