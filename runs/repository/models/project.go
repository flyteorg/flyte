package models

import (
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
)

// ProjectColumns are the allowed columns for project queries.
var ProjectColumns = sets.New(
	"id",
	"name",
	"state",
	"created_at",
	"updated_at",
	"org",
)

// ProjectKey is a composite key for a project.
type ProjectKey struct {
	Org string `gorm:"primaryKey;index:idx_projects_identifier,priority:1" db:"org"`
	ID  string `gorm:"primaryKey;index:idx_projects_identifier,priority:2" db:"id"`
}

// Project stores project metadata.
type Project struct {
	ProjectKey `gorm:"embedded"`

	Name        string `gorm:"column:name" db:"name"`
	Description string `gorm:"column:description" db:"description"`
	Labels      []byte `gorm:"column:labels" db:"labels"`
	State       int32  `gorm:"column:state;index:idx_projects_state" db:"state"`

	CreatedAt time.Time `gorm:"column:created_at" db:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" db:"updated_at"`
}

// TableName specifies the table name.
func (Project) TableName() string { return "projects" }
