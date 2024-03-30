package models

import (
	"time"

	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type OverrideAttributes struct {
	BaseModel
	Version          time.Time
	DocumentLocation storage.DataReference
	Active           bool `gorm:"index:idx_override_attributes_active"`
}
