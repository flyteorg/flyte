package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func TestGetRunName_RootAction(t *testing.T) {
	a := &Action{Name: "run-123"}
	assert.Equal(t, "run-123", a.GetRunName())
}

func TestGetRunName_ChildAction(t *testing.T) {
	parent := "parent-action"
	a := &Action{
		Name:             "child-action",
		ParentActionName: &parent,
		ActionSpec: datatypes.JSON([]byte(`{
			"action_id": {
				"run": {"org": "org", "project": "proj", "domain": "dev", "name": "my-run"},
				"name": "child-action"
			}
		}`)),
	}
	assert.Equal(t, "my-run", a.GetRunName())
}

func TestGetRunName_ChildAction_EmptySpec(t *testing.T) {
	parent := "parent-action"
	a := &Action{
		Name:             "child-action",
		ParentActionName: &parent,
		ActionSpec:       datatypes.JSON([]byte(`{}`)),
	}
	assert.Equal(t, "", a.GetRunName())
}

func TestGetRunName_ChildAction_NilSpec(t *testing.T) {
	parent := "parent-action"
	a := &Action{
		Name:             "child-action",
		ParentActionName: &parent,
	}
	assert.Equal(t, "", a.GetRunName())
}
