package models

// Represents project-domain customizable configuration.
type ProjectAttributes struct {
	BaseModel
	Project  string `gorm:"primary_key"`
	Resource string `gorm:"primary_key"`
	// Serialized flyteidl.admin.MatchingAttributes.
	Attributes []byte
}
