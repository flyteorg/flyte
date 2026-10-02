package controller

import (
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

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
