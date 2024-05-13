package models

import (
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

// Database model to encapsulate project configuration document metadata.
type ConfigurationDocumentMetadata struct {
	BaseModel
	Version          string `gorm:"index"`
	DocumentLocation storage.DataReference
	Active           bool `gorm:"index,where:active = true"`
}
