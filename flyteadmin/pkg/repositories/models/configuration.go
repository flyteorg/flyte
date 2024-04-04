package models

import (
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type Configuration struct {
	BaseModel
	Version          string
	DocumentLocation storage.DataReference
	Active           bool `gorm:"index:idx_configuration_active"`
}
