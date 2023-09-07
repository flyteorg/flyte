package aws

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/magiconair/properties/assert"
)

func TestHashIdentifier(t *testing.T) {
	identifier := core.Identifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
		Version: "ignored",
	}
	hashedValue := hashIdentifier(identifier)
	assert.Equal(t, uint64(16301494360130577061), hashedValue)
}
