package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/flyteorg/flytestdlib/internal/utils"

	"github.com/stretchr/testify/assert"
)

func TestRegexp_MarshalJSON(t *testing.T) {
	validRegexps := []string{
		"",
		".*",
		"^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}$",
	}

	for i, validRegexp := range validRegexps {
		t.Run(fmt.Sprintf("Valid %v", i), func(t *testing.T) {
			expected := Regexp{Regexp: utils.MustCompileRegexp(validRegexp)}

			b, err := expected.MarshalJSON()
			assert.NoError(t, err)

			actual := Regexp{}
			err = actual.UnmarshalJSON(b)
			assert.NoError(t, err)

			assert.True(t, reflect.DeepEqual(expected, actual))
		})
	}
}

func TestRegexp_UnmarshalJSON(t *testing.T) {
	invalidValues := []interface{}{
		"^(",
		123,
		true,
	}
	for i, invalidRegexp := range invalidValues {
		t.Run(fmt.Sprintf("Invalid %v", i), func(t *testing.T) {
			raw, err := json.Marshal(invalidRegexp)
			assert.NoError(t, err)

			actual := Regexp{}
			err = actual.UnmarshalJSON(raw)
			assert.Error(t, err)
		})
	}

	t.Run("Empty regexp", func(t *testing.T) {
		expected := Regexp{}

		actual := Regexp{}
		err := actual.UnmarshalJSON([]byte{})
		assert.NoError(t, err)
		assert.True(t, reflect.DeepEqual(expected, actual))
	})
}
