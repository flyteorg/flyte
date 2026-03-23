package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutputRefs(t *testing.T) {
	t.Run("empty base returns nil", func(t *testing.T) {
		assert.Nil(t, outputRefs("", "action-0"))
	})

	t.Run("appends action name and outputs.pb", func(t *testing.T) {
		refs := outputRefs("s3://flyte-data/run123", "a0")
		assert.Equal(t, "s3://flyte-data/run123/a0/outputs.pb", refs.GetOutputUri())
	})

	t.Run("trims trailing slash from base", func(t *testing.T) {
		refs := outputRefs("s3://flyte-data/run123/", "a0")
		assert.Equal(t, "s3://flyte-data/run123/a0/outputs.pb", refs.GetOutputUri())
	})
}
