package server

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/jinzhu/gorm/dialects/postgres"

	"gorm.io/gorm"
)

var Migrations = []*gormigrate.Migration{
	{
		ID: "2023-10-12-inits",
		Migrate: func(tx *gorm.DB) error {
			type ArtifactsKey struct {
				gorm.Model
				Project string `gorm:"primary_key;type:varchar(64)"`
				Domain  string `gorm:"primary_key;type:varchar(64)"`
				Name    string `gorm:"primary_key;type:varchar(255)"`
			}
			type Artifact struct {
				gorm.Model
				ArtifactsKey
				Version    string          `gorm:"type:varchar(255);index:idx_artifact_version"`
				Partitions postgres.Hstore `gorm:"type:hstore;index:idx_artifact_partitions"`

				LiteralType  []byte `gorm:"not null"`
				LiteralValue []byte `gorm:"not null"`

				Description              string `gorm:"type:varchar(255)"`
				OffloadedLongDescription string `gorm:"type:varchar(255)"`
				MetadataType             string `gorm:"type:varchar(64)"`
				OffloadedUserMetadata    string `gorm:"type:varchar(255)"`

				// Project/Domain assumed to always be the same as the Artifact
				ExecutionName   string `gorm:"type:varchar(255)"`
				WorkflowProject string `gorm:"type:varchar(64)"`
				WorkflowDomain  string `gorm:"type:varchar(64)"`
				WorkflowName    string `gorm:"type:varchar(255)"`
				WorkflowVersion string `gorm:"type:varchar(255)"`
				TaskProject     string `gorm:"type:varchar(64)"`
				TaskDomain      string `gorm:"type:varchar(64)"`
				TaskName        string `gorm:"type:varchar(255)"`
				TaskVersion     string `gorm:"type:varchar(255)"`
				NodeID          string `gorm:"type:varchar(64)"`
				// See Admin migration for note.
				// Here nullable in the case of workflow output.
				RetryAttempt *uint32
			}
			return tx.AutoMigrate(
				&ArtifactsKey{}, &Artifact{},
			)
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(
				"markers",
			)
		},
	},
}
