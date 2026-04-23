package models

import "time"

// AppConditions stores the serialized condition history for a single app.
type AppConditions struct {
	Project    string    `db:"project"`
	Domain     string    `db:"domain"`
	Name       string    `db:"name"`
	Conditions []byte    `db:"conditions"` // proto-serialized flyteidl2.app.ConditionList
	UpdatedAt  time.Time `db:"updated_at"`
}
