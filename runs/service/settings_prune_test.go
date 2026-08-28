package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings"
)

// parseSettings builds a Settings from protojson, failing the test on bad input.
func parseSettings(t *testing.T, in string) *settings.Settings {
	t.Helper()
	s := &settings.Settings{}
	require.NoError(t, protojson.Unmarshal([]byte(in), s))
	return s
}

func TestPruneSettings(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "empty leaf message is removed",
			in:   `{"run":{"defaultQueue":{}}}`,
			want: `{}`,
		},
		{
			name: "empty group is removed",
			in:   `{"run":{}}`,
			want: `{}`,
		},
		{
			name: "nested empty messages are removed at every depth",
			in:   `{"taskResource":{"min":{"cpu":{}},"max":{}}}`,
			want: `{}`,
		},
		{
			name: "map setting with no entries is removed",
			in:   `{"environmentVariables":{"mapValue":{"entries":{}}}}`,
			want: `{}`,
		},
		{
			name: "explicit UNSET is kept",
			in:   `{"run":{"defaultQueue":{"state":"SETTING_STATE_UNSET"}}}`,
			want: `{"run":{"defaultQueue":{"state":"SETTING_STATE_UNSET"}}}`,
		},
		{
			name: "value is kept while its empty sibling is removed",
			in:   `{"run":{"defaultQueue":{"state":"SETTING_STATE_VALUE","stringValue":"fast-queue"},"runBaseDir":{}}}`,
			want: `{"run":{"defaultQueue":{"state":"SETTING_STATE_VALUE","stringValue":"fast-queue"}}}`,
		},
		{
			name: "settings with no empty messages are unchanged",
			in:   `{"environmentVariables":{"state":"SETTING_STATE_VALUE","mapValue":{"entries":{"LOG_LEVEL":"debug"}}}}`,
			want: `{"environmentVariables":{"state":"SETTING_STATE_VALUE","mapValue":{"entries":{"LOG_LEVEL":"debug"}}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSettings(t, tt.in)
			pruneSettings(got)

			want := parseSettings(t, tt.want)
			assert.True(t, proto.Equal(want, got), "got %s", protojson.Format(got))
		})
	}
}

func TestPruneSettings_NilDoesNotPanic(t *testing.T) {
	pruneSettings(nil)
}
