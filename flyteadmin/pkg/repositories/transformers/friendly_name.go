package transformers

import (
	"github.com/wolfeidau/humanhash"
	"k8s.io/apimachinery/pkg/util/rand"
)

// Length of the random string used for generating hash keys; can be any positive integer
const RandStringLength = 20

/* #nosec */
func CreateFriendlyName(seed int64) string {
	rand.Seed(seed)
	hashKey := []byte(rand.String(RandStringLength))
	// Ignoring the error as it's guaranteed hash key longer than result in this context.
	result, _ := humanhash.Humanize(hashKey, 4)
	return result
}
