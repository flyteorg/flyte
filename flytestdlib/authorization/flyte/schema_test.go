package flyte

import (
	"os"
	"testing"

	cedar "github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/x/exp/schema"
	"github.com/stretchr/testify/require"
)

func TestPublishedModel(t *testing.T) {
	raw, err := os.ReadFile("../../../docs/authorization/schema.cedarschema")
	require.NoError(t, err)
	var s schema.Schema
	require.NoError(t, s.UnmarshalCedar(raw))
	model, err := s.Resolve()
	require.NoError(t, err)
	for _, r := range rules {
		_, ok := model.Actions[cedar.NewEntityUID("Flyte::Action", cedar.String(r.action))]
		require.True(t, ok, "missing schema action %s", r.action)
	}
	for _, action := range []string{"StorageOverride", "TaskExecute", "TriggerExecute"} {
		_, ok := model.Actions[cedar.NewEntityUID("Flyte::Action", cedar.String(action))]
		require.True(t, ok, "missing schema action %s", action)
	}
	raw, err = os.ReadFile("../../../docs/authorization/example.cedar")
	require.NoError(t, err)
	_, err = cedar.NewPolicySetFromBytes("example.cedar", raw)
	require.NoError(t, err)
}
