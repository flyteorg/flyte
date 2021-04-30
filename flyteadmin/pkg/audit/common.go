package audit

import (
	"time"
)

type AuthenticatedClientMeta struct {
	ClientIds     []string
	TokenIssuedAt time.Time
	ClientIP      string
	Subject       string
}
