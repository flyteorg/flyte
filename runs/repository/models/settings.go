package models

import (
	"fmt"
	"time"
)

// Settings is one row per settings scope (instance, domain, or project).
type Settings struct {
	ID        uint      `db:"id"`
	Key       string    `db:"key"`
	Data      []byte    `db:"data"`
	Version   uint64    `db:"version"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// EncodeSettingsKey encodes a settings scope as "v1::{domain}:{project}".
// The org segment is always empty: OSS Flyte has no organization concept, so
// settings are stored and looked up under the same key whatever org a client
// sends. Empty domain/project segments are kept, so an instance-level key
// looks like "v1:::".
func EncodeSettingsKey(domain, project string) string {
	return fmt.Sprintf("v1::%s:%s", domain, project)
}
