package gormimpl

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var Migrations = []*gormigrate.Migration{
	{
		ID: "2023-10-12-inits",
		Migrate: func(tx *gorm.DB) error {
			type Marker struct {
			}
			return tx.AutoMigrate(
				&Marker{},
			)
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(
				"markers",
			)
		},
	},
}
