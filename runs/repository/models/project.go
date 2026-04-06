package models

import (
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
)

type Project struct {
	Identifier  string `gorm:"primary_key" db:"identifier"`
	Name        string `valid:"length(0|255)" db:"name"` // Human-readable name, not a unique identifier.
	Description string `gorm:"type:varchar(300)" db:"description"`
	Labels      []byte `db:"labels"`
	// GORM doesn't save the zero value for ints, so we use a pointer for the State field.
	State     *int32    `gorm:"default:0;index" db:"state"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

var ProjectColumns = sets.New[string](
	"identifier",
	"name",
	"description",
	"labels",
	"state",
	"created_at",
	"updated_at",
)
