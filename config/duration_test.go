package config

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDuration_MarshalJSON(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		expected := Duration{
			Duration: time.Second * 2,
		}

		b, err := expected.MarshalJSON()
		assert.NoError(t, err)

		actual := Duration{}
		err = actual.UnmarshalJSON(b)
		assert.NoError(t, err)

		assert.True(t, reflect.DeepEqual(expected, actual))
	})
}

func TestDuration_UnmarshalJSON(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		actual := Duration{}
		err := actual.UnmarshalJSON([]byte{})
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(0), actual.Duration)
	})

	t.Run("Invalid_string", func(t *testing.T) {
		input := "blah"
		raw, err := json.Marshal(input)
		assert.NoError(t, err)

		actual := Duration{}
		err = actual.UnmarshalJSON(raw)
		assert.Error(t, err)
	})

	t.Run("Valid_float", func(t *testing.T) {
		input := float64(12345)
		raw, err := json.Marshal(input)
		assert.NoError(t, err)

		actual := Duration{}
		err = actual.UnmarshalJSON(raw)
		assert.NoError(t, err)
	})

	t.Run("Invalid_bool", func(t *testing.T) {
		input := true
		raw, err := json.Marshal(input)
		assert.NoError(t, err)

		actual := Duration{}
		err = actual.UnmarshalJSON(raw)
		assert.Error(t, err)
	})
}
