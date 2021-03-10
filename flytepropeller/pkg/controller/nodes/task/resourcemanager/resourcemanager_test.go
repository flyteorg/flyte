package resourcemanager

import (
	"testing"

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
