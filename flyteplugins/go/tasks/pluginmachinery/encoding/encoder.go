package encoding

import (
	"encoding/base32"
	"fmt"
	"hash"
	"hash/fnv"
	"strings"
)

const specialEncoderKey = "abcdefghijklmnopqrstuvwxyz123456"

var Base32Encoder = base32.NewEncoding(specialEncoderKey).WithPadding(base32.NoPadding)

// Algorithm defines an enum for the encoding algorithm to use.
type Algorithm uint32

const (
	// Algorithm32 uses fnv32 bit encoder.
	Algorithm32 Algorithm = iota

	// Algorithm64 uses fnv64 bit encoder.
	Algorithm64

	// Algorithm128 uses fnv128 bit encoder.
	Algorithm128
)

type Option interface {
	option()
}

// AlgorithmOption defines a wrapper to pass the algorithm to encoding functions.
type AlgorithmOption struct {
	Option
	algo Algorithm
}

// NewAlgorithmOption wraps the Algorithm into an AlgorithmOption to pass to the encoding functions.
func NewAlgorithmOption(algo Algorithm) AlgorithmOption {
	return AlgorithmOption{
		algo: algo,
	}
}

// FixedLengthUniqueID creates a new UniqueID that is based on the inputID and of a specified length, if the given id is
// longer than the maxLength.
func FixedLengthUniqueID(inputID string, maxLength int, options ...Option) (string, error) {
	if len(inputID) <= maxLength {
		return inputID, nil
	}

	var hasher hash.Hash
	for _, option := range options {
		if algoOption, casted := option.(AlgorithmOption); casted {
			switch algoOption.algo {
			case Algorithm32:
				hasher = fnv.New32a()
			case Algorithm64:
				hasher = fnv.New64a()
			case Algorithm128:
				hasher = fnv.New128a()
			}
		}
	}

	if hasher == nil {
		hasher = fnv.New32a()
	}

	// Using 32a/64a an error can never happen, so this will always remain not covered by a unit test
	_, _ = hasher.Write([]byte(inputID)) // #nosec
	b := hasher.Sum(nil)

	// Encoding Length Calculation:
	// Base32 Encoder will encode every 5 bits into an output character (2 ^ 5 = 32)
	// output length = ciel(<input bits> / 5)
	//  for 32a hashing	= ceil(32 / 5) = 7
	//  for 64a hashing	= ceil(64 / 5) = 13
	// We prefix with character `f` so the final output is 8 or 14

	finalStr := "f" + Base32Encoder.EncodeToString(b)
	if len(finalStr) > maxLength {
		return finalStr, fmt.Errorf("max Length is too small, cannot create an encoded string that is so small")
	}
	return finalStr, nil
}

// FixedLengthUniqueIDForParts creates a new uniqueID using the parts concatenated using `-` and ensures that the
// uniqueID is not longer than the maxLength. In case a simple concatenation yields a longer string, a new hashed ID is
// created which is always around 8 characters in length.
func FixedLengthUniqueIDForParts(maxLength int, parts []string, options ...Option) (string, error) {
	b := strings.Builder{}
	for i, p := range parts {
		if i > 0 && b.Len() > 0 {
			// Ignoring the error as it always returns a nil error
			_, _ = b.WriteRune('-') // #nosec
		}

		// Ignoring the error as this is always nil
		_, _ = b.WriteString(p) // #nosec
	}

	return FixedLengthUniqueID(b.String(), maxLength, options...)
}
