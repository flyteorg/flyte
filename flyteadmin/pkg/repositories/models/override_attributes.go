package models

import (
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type OverrideAttributes struct {
	BaseModel
	Version          string
	DocumentLocation storage.DataReference
	Active           bool `gorm:"index:idx_override_attributes_active"`
}
