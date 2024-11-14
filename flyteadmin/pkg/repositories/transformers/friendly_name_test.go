package transformers

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const AllowedFriendlyNameStr = "abcdefghijklmnopqrstuvwxyz-"
const FriendlyNameLengthLimit = 255

var AllowedFriendlyNameChars = []rune(AllowedFriendlyNameStr)

func TestCreateFriendlyName(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		randString := CreateFriendlyName(time.Now().UnixNano())
		assert.LessOrEqual(t, len(randString), FriendlyNameLengthLimit)
		for i := 0; i < len(randString); i++ {
			assert.Contains(t, AllowedFriendlyNameChars, rune(randString[i]))
		}
		hyphenCount := strings.Count(randString, "-")
		assert.Equal(t, 3, hyphenCount, "FriendlyName should contain exactly three hyphens")
		words := strings.Split(randString, "-")
		assert.Equal(t, 4, len(words), "FriendlyName should be split into exactly four words")
	})

}
