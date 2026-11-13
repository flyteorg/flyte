package controller

import (
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func taskTemplateWithQueuedTimeout(t *testing.T, queued *durationpb.Duration) []byte {
	t.Helper()
	taskTemplate := &core.TaskTemplate{}
	if err := proto.Unmarshal(validTaskAction().Spec.TaskTemplate, taskTemplate); err != nil {
		t.Fatal(err)
	}
	taskTemplate.Metadata = &core.TaskMetadata{
		Timeout:  durationpb.New(time.Minute),
		Timeouts: &core.TimeoutStrategy{QueuedTimeout: queued, Deadline: durationpb.New(time.Hour)},
	}
	data, err := proto.Marshal(taskTemplate)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestTimeoutsFromTaskTemplateQueuedTimeout(t *testing.T) {
	t.Run("reads the queued timeout next to the max runtime and the deadline", func(t *testing.T) {
		timeouts, err := timeoutsFromTaskTemplate(taskTemplateWithQueuedTimeout(t, durationpb.New(5*time.Second)))
		if err != nil {
			t.Fatal(err)
		}
		want := taskTimeouts{queued: 5 * time.Second, maxRuntime: time.Minute, deadline: time.Hour}
		if timeouts != want {
			t.Fatalf("got %+v, want %+v", timeouts, want)
		}
	})

	for _, tc := range []struct {
		name   string
		queued *durationpb.Duration
	}{
		{name: "below zero", queued: durationpb.New(-time.Second)},
		{name: "out of range nanos", queued: &durationpb.Duration{Seconds: 1, Nanos: 1_000_000_000}},
	} {
		t.Run("rejects a "+tc.name+" queued timeout", func(t *testing.T) {
			ta := validTaskAction()
			ta.Spec.TaskTemplate = taskTemplateWithQueuedTimeout(t, tc.queued)
			_, _, reason, err := validateTaskAction(ta, &mockPluginResolver{plugin: mockPlugin{}})
			if err == nil || !strings.Contains(err.Error(), "invalid queued timeout") {
				t.Fatalf("expected an invalid queued timeout error, got %v", err)
			}
			if reason != flyteorgv1.ConditionReasonInvalidSpec {
				t.Fatalf("expected reason %q, got %q", flyteorgv1.ConditionReasonInvalidSpec, reason)
			}
		})
	}
}
