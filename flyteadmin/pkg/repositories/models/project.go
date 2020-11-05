package models

type Project struct {
	BaseModel
	Identifier  string `gorm:"primary_key"`
	Name        string // Human-readable name, not a unique identifier.
	Description string `gorm:"type:varchar(300)"`
	Labels      []byte
	// GORM doesn't save the zero value for ints, so we use a pointer for the State field
	State *int32 `gorm:"default:0;index"`
}
