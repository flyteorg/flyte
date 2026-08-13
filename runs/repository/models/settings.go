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

// DefaultOrg mirrors secret.DefaultOrganization; duplicated so runs/ doesn't
// take a flyteplugins import for one constant.
const DefaultOrg = "flyte"

// NormalizeOrg returns DefaultOrg for an empty org, otherwise org unchanged.
func NormalizeOrg(org string) string {
	if org == "" {
		return DefaultOrg
	}
	return org
}

// EncodeSettingsKey encodes a settings scope as "v1:{org}:{domain}:{project}".
// Empty org is normalized to DefaultOrg; empty domain/project segments are
// kept, so an instance-level key looks like "v1:flyte::".
func EncodeSettingsKey(org, domain, project string) string {
	return fmt.Sprintf("v1:%s:%s:%s", NormalizeOrg(org), domain, project)
}
