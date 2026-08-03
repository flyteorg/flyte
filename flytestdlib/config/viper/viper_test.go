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

func Test_restoreDottedMapKeys(t *testing.T) {
	t.Run("siblings sharing dotted prefix", func(t *testing.T) {
		viper := map[string]interface{}{
			"annotations": map[string]interface{}{
				"test": map[string]interface{}{
					"annotation": "true",
					"other":      "false",
				},
			},
		}
		raw := map[string]interface{}{
			"annotations": map[string]interface{}{
				"test.annotation": "true",
				"test.other":      "false",
			},
		}

		restoreDottedMapKeys(viper, raw)

		assert.Equal(t, map[string]interface{}{
			"annotations": map[string]interface{}{
				"test.annotation": "true",
				"test.other":      "false",
			},
		}, viper)
	})

	t.Run("case-insensitive recursion into nested map", func(t *testing.T) {
		viper := map[string]interface{}{
			"annotations": map[string]interface{}{
				"test": map[string]interface{}{"annotation": "true"},
			},
		}
		raw := map[string]interface{}{
			"Annotations": map[string]interface{}{
				"test.annotation": "true",
			},
		}

		restoreDottedMapKeys(viper, raw)

		assert.Equal(t, map[string]interface{}{
			"annotations": map[string]interface{}{
				"test.annotation": "true",
			},
		}, viper)
	})

	t.Run("dotted key with map value", func(t *testing.T) {
		viper := map[string]interface{}{
			"a": map[string]interface{}{
				"b": map[string]interface{}{
					"c": map[string]interface{}{"d": "v"},
				},
			},
		}
		raw := map[string]interface{}{
			"a.b": map[string]interface{}{
				"c.d": "v",
			},
		}

		restoreDottedMapKeys(viper, raw)

		assert.Equal(t, map[string]interface{}{
			"a.b": map[string]interface{}{
				"c.d": "v",
			},
		}, viper)
	})

	t.Run("dotted and non-dotted siblings at top level", func(t *testing.T) {
		viper := map[string]interface{}{
			"a": map[string]interface{}{
				"b": "from-dotted",
				"x": "from-non-dotted",
			},
		}
		raw := map[string]interface{}{
			"a.b": "from-dotted",
			"a":   map[string]interface{}{"x": "from-non-dotted"},
		}

		restoreDottedMapKeys(viper, raw)

		assert.Equal(t, map[string]interface{}{
			"a.b": "from-dotted",
			"a":   map[string]interface{}{"x": "from-non-dotted"},
		}, viper)
	})

	// This test case covers the usecases mentioned in issue #6166
	t.Run("k8s annotation key", func(t *testing.T) {
		viper := map[string]interface{}{
			"default-annotations": map[string]interface{}{
				"cluster-autoscaler": map[string]interface{}{
					"kubernetes": map[string]interface{}{
						"io/safe-to-evict": "false",
					},
				},
			},
		}
		raw := map[string]interface{}{
			"default-annotations": map[string]interface{}{
				"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
			},
		}

		restoreDottedMapKeys(viper, raw)

		assert.Equal(t, map[string]interface{}{
			"default-annotations": map[string]interface{}{
				"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
			},
		}, viper)
	})
}

func Test_pruneSplitPath(t *testing.T) {
	t.Run("preserves sibling under shared prefix", func(t *testing.T) {
		m := map[string]interface{}{
			"test": map[string]interface{}{
				"annotation": "v1",
				"other":      "v2",
			},
		}

		pruneSplitPath(m, []string{"test", "annotation"})

		assert.Equal(t, map[string]interface{}{
			"test": map[string]interface{}{"other": "v2"},
		}, m)
	})

	t.Run("collapses entire chain when emptied", func(t *testing.T) {
		m := map[string]interface{}{
			"a": map[string]interface{}{
				"b": map[string]interface{}{"c": "v"},
			},
		}

		pruneSplitPath(m, []string{"a", "b", "c"})

		assert.Equal(t, map[string]interface{}{}, m)
	})

	t.Run("path diverges from viper structure", func(t *testing.T) {
		m := map[string]interface{}{"a": "scalar"}

		pruneSplitPath(m, []string{"a", "b"})

		assert.Equal(t, map[string]interface{}{"a": "scalar"}, m)
	})
}

func Test_findViperKey(t *testing.T) {
	t.Run("case-insensitive match", func(t *testing.T) {
		k, ok := findViperKey(map[string]interface{}{"foo": 1}, "Foo")
		assert.True(t, ok)
		assert.Equal(t, "foo", k)
	})

	t.Run("not found", func(t *testing.T) {
		_, ok := findViperKey(map[string]interface{}{"foo": 1}, "bar")
		assert.False(t, ok)
	})
}
