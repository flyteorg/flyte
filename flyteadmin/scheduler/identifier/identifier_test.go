package identifier

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestGetIdentifierString(t *testing.T) {
	t.Run("with org", func(t *testing.T) {
		identifier := &core.Identifier{
			Org:     "org",
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Version: "version",
		}
		expected := "org:project:domain:name:version"
		actual := getIdentifierString(identifier)
		assert.Equal(t, expected, actual)
	})
	t.Run("without org", func(t *testing.T) {
		identifier := &core.Identifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
			Version: "version",
		}
		expected := "project:domain:name:version"
		actual := getIdentifierString(identifier)
		assert.Equal(t, expected, actual)
	})
}
