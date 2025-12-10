package codex

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGobStateCodec(t *testing.T) {
	g := GobStateCodec{}

	type sample struct {
		A int
		B string
	}

	type sample2 struct {
		A int
		B string
		C int
	}

	type sample3 struct {
		B string
		C int
	}

	t.Run("symmetric", func(t *testing.T) {
		b := bytes.NewBuffer([]byte{})
		s := &sample{A: 10, B: "hello"}
		assert.NoError(t, g.Encode(&s, b))

		s2 := &sample{}
		assert.NoError(t, g.Decode(b, s2))
		assert.Equal(t, s2.A, s.A)
		assert.Equal(t, s2.B, s.B)
	})

	t.Run("updated", func(t *testing.T) {
		b := bytes.NewBuffer([]byte{})
		s := &sample{A: 10, B: "hello"}
		assert.NoError(t, g.Encode(&s, b))

		s2 := &sample2{}
		assert.NoError(t, g.Decode(b, s2))
		assert.Equal(t, s2.A, s.A)
		assert.Equal(t, s2.B, s.B)
		assert.Equal(t, s2.C, 0)
	})

	t.Run("bad-update", func(t *testing.T) {
		b := bytes.NewBuffer([]byte{})
		s := &sample{A: 10, B: "hello"}
		assert.NoError(t, g.Encode(&s, b))

		s3 := &sample3{}
		assert.NoError(t, g.Decode(b, s3))
		assert.Equal(t, s3.B, s.B)
		assert.Equal(t, s3.C, 0)
	})
}
