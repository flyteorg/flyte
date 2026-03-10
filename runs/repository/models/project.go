package models

import (
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
)

type Project struct {
	Identifier  string `gorm:"primary_key"`
	Name        string `valid:"length(0|255)"` // Human-readable name, not a unique identifier.
	Description string `gorm:"type:varchar(300)"`
	Labels      []byte
	// GORM doesn't save the zero value for ints, so we use a pointer for the State field.
	State *int32 `gorm:"default:0;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
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
