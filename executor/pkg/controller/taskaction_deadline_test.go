package controller

import (
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func taskTemplateWithTimeouts(t *testing.T, timeout, deadline *durationpb.Duration) []byte {
	t.Helper()
	metadata := &core.TaskMetadata{Timeout: timeout}
	if deadline != nil {
		metadata.Timeouts = &core.TimeoutStrategy{Deadline: deadline}
	}
	taskTemplate, err := proto.Marshal(&core.TaskTemplate{Type: "container", Metadata: metadata})
	if err != nil {
		t.Fatal(err)
	}
	return taskTemplate
}

func TestTimeoutsFromTaskTemplate(t *testing.T) {
	t.Run("reads max runtime and deadline independently", func(t *testing.T) {
		timeouts, err := timeoutsFromTaskTemplate(
			taskTemplateWithTimeouts(t, durationpb.New(time.Minute), durationpb.New(time.Hour)),
		)
		if err != nil {
			t.Fatal(err)
		}
		if timeouts != (taskTimeouts{maxRuntime: time.Minute, deadline: time.Hour}) {
			t.Fatalf("got %+v", timeouts)
		}
	})

	t.Run("does not treat the deadline as a max runtime", func(t *testing.T) {
		timeouts, err := timeoutsFromTaskTemplate(taskTemplateWithTimeouts(t, nil, durationpb.New(time.Hour)))
		if err != nil {
			t.Fatal(err)
		}
		if timeouts != (taskTimeouts{deadline: time.Hour}) {
			t.Fatalf("got %+v", timeouts)
		}
	})

	for _, tc := range []struct {
		name     string
		deadline *durationpb.Duration
	}{
		{name: "negative", deadline: durationpb.New(-time.Second)},
		{name: "out of range nanos", deadline: &durationpb.Duration{Seconds: 1, Nanos: 1_000_000_000}},
	} {
		t.Run("rejects a "+tc.name+" deadline", func(t *testing.T) {
			_, err := timeoutsFromTaskTemplate(taskTemplateWithTimeouts(t, nil, tc.deadline))
			if err == nil || !strings.Contains(err.Error(), "invalid task deadline") {
				t.Fatalf("expected an invalid task deadline error, got %v", err)
			}

			ta := validTaskAction()
			ta.Spec.TaskTemplate = taskTemplateWithTimeouts(t, nil, tc.deadline)
			_, _, reason, err := validateTaskAction(ta, &mockPluginResolver{plugin: mockPlugin{}})
			if err == nil || reason != flyteorgv1.ConditionReasonInvalidSpec {
				t.Fatalf("expected InvalidSpec, got reason %q and error %v", reason, err)
			}
		})
	}
}

func TestEffectiveDeadline(t *testing.T) {
	created := time.Date(2026, time.October, 28, 0, 0, 0, 0, time.UTC)
	attemptStarted := metav1.NewTime(created.Add(10 * time.Second))
	newTaskAction := func(started *metav1.Time) *flyteorgv1.TaskAction {
		return &flyteorgv1.TaskAction{
			ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.NewTime(created)},
			Status:     flyteorgv1.TaskActionStatus{AttemptStartedAt: started},
		}
	}

	for _, tc := range []struct {
		name     string
		started  *metav1.Time
		timeouts taskTimeouts
		want     time.Time
		wantOK   bool
	}{
		{name: "unbounded", started: &attemptStarted},
		{name: "deadline only, before the attempt starts", timeouts: taskTimeouts{deadline: time.Minute}, want: created.Add(time.Minute), wantOK: true},
		{name: "max runtime only", started: &attemptStarted, timeouts: taskTimeouts{maxRuntime: time.Minute}, want: created.Add(70 * time.Second), wantOK: true},
		{name: "max runtime, attempt not started", timeouts: taskTimeouts{maxRuntime: time.Minute}},
		{name: "deadline first", started: &attemptStarted, timeouts: taskTimeouts{maxRuntime: time.Minute, deadline: 30 * time.Second}, want: created.Add(30 * time.Second), wantOK: true},
		{name: "max runtime first", started: &attemptStarted, timeouts: taskTimeouts{maxRuntime: time.Minute, deadline: time.Hour}, want: created.Add(70 * time.Second), wantOK: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := effectiveDeadline(newTaskAction(tc.started), tc.timeouts)
			if ok != tc.wantOK || !got.Equal(tc.want) {
				t.Fatalf("got (%v, %v), want (%v, %v)", got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestIsDeadlineTimeout(t *testing.T) {
	created := time.Date(2026, time.October, 28, 0, 0, 0, 0, time.UTC)
	timeouts := taskTimeouts{maxRuntime: time.Minute, deadline: time.Hour}
	newTaskAction := func(timeoutAt *time.Time) *flyteorgv1.TaskAction {
		ta := &flyteorgv1.TaskAction{ObjectMeta: metav1.ObjectMeta{CreationTimestamp: metav1.NewTime(created)}}
		if timeoutAt != nil {
			at := metav1.NewTime(*timeoutAt)
			ta.Status.TimeoutAt = &at
		}
		return ta
	}
	maxRuntimeEnd := created.Add(time.Minute)
	deadlineEnd := created.Add(time.Hour)

	if isDeadlineTimeout(newTaskAction(nil), timeouts) {
		t.Fatal("no pending timeout is not a deadline timeout")
	}
	if isDeadlineTimeout(newTaskAction(&maxRuntimeEnd), timeouts) {
		t.Fatal("a max-runtime timeout before the deadline is not a deadline timeout")
	}
	if !isDeadlineTimeout(newTaskAction(&deadlineEnd), timeouts) {
		t.Fatal("a timeout at the deadline is a deadline timeout")
	}
	if isDeadlineTimeout(newTaskAction(&deadlineEnd), taskTimeouts{maxRuntime: time.Minute}) {
		t.Fatal("without a deadline nothing is a deadline timeout")
	}
}
