package models

import "time"

// ReservationKey uniquely identifies a reservation
type ReservationKey struct {
	DatasetProject string `gorm:"primary_key"`
	DatasetName    string `gorm:"primary_key"`
	DatasetDomain  string `gorm:"primary_key"`
	DatasetVersion string `gorm:"primary_key"`
	TagName        string `gorm:"primary_key"`
}

// Reservation tracks the metadata needed to allow
// task cache serialization
type Reservation struct {
	BaseModel
	ReservationKey

	// Identifies who owns the reservation
	OwnerID string

	// When the reservation will expire
	ExpiresAt          time.Time
	SerializedMetadata []byte
}
