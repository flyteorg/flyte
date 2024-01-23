package nodes

import (
	"os"
	"testing"

	protoV1 "github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/proto"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

func TestCreateAliasMap(t *testing.T) {
	{
		aliases := []v1alpha1.Alias{
			{Alias: core.Alias{Var: "x", Alias: "y"}},
		}
		m := CreateAliasMap(aliases)
		assert.Equal(t, map[string]string{
			"y": "x",
		}, m)
	}
	{
		var aliases []v1alpha1.Alias
		m := CreateAliasMap(aliases)
		assert.Equal(t, map[string]string{}, m)
	}
	{
		m := CreateAliasMap(nil)
		assert.Equal(t, map[string]string{}, m)
	}
}

func TestOutputTemp(t *testing.T) {
	f, err := os.ReadFile("/Users/haytham/Downloads/outputs.pb")
	assert.NoError(t, err)

	msg := []protoV1.Message{&core.OutputData{}, &core.LiteralMap{}}
	var lastErr error
	var index int
	for i, m := range msg {
		err = proto.UnmarshalOptions{DiscardUnknown: false, AllowPartial: false}.Unmarshal(f, protoV1.MessageV2(m))
		if err != nil {
			lastErr = err
			continue
		}

		if len(msg) == 1 || len(protoV1.MessageV2(m).ProtoReflect().GetUnknown()) == 0 {
			index = i
			lastErr = nil
			break
		}
	}

	assert.NoError(t, lastErr)
	t.Log(index)
	t.Log(msg[index])
	t.FailNow()
}
