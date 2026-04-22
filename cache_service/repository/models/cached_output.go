package models

import "time"

type CachedOutput struct {
	Key         string    `db:"key"`
	OutputURI   string    `db:"output_uri"`
	Metadata    []byte    `db:"metadata"`
	LastUpdated time.Time `db:"last_updated"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
