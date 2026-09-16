package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratorRetainsImportedTypesAndUnexportedDefaults(t *testing.T) {
	generator, err := NewGenerator(
		"github.com/flyteorg/flyte/v2/flytestdlib/cli/pflags/api/testdata/loadtarget",
		"Config", "defaultConfig", false,
	)
	require.NoError(t, err)
	provider, err := generator.Generate(t.Context())
	require.NoError(t, err)

	fields := make(map[string]FieldInfo)
	for _, field := range provider.fields {
		fields[field.Name] = field
	}
	require.Len(t, fields, 2, "the imported pflag exclusion must be preserved")
	require.Contains(t, fields, "remote.name")
	assert.Equal(t, "String", fields["remote.name"].FlagMethodName)
	assert.Equal(t, `"remote name"`, fields["remote.name"].UsageString)
	assert.Equal(t, "defaultConfig.Remote.Name", fields["remote.name"].DefaultValue)

	// Imported pointer methods must stop field traversal, and the Stringer
	// method must remain available when deriving the default value.
	require.Contains(t, fields, "remote.encoding")
	assert.Equal(t, "String", fields["remote.encoding"].FlagMethodName)
	assert.Equal(t, "defaultConfig.Remote.Encoding.String()", fields["remote.encoding"].DefaultValue)
}
