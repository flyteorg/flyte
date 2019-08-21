package models

type DatasetKey struct {
	Project string `gorm:"primary_key"`
	Name    string `gorm:"primary_key"`
	Domain  string `gorm:"primary_key"`
	Version string `gorm:"primary_key"`
}

type Dataset struct {
	BaseModel
	DatasetKey
	SerializedMetadata []byte
}
