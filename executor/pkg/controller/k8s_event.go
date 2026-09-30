package controller

import (
	"strings"
)

// truncateUTF8 cuts s to at most limit bytes, with any invalid UTF-8
// removed. limit must be non-negative.
func truncateUTF8(s string, limit int) string {
	if len(s) > limit {
		s = s[:limit]
	}
	return strings.ToValidUTF8(s, "")
}
