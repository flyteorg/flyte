package models

import (
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
)

// NamedEntityMetadata primary key
type NamedEntityMetadataKey struct {
	ResourceType core.ResourceType `gorm:"primary_key;named_entity_metadata_type_project_domain_name_idx"`
	Project      string            `gorm:"primary_key;named_entity_metadata_type_project_domain_name_idx"`
	Domain       string            `gorm:"primary_key;named_entity_metadata_type_project_domain_name_idx"`
	Name         string            `gorm:"primary_key;named_entity_metadata_type_project_domain_name_idx"`
}

// Fields to be composed into any named entity
type NamedEntityMetadataFields struct {
	Description string `gorm:"type:varchar(300)"`
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
	Project      string
	Domain       string
	Name         string
}

// Composes an identifier (NamedEntity) and its associated metadata fields
type NamedEntity struct {
	NamedEntityKey
	NamedEntityMetadataFields
}
