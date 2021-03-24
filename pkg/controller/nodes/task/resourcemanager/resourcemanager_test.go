package resourcemanager

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	core2 "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	rmConfig "github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/resourcemanager/config"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

func TestComposeTokenPrefix(t *testing.T) {
	type args struct {
		id *core.TaskExecutionIdentifier
	}
	tests := []struct {
		name string
		args args
		want TokenPrefix
	}{
		{name: "composing the prefix from task execution id",
			args: args{id: &core.TaskExecutionIdentifier{
				TaskId:          nil,
				NodeExecutionId: &core.NodeExecutionIdentifier{ExecutionId: &core.WorkflowExecutionIdentifier{Project: "testproject", Domain: "testdomain", Name: "testname"}},
			}},
			want: TokenPrefix("ex:testproject:testdomain:testname"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ComposeTokenPrefix(tt.args.id); got != tt.want {
				t.Errorf("ComposeTokenPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToken_prepend(t *testing.T) {
	type args struct {
		prefix TokenPrefix
	}
	tests := []struct {
		name string
		t    Token
		args args
		want Token
	}{
		{name: "Prepend should prepend", t: Token("abcdefg-hijklmn"), args: args{prefix: TokenPrefix("tok:prefix")},
			want: Token("tok:prefix-abcdefg-hijklmn")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.prepend(tt.args.prefix); got != tt.want {
				t.Errorf("prepend() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaskResourceManager(t *testing.T) {
	rmBuilder, _ := GetResourceManagerBuilderByType(context.TODO(), rmConfig.TypeNoop, promutils.NewTestScope())
	rm, _ := rmBuilder.BuildResourceManager(context.TODO())
	taskResourceManager := GetTaskResourceManager(rm, "namespace", &core.TaskExecutionIdentifier{})
	_, err := taskResourceManager.AllocateResource(context.TODO(), "namespace", "allocation token", core2.ResourceConstraintsSpec{})
	assert.NoError(t, err)
	resourcePoolInfo := taskResourceManager.GetResourcePoolInfo()
	assert.EqualValues(t, []*event.ResourcePoolInfo{
		{
			Namespace:       "namespace",
			AllocationToken: "allocation token",
		},
	}, resourcePoolInfo)
}
