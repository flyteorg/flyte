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
	ArtifactID  string   `gorm:"index:tags_dataset_uuid_artifact_id_idx,priority:2"`
	DatasetUUID string   `gorm:"type:uuid;index:tags_dataset_uuid_idx;index:tags_dataset_uuid_artifact_id_idx,priority:1"`
	Artifact    Artifact `gorm:"references:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID;foreignkey:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID"`
}
