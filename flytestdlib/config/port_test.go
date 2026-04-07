package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type PortTestCase struct {
	Expected Port
	Input    interface{}
}

func TestPort_MarshalJSON(t *testing.T) {
	validPorts := []PortTestCase{
		{Expected: Port{Port: 8080}, Input: 8080},
		{Expected: Port{Port: 1}, Input: 1},
		{Expected: Port{Port: 65535}, Input: "65535"},
		{Expected: Port{Port: 65535}, Input: 65535},
	}

	for i, validPort := range validPorts {
		t.Run(fmt.Sprintf("Valid %v [%v]", i, validPort.Input), func(t *testing.T) {
			b, err := json.Marshal(validPort.Input)
			assert.NoError(t, err)

			actual := Port{}
			err = actual.UnmarshalJSON(b)
			assert.NoError(t, err)

			assert.True(t, reflect.DeepEqual(validPort.Expected, actual))
		})
	}
}

func TestPort_UnmarshalJSON(t *testing.T) {
	invalidValues := []interface{}{
		"%gh&%ij",
		1000000,
		true,
	}

	for i, invalidPort := range invalidValues {
		t.Run(fmt.Sprintf("Invalid %v", i), func(t *testing.T) {
			raw, err := json.Marshal(invalidPort)
			assert.NoError(t, err)

			actual := URL{}
			err = actual.UnmarshalJSON(raw)
			assert.Error(t, err)
		})
	}

	t.Run("Invalid json", func(t *testing.T) {
		actual := Port{}
		err := actual.UnmarshalJSON([]byte{})
		assert.Error(t, err)
	})

	t.Run("Invalid Range", func(t *testing.T) {
		b, err := json.Marshal(float64(100000))
		assert.NoError(t, err)

		actual := Port{}
		err = actual.UnmarshalJSON(b)
		assert.Error(t, err)
	})

	t.Run("Unmarshal Empty", func(t *testing.T) {
		p := Port{}
		raw, err := json.Marshal(p)
		assert.NoError(t, err)

		actual := Port{}
		assert.NoError(t, actual.UnmarshalJSON(raw))
		assert.Equal(t, 0, actual.Port)
	})
}
