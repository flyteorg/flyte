package models

type ArtifactKey struct {
	DatasetProject string `gorm:"primary_key"`
	DatasetName    string `gorm:"primary_key"`
	DatasetDomain  string `gorm:"primary_key"`
	DatasetVersion string `gorm:"primary_key"`
	ArtifactID     string `gorm:"primary_key"`
}

type Artifact struct {
	BaseModel
	ArtifactKey
	DatasetUUID        string         `gorm:"type:uuid;index:artifacts_dataset_uuid_idx"`
	Dataset            Dataset        `gorm:"association_autocreate:false"`
	ArtifactData       []ArtifactData `gorm:"association_foreignkey:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID;foreignkey:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID"`
	Partitions         []Partition    `gorm:"association_foreignkey:ArtifactID;foreignkey:ArtifactID"`
	SerializedMetadata []byte
}

type ArtifactData struct {
	BaseModel
	ArtifactKey
	Name     string `gorm:"primary_key"`
	Location string
}
