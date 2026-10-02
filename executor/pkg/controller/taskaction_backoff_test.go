package controller

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	testingclock "k8s.io/utils/clock/testing"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func taskTemplateWithBackoff(t *testing.T, backoff *core.Backoff) []byte {
	t.Helper()
	taskTemplate, err := proto.Marshal(&core.TaskTemplate{
		Type:     "container",
		Metadata: &core.TaskMetadata{Retries: &core.RetryStrategy{Retries: 2, Backoff: backoff}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return taskTemplate
}

func TestValidateRetryBackoff(t *testing.T) {
	if err := validateRetryBackoff(nil); err != nil {
		t.Fatalf("expected no backoff to be accepted, got: %v", err)
	}
	if err := validateRetryBackoff(&core.Backoff{
		Base:   durationpb.New(10 * time.Second),
		Factor: proto.Float64(2),
		Cap:    durationpb.New(10 * time.Minute),
	}); err != nil {
		t.Fatalf("expected a capped exponential backoff to be accepted, got: %v", err)
	}

	cases := []struct {
		name    string
		backoff *core.Backoff
	}{
		{"factor below 1", &core.Backoff{Base: durationpb.New(time.Second), Factor: proto.Float64(0.5)}},
		{"factor NaN", &core.Backoff{Base: durationpb.New(time.Second), Factor: proto.Float64(math.NaN())}},
		{"factor infinite", &core.Backoff{Base: durationpb.New(time.Second), Factor: proto.Float64(math.Inf(1))}},
		{"factor above 1 without cap", &core.Backoff{Base: durationpb.New(time.Second), Factor: proto.Float64(2)}},
		{"negative base", &core.Backoff{Base: durationpb.New(-time.Second)}},
		{"negative cap", &core.Backoff{Base: durationpb.New(time.Second), Cap: durationpb.New(-time.Second)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRetryBackoff(tc.backoff)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "backoff") {
				t.Errorf("expected error to mention backoff, got: %v", err)
			}
		})
	}
}

func TestScheduleNextAttempt(t *testing.T) {
	now := time.Date(2026, time.August, 25, 0, 0, 0, 0, time.UTC)
	r := &TaskActionReconciler{Clock: testingclock.NewFakeClock(now)}
	valid := &core.Backoff{Base: durationpb.New(10 * time.Second), Factor: proto.Float64(2), Cap: durationpb.New(10 * time.Minute)}
	uncapped := &core.Backoff{Base: durationpb.New(10 * time.Second), Factor: proto.Float64(2)}

	cases := []struct {
		name           string
		backoff        *core.Backoff
		failedAttempts uint32
		want           *time.Time
	}{
		{"no backoff launches at once", nil, 1, nil},
		{"first retry waits the base", valid, 1, ptr(now.Add(10 * time.Second))},
		{"third retry waits base * factor**2", valid, 3, ptr(now.Add(40 * time.Second))},
		{"a backoff the contract rejects launches at once", uncapped, 1, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ta := &flyteorgv1.TaskAction{Spec: flyteorgv1.TaskActionSpec{TaskTemplate: taskTemplateWithBackoff(t, tc.backoff)}}
			r.scheduleNextAttempt(context.Background(), ta, tc.failedAttempts)
			switch {
			case tc.want == nil && ta.Status.NextAttemptAt != nil:
				t.Fatalf("expected no NextAttemptAt, got %v", ta.Status.NextAttemptAt.Time)
			case tc.want != nil && (ta.Status.NextAttemptAt == nil || !ta.Status.NextAttemptAt.Time.Equal(*tc.want)):
				t.Fatalf("expected NextAttemptAt %v, got %v", *tc.want, ta.Status.NextAttemptAt)
			}
		})
	}
}

func ptr[T any](v T) *T { return &v }

func TestRetryBackoffDelay(t *testing.T) {
	exponential := &core.Backoff{
		Base:   durationpb.New(10 * time.Second),
		Factor: proto.Float64(2),
		Cap:    durationpb.New(10 * time.Minute),
	}
	tests := []struct {
		name    string
		backoff *core.Backoff
		retry   uint32
		want    time.Duration
	}{
		{name: "no backoff", backoff: nil, retry: 0, want: 0},
		{name: "no base", backoff: &core.Backoff{Factor: proto.Float64(2)}, retry: 3, want: 0},
		{name: "first retry waits the base", backoff: exponential, retry: 0, want: 10 * time.Second},
		{name: "second retry doubles", backoff: exponential, retry: 1, want: 20 * time.Second},
		{name: "fifth retry", backoff: exponential, retry: 4, want: 160 * time.Second},
		{name: "capped", backoff: exponential, retry: 6, want: 10 * time.Minute},
		{name: "huge retry index stays capped", backoff: exponential, retry: 100, want: 10 * time.Minute},
		{
			name:    "no factor is constant",
			backoff: &core.Backoff{Base: durationpb.New(30 * time.Second)},
			retry:   5,
			want:    30 * time.Second,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := retryBackoffDelay(tc.backoff, tc.retry); got != tc.want {
				t.Fatalf("retryBackoffDelay(%v, %d) = %s, want %s", tc.backoff, tc.retry, got, tc.want)
			}
		})
	}
}
