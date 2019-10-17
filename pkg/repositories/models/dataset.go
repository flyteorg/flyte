package models

type DatasetKey struct {
	Project string `gorm:"primary_key"`
	Name    string `gorm:"primary_key"`
	Domain  string `gorm:"primary_key"`
	Version string `gorm:"primary_key"`
	UUID    string `gorm:"type:uuid;unique;default:uuid_generate_v4()"`
}

type Dataset struct {
	BaseModel
	DatasetKey
	SerializedMetadata []byte
	PartitionKeys      []PartitionKey
}

type PartitionKey struct {
	BaseModel
	DatasetUUID string `gorm:"primary_key"`
	Key         string `gorm:"primary_key"`
}
