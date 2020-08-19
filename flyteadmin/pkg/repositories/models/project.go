package models

type Project struct {
	BaseModel
	Identifier  string `gorm:"primary_key"`
	Name        string // Human-readable name, not a unique identifier.
	Description string `gorm:"type:varchar(300)"`
	Labels      []byte
}
