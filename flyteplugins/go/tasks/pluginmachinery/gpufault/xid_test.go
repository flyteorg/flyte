package gpufault

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameFor(t *testing.T) {
	tests := []struct {
		name string
		kind Kind
		code int
		want string
	}{
		{name: "documented xid", kind: KindXid, code: 79, want: "GPU has fallen off the bus"},
		{name: "unknown xid falls back", kind: KindXid, code: 4242, want: "Xid 4242"},
		{name: "sxid is always generic", kind: KindSXid, code: 12028, want: "SXid 12028"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NameFor(tt.kind, tt.code))
		})
	}
}

func TestSeverityFor(t *testing.T) {
	tests := []struct {
		name string
		kind Kind
		code int
		want Severity
	}{
		{name: "workload fault", kind: KindXid, code: 31, want: SeverityUser},
		{name: "hardware fault", kind: KindXid, code: 79, want: SeverityCritical},
		{name: "listed warning", kind: KindXid, code: 92, want: SeverityWarn},
		{name: "unlisted code is a warning", kind: KindXid, code: 4242, want: SeverityWarn},
		{name: "every sxid is critical", kind: KindSXid, code: 24001, want: SeverityCritical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SeverityFor(tt.kind, tt.code))
		})
	}
}

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		in   string
		want Severity
	}{
		{in: "user", want: SeverityUser},
		{in: "warn", want: SeverityWarn},
		{in: "critical", want: SeverityCritical},
		{in: "something-new", want: SeverityWarn},
		{in: "", want: SeverityWarn},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.want, ParseSeverity(tt.in))
		})
	}
}

func TestSeverityLabel(t *testing.T) {
	assert.Equal(t, "USER", SeverityUser.Label())
	assert.Equal(t, "WARN", SeverityWarn.Label())
	assert.Equal(t, "CRITICAL", SeverityCritical.Label())
	assert.Equal(t, "WARN", Severity("nonsense").Label())
}
