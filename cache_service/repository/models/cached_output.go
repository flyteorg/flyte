package models

import "time"

type CachedOutput struct {
	Key         string    `gorm:"primaryKey;size:512" db:"key"`
	OutputURI   string    `gorm:"not null" db:"output_uri"`
	Metadata    []byte    `gorm:"not null" db:"metadata"`
	LastUpdated time.Time `gorm:"not null;index" db:"last_updated"`
	CreatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" db:"created_at"`
	UpdatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" db:"updated_at"`
}

func (CachedOutput) TableName() string { return "cache_service_outputs" }
