package models

type PartitionKey struct {
	DatasetProject string `gorm:"primary_key"`
	DatasetName    string `gorm:"primary_key"`
	DatasetDomain  string `gorm:"primary_key"`
	DatasetVersion string `gorm:"primary_key"`
	PartitionKey   string `gorm:"primary_key"`
	PartitionValue string `gorm:"primary_key"`
}

type Partition struct {
	BaseModel
	PartitionKey
	ArtifactID string
	Artifact   Artifact `gorm:"association_foreignkey:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID;foreignkey:DatasetProject,DatasetName,DatasetDomain,DatasetVersion,ArtifactID"`
}
