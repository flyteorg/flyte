package models

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// NamedEntityMetadata primary key
type NamedEntityMetadataKey struct {
	ResourceType core.ResourceType `gorm:"primary_key;index:named_entity_metadata_type_project_domain_name_idx" valid:"length(0|255)"`
	Project      string            `gorm:"primary_key;index:named_entity_metadata_type_project_domain_name_idx" valid:"length(0|255)"`
	Domain       string            `gorm:"primary_key;index:named_entity_metadata_type_project_domain_name_idx" valid:"length(0|255)"`
	Name         string            `gorm:"primary_key;index:named_entity_metadata_type_project_domain_name_idx" valid:"length(0|255)"`
}

// Fields to be composed into any named entity
type NamedEntityMetadataFields struct {
	Description string `gorm:"type:varchar(300)"`
	// GORM doesn't save the zero value for ints, so we use a pointer for the State field
	State *int32 `gorm:"default:0"`
}

// Database model to encapsulate metadata associated with a NamedEntity
type NamedEntityMetadata struct {
	BaseModel
	NamedEntityMetadataKey
	NamedEntityMetadataFields
}

// NamedEntity key. This is used as a lookup for NamedEntityMetadata, so the
// fields here should match the ones in NamedEntityMetadataKey.
type NamedEntityKey struct {
	ResourceType core.ResourceType
	Project      string `valid:"length(0|255)"`
	Domain       string `valid:"length(0|255)"`
	Name         string `valid:"length(0|255)"`
}

// Composes an identifier (NamedEntity) and its associated metadata fields
type NamedEntity struct {
	NamedEntityKey
	NamedEntityMetadataFields
}

var (
	NamedEntityColumns         = modelColumns(NamedEntity{})
	NamedEntityMetadataColumns = modelColumns(NamedEntityMetadata{})
)
