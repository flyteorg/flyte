package viper

import (
	"encoding/base64"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_stringToByteArray(t *testing.T) {
	t.Run("Expected types", func(t *testing.T) {
		input := "hello world"
		base64Encoded := base64.StdEncoding.EncodeToString([]byte(input))
		res, err := stringToByteArray(reflect.TypeOf(base64Encoded), reflect.TypeOf([]byte{}), base64Encoded)
		assert.NoError(t, err)
		assert.Equal(t, []byte(input), res)
	})

	t.Run("Expected types - array string", func(t *testing.T) {
		input := []string{"hello world"}
		base64Encoded := base64.StdEncoding.EncodeToString([]byte(input[0]))
		res, err := stringToByteArray(reflect.TypeOf(input), reflect.TypeOf([]byte{}), []string{base64Encoded})
		assert.NoError(t, err)
		assert.Equal(t, []byte(input[0]), res)
	})

	t.Run("Expected types - invalid encoding", func(t *testing.T) {
		input := []string{"hello world"}
		_, err := stringToByteArray(reflect.TypeOf(input), reflect.TypeOf([]byte{}), []string{"invalid base64"})
		assert.Error(t, err)
	})

	t.Run("Expected types - empty array string", func(t *testing.T) {
		input := []string{"hello world"}
		res, err := stringToByteArray(reflect.TypeOf(input), reflect.TypeOf([]byte{}), []string{})
		assert.NoError(t, err)
		assert.Equal(t, []string{}, res)
	})

	t.Run("Unexpected types", func(t *testing.T) {
		input := 5
		res, err := stringToByteArray(reflect.TypeOf(input), reflect.TypeOf([]byte{}), input)
		assert.NoError(t, err)
		assert.NotEqual(t, []byte("hello"), res)
	})

	t.Run("Unexpected types", func(t *testing.T) {
		input := 5
		res, err := stringToByteArray(reflect.TypeOf(input), reflect.TypeOf(""), input)
		assert.NoError(t, err)
		assert.NotEqual(t, []byte("hello"), res)
	})
}
