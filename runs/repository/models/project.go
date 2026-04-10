package models

import (
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
)

type Project struct {
	Identifier  string    `db:"identifier"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Labels      []byte    `db:"labels"`
	State       *int32    `db:"state"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
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
