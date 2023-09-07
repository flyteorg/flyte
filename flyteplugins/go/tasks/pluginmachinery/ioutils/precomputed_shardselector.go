package ioutils

import (
	"context"
	"hash/fnv"
	"strings"

	"github.com/pkg/errors"
)

// Generates the entire latin alphabet and appends it to the passed in array and returns the new array
func GenerateAlphabet(b []rune) []rune {
	for i := 'a'; i <= 'z'; i++ {
		b = append(b, i)
	}
	return b
}

// Generates all arabic numerals and appends to the passed in array and returns the new array/slice
func GenerateArabicNumerals(b []rune) []rune {
	for i := '0'; i <= '9'; i++ {
		b = append(b, i)
	}
	return b
}

func createAlphabetAndNumerals() []rune {
	b := make([]rune, 0, 36)
	b = GenerateAlphabet(b)
	return GenerateArabicNumerals(b)
}

// this sharder distributes data into one of the precomputed buckets. The bucket is deterministically determined given the input s
type PrecomputedShardSelector struct {
	precomputedPrefixes []string
	buckets             uint32
}

// Generates deterministic shard id for the given string s
func (d *PrecomputedShardSelector) GetShardPrefix(_ context.Context, s []byte) (string, error) {
	h := fnv.New32a()
	_, err := h.Write(s)
	if err != nil {
		return "", errors.Wrap(err, "failed to create shard prefix, reason hash failure.")
	}
	idx := h.Sum32() % d.buckets
	return d.precomputedPrefixes[idx], nil
}

// Creates a PrecomputedShardSelector with 36*36 unique shards. Each shard is of the format {[0-9a-z][0-9a-z]}, i.e. 2 character long.
func NewBase36PrefixShardSelector(ctx context.Context) (ShardSelector, error) {
	permittedChars := createAlphabetAndNumerals()
	n := len(permittedChars)
	precomputedPrefixes := make([]string, 0, n*n)
	for _, c1 := range permittedChars {
		for _, c2 := range permittedChars {
			sb := strings.Builder{}
			sb.WriteRune(c1)
			sb.WriteRune(c2)
			precomputedPrefixes = append(precomputedPrefixes, sb.String())
		}
	}

	return NewConstantShardSelector(precomputedPrefixes), nil
}

// uses the given shards to select a shard
func NewConstantShardSelector(shards []string) ShardSelector {
	return &PrecomputedShardSelector{
		precomputedPrefixes: shards,
		buckets:             uint32(len(shards)),
	}
}
