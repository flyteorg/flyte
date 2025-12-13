package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnionMaps(t *testing.T) {
	assert.EqualValues(t, map[string]string{
		"left": "only",
	}, UnionMaps(map[string]string{
		"left": "only",
	}, nil))

	assert.EqualValues(t, map[string]string{
		"right": "only",
	}, UnionMaps(nil, map[string]string{
		"right": "only",
	}))

	assert.EqualValues(t, map[string]string{
		"left":  "val",
		"right": "val",
	}, UnionMaps(map[string]string{
		"left": "val",
	}, map[string]string{
		"right": "val",
	}))
}
