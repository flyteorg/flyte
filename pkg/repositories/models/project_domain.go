package models

// Represents project-domain customizable configuration.
type ProjectDomain struct {
	BaseModel
	Project string `gorm:"primary_key"`
	Domain  string `gorm:"primary_key"`
	// Key-value pairs of substitutable resource attributes.
	Attributes []byte
}
