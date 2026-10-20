package gpufault

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRelevantToFailure(t *testing.T) {
	failedAt := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		activeFrom  time.Time
		activeUntil time.Time
		want        bool
	}{
		{
			name:        "a fault recorded once, just before the failure",
			activeFrom:  failedAt.Add(-time.Minute),
			activeUntil: failedAt.Add(-time.Minute),
			want:        true,
		},
		{
			// Dying hardware goes on faulting after the container is gone. Judging this
			// by its last observation alone would discard it, and the longer it kept
			// faulting the more certainly it would be discarded.
			name:        "a fault that started before the failure and kept firing well after it",
			activeFrom:  failedAt.Add(-time.Minute),
			activeUntil: failedAt.Add(11 * time.Minute),
			want:        true,
		},
		{
			// Judging this by its first recording alone would discard it.
			name:        "a fault older than the window but still firing at the failure",
			activeFrom:  failedAt.Add(-40 * time.Minute),
			activeUntil: failedAt,
			want:        true,
		},
		{
			name:        "a fault spanning the whole window and beyond in both directions",
			activeFrom:  failedAt.Add(-10 * time.Hour),
			activeUntil: failedAt.Add(10 * time.Hour),
			want:        true,
		},
		{
			name:        "a fault first recorded inside the slack",
			activeFrom:  failedAt.Add(AfterFailureSlack - time.Second),
			activeUntil: failedAt.Add(time.Hour),
			want:        true,
		},
		{
			name:        "a fault that only started after the slack",
			activeFrom:  failedAt.Add(AfterFailureSlack + time.Second),
			activeUntil: failedAt.Add(time.Hour),
			want:        false,
		},
		{
			name:        "a fault that stopped firing just before the window opened",
			activeFrom:  failedAt.Add(-90 * time.Minute),
			activeUntil: failedAt.Add(-RelevanceWindow - time.Second),
			want:        false,
		},
		{
			name:        "a fault still firing exactly as the window opens",
			activeFrom:  failedAt.Add(-90 * time.Minute),
			activeUntil: failedAt.Add(-RelevanceWindow),
			want:        true,
		},
		{
			// A fault recorded once rather than aggregated carries no last observation,
			// and is judged as the moment it is.
			name:       "only a first recording, inside the window",
			activeFrom: failedAt.Add(-time.Minute),
			want:       true,
		},
		{
			name:       "only a first recording, outside the window",
			activeFrom: failedAt.Add(-2 * RelevanceWindow),
			want:       false,
		},
		{
			name:        "only a last observation, inside the window",
			activeUntil: failedAt.Add(-time.Minute),
			want:        true,
		},
		{
			name:        "only a last observation, outside the window",
			activeUntil: failedAt.Add(-2 * RelevanceWindow),
			want:        false,
		},
		{
			name: "no times at all explains nothing",
			want: false,
		},
		{
			// No honest recorder produces this, so nothing can be concluded from it.
			name:        "a last observation older than the first recording explains nothing",
			activeFrom:  failedAt,
			activeUntil: failedAt.Add(-time.Hour),
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, RelevantToFailure(tt.activeFrom, tt.activeUntil, failedAt))
		})
	}
}

// The window is measured from the failure rather than from when the check runs, so a
// consumer that gets to a failure late reaches the same verdict as one that got there
// immediately.
func TestRelevantToFailureIsAnchoredOnTheFailure(t *testing.T) {
	failedAt := time.Now().Add(-6 * time.Hour)
	faultedAt := failedAt.Add(-5 * time.Minute)

	assert.True(t, RelevantToFailure(faultedAt, faultedAt, failedAt))
	assert.False(t, RelevantToFailure(faultedAt, faultedAt, time.Now()))
}
