package models

// Main Use cases:
// 1. Filter artifacts by partition key/val in a dataset from UI [x]
// 2. Get the artifact that has the partitions (x,y,z + tag_name) = latest [?]
type Partition struct {
	BaseModel
	DatasetUUID    string `gorm:"primary_key"`
	PartitionKey   string `gorm:"primary_key"`
	PartitionValue string `gorm:"primary_key"`
	ArtifactID     string
}
