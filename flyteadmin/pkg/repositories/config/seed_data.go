package config

import (
	"context"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

type Project struct {
	Name        string
	Description string
}

// Returns a function to seed the database with default values.
func SeedProjects(db *gorm.DB, projects []Project) error {
	tx := db.Begin()
	for _, project := range projects {
		projectModel := models.Project{
			Identifier:  project.Name,
			Name:        project.Name,
			Description: project.Description,
		}
		if err := tx.Where(models.Project{Identifier: project.Name}).Omit("id").FirstOrCreate(&projectModel).Error; err != nil {
			logger.Warningf(context.Background(), "failed to save project [%s]", project)
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}
