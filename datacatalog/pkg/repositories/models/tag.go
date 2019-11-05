package models

type TagKey struct {
	DatasetProject string `gorm:"primary_key"`
	DatasetName    string `gorm:"primary_key"`
	DatasetDomain  string `gorm:"primary_key"`
	DatasetVersion string `gorm:"primary_key"`
	TagName        string `gorm:"primary_key"`
}

type Tag struct {
	BaseModel
	TagKey
	ArtifactID  string
	DatasetUUID string   `gorm:"type:uuid;index:tags_dataset_uuid_idx"`
	Artifact    Artifact `gorm:"association_foreignkey:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID;foreignkey:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID"`
}
