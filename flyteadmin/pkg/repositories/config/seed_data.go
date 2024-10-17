package config

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

type SeedProject struct {
	Name        string `json:"name" pflag:",Name of flyte project to create"`
	Description string `json:"description" pflag:",Description of flyte project to create"`
}

func NewSeedProjectsFromNames(names []string) []*SeedProject {
	projects := make([]*SeedProject, len(names))
	for i, name := range names {
		projects[i] = &SeedProject{
			Name:        name,
			Description: fmt.Sprintf("%s description", name),
		}
	}
	return projects
}

func (p *SeedProject) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var stringVal string
	if err := unmarshal(&stringVal); err == nil {
		p.Name = stringVal
		p.Description = ""
		return nil
	}

	// Define an alias to avoid infinite recursion
	type seedProjectAlias SeedProject
	var project seedProjectAlias
	if err := unmarshal(&project); err == nil {
		*p = SeedProject(project)
		return nil
	}

	// If neither unmarshalling succeeded, return an error
	return fmt.Errorf("seedProjects must be a list of strings or a list of project definitions")
}

// Returns a function to seed the database with default values.
func SeedProjects(db *gorm.DB, projects []*SeedProject) error {
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
