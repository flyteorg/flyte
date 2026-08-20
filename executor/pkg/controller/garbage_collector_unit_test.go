package controller

import (
	"testing"
	"time"
)

func TestShouldDeleteTerminalTaskAction(t *testing.T) {
	now := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		completedTime string
		maxTTL        time.Duration
		want          bool
	}{
		{
			name:          "expired positive ttl",
			completedTime: now.Add(-2 * time.Hour).Format(labelTimeFormat),
			maxTTL:        time.Hour,
			want:          true,
		},
		{
			name:          "recent positive ttl",
			completedTime: now.Format(labelTimeFormat),
			maxTTL:        time.Hour,
			want:          false,
		},
		{
			name:          "zero ttl deletes immediately",
			completedTime: now.Format(labelTimeFormat),
			maxTTL:        0,
			want:          true,
		},
		{
			name:          "negative ttl deletes immediately",
			completedTime: now.Format(labelTimeFormat),
			maxTTL:        -time.Second,
			want:          true,
		},
		{
			name:          "missing completed time is retained",
			completedTime: "",
			maxTTL:        0,
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldDeleteTerminalTaskAction(tt.completedTime, tt.maxTTL, now)
			if got != tt.want {
				t.Fatalf("shouldDeleteTerminalTaskAction() = %v, want %v", got, tt.want)
			}
		})
	}
}
