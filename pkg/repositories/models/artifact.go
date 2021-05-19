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
	ArtifactData       []ArtifactData `gorm:"references:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID;foreignkey:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID"`
	Partitions         []Partition    `gorm:"references:ArtifactID;foreignkey:ArtifactID"`
	Tags               []Tag          `gorm:"references:ArtifactID,DatasetUUID;foreignkey:ArtifactID,DatasetUUID"`
	SerializedMetadata []byte
}

type ArtifactData struct {
	BaseModel
	ArtifactKey
	Name     string `gorm:"primary_key"`
	Location string
}
