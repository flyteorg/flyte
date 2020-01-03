package models

// Represents project-domain customizable configuration.
type ProjectDomainAttributes struct {
	BaseModel
	Project  string `gorm:"primary_key"`
	Domain   string `gorm:"primary_key"`
	Resource string `gorm:"primary_key"`
	// Serialized flyteidl.admin.MatchingAttributes.
	Attributes []byte
}
