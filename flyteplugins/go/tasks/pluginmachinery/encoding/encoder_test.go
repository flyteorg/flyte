package encoding

import (
	"hash"
	"hash/fnv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixedLengthUniqueID(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		maxLength   int
		output      string
		expectError bool
	}{
		{"smallerThanMax", "x", 5, "x", false},
		{"veryLowLimit", "xx", 1, "flfryc2i", true},
		{"highLimit", "xxxxxx", 5, "fufiti6i", true},
		{"higherLimit", "xxxxx", 10, "xxxxx", false},
		{"largeID", "xxxxxxxxxxxxxxxxxxxxxxxx", 20, "fuaa3aji", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			i, err := FixedLengthUniqueID(test.input, test.maxLength)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, test.output, i)
		})
	}
}

func TestFixedLengthUniqueIDForParts(t *testing.T) {
	tests := []struct {
		name        string
		parts       []string
		maxLength   int
		algorithm   Algorithm
		output      string
		expectError bool
	}{
		{"smallerThanMax", []string{"x", "y", "z"}, 10, Algorithm32, "x-y-z", false},
		{"veryLowLimit", []string{"x", "y"}, 1, Algorithm32, "fz2jizji", true},
		{"fittingID", []string{"x"}, 2, Algorithm32, "x", false},
		{"highLimit", []string{"x", "y", "z"}, 4, Algorithm32, "fxzsoqrq", true},
		{"largeID", []string{"x", "y", "z", "m", "n", "y", "z", "m", "n", "y", "z", "m", "n"}, 15, Algorithm32, "fe63sz6y", false},
		{"largeID", []string{"x", "y", "z", "m", "n", "y", "z", "m", "n", "y", "z", "m", "n"}, 15, Algorithm64, "fwp4bky2kucex5", false},
		{"largeID", []string{"x", "y", "z", "m", "n", "y", "z", "m", "n", "y", "z", "m", "n", "z", "m", "n", "y", "z", "m", "n"}, 30, Algorithm128, "fbmesl15enghpjepzjm5cp1zfqe", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			i, err := FixedLengthUniqueIDForParts(test.maxLength, test.parts, NewAlgorithmOption(test.algorithm))
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.output, i)
		})
	}
}

func benchmarkKB(b *testing.B, h hash.Hash) {
	b.SetBytes(1024)
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i)
	}

	in := make([]byte, 0, h.Size())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Reset()
		h.Write(data)
		h.Sum(in)
	}
}

// Documented Results:
// goos: darwin
// goarch: amd64
// pkg: github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/encoding
// cpu: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz
// BenchmarkFixedLengthUniqueID
// BenchmarkFixedLengthUniqueID/New32a
// BenchmarkFixedLengthUniqueID/New32a-16         	 1000000	      1088 ns/op	 941.25 MB/s
// BenchmarkFixedLengthUniqueID/New64a
// BenchmarkFixedLengthUniqueID/New64a-16         	 1239402	       951.3 ns/op	1076.39 MB/s
func BenchmarkFixedLengthUniqueID(b *testing.B) {
	b.Run("New32a", func(b *testing.B) {
		benchmarkKB(b, fnv.New32a())
	})

	b.Run("New64a", func(b *testing.B) {
		benchmarkKB(b, fnv.New64a())
	})
}
