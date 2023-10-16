package db

import (
	"gorm.io/gorm"
)

type ArtifactKey struct {
	gorm.Model
	Project string `gorm:"index:idx_pdn;index:idx_proj;type:varchar(64)"`
	Domain  string `gorm:"index:idx_pdn;index:idx_dom;type:varchar(64)"`
	Name    string `gorm:"index:idx_pdn;index:idx_name;type:varchar(255)"`
}
type Artifact struct {
	gorm.Model
	// gatepr: this doesn't actually create a foreign key...
	ArtifactKeyID uint
	ArtifactKey   ArtifactKey `gorm:"foreignKey:ArtifactKeyID;references:ID"`
	Version       string      `gorm:"not null;type:varchar(255);index:idx_artifact_version"`
	ExecutionName string      `gorm:"type:varchar(255)"`
}

//type Artifact struct {
//	gorm.Model
//	// gatepr: this doesn't actually create a foreign key...
//	ArtifactKeyID uint
//	ArtifactKey   ArtifactKey     `gorm:"foreignKey:ArtifactKeyID;references:ID"`
//	Version       string          `gorm:"not null;type:varchar(255);index:idx_artifact_version"`
//	Partitions    postgres.Hstore `gorm:"type:hstore;index:idx_artifact_partitions"`
//
//	LiteralType  []byte `gorm:"not null"`
//	LiteralValue []byte `gorm:"not null"`
//
//	Description           string `gorm:"type:varchar(255)"`
//	MetadataType          string `gorm:"type:varchar(64)"`
//	OffloadedUserMetadata string `gorm:"type:varchar(255)"`
//
//	// Project/Domain assumed to always be the same as the Artifact
//	ExecutionName   string `gorm:"type:varchar(255)"`
//	WorkflowProject string `gorm:"type:varchar(64)"`
//	WorkflowDomain  string `gorm:"type:varchar(64)"`
//	WorkflowName    string `gorm:"type:varchar(255)"`
//	WorkflowVersion string `gorm:"type:varchar(255)"`
//	TaskProject     string `gorm:"type:varchar(64)"`
//	TaskDomain      string `gorm:"type:varchar(64)"`
//	TaskName        string `gorm:"type:varchar(255)"`
//	TaskVersion     string `gorm:"type:varchar(255)"`
//	NodeID          string `gorm:"type:varchar(64)"`
//	// See Admin migration for note.
//	// Here nullable in the case of workflow output.
//	RetryAttempt *uint32
//}
