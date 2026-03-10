package transform

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Foo struct{}

func (f Foo) Transform(data []byte) ([]byte, error) {
	return bytes.Join([][]byte{[]byte("foo transformed:"), data}, []byte(" ")), nil
}

type Bar struct{}

func (b Bar) Transform(data []byte) ([]byte, error) {
	return bytes.Join([][]byte{[]byte("bar transformed:"), data}, []byte(" ")), nil
}

func TestTransformer(t *testing.T) {
	tx := NewTransformer(Foo{}, Bar{})
	result, err := tx.Transform([]byte("test"))
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(
		t,
		result,
		[]byte("bar transformed: foo transformed: test"),
		"strings should match",
	)
}
