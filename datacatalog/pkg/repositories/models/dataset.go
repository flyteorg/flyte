package models

import (
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type DatasetKey struct {
	Project string `gorm:"primary_key;"`                          // part of pkey, no index needed as it is first column in the pkey
	Name    string `gorm:"primary_key;index:dataset_name_idx"`    // part of pkey and has separate index for filtering
	Domain  string `gorm:"primary_key;index:dataset_domain_idx"`  // part of pkey and has separate index for filtering
	Version string `gorm:"primary_key;index:dataset_version_idx"` // part of pkey and has separate index for filtering
	UUID    string `gorm:"type:uuid;unique;"`
}

type Dataset struct {
	BaseModel
	DatasetKey
	SerializedMetadata []byte
	PartitionKeys      []PartitionKey `gorm:"references:UUID;foreignkey:DatasetUUID"`
}

type PartitionKey struct {
	BaseModel
	DatasetUUID string `gorm:"type:uuid;primary_key"`
	Name        string `gorm:"primary_key"`
}

// BeforeCreate so that we set the UUID in golang rather than from a DB function call
func (dataset *Dataset) BeforeCreate(tx *gorm.DB) error {
	if dataset.UUID == "" {
		generated := uuid.NewV4()
		tx.Model(dataset).Update("UUID", generated)
	}
	return nil
}
