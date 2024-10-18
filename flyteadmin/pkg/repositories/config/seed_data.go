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

func UniqueProjectsFromNames(names []string) []SeedProject {
	return uniqueProjects(names, nil)
}

// MergeSeedProjectsWithUniqueNames merges seed projects from names and details while maintaining uniqueness
func MergeSeedProjectsWithUniqueNames(seedProjects []string, seedProjectsWithDetails []SeedProject) []SeedProject {
	return uniqueProjects(seedProjects, seedProjectsWithDetails)
}

func uniqueProjects(seedProjects []string, seedProjectsWithDetails []SeedProject) []SeedProject {
	// Track unique project names
	seen := make(map[string]struct{})

	// Create the final result slice
	var combinedProjects []SeedProject

	// First, add all projects from SeedProjectsWithDetails to the map
	for _, project := range seedProjectsWithDetails {
		// Handle the duplication
		if _, exists := seen[project.Name]; !exists {
			seen[project.Name] = struct{}{}
			combinedProjects = append(combinedProjects, project)
		}
	}

	// Process SeedProjects
	for _, projectName := range seedProjects {
		// Check if project not exists in SeedProjectsWithDetails
		if _, exists := seen[projectName]; !exists {
			seen[projectName] = struct{}{}
			combinedProjects = append(combinedProjects, SeedProject{
				Name:        projectName,
				Description: fmt.Sprintf("%s description", projectName),
			})
		}
	}

	return combinedProjects
}

// Returns a function to seed the database with default values.
func SeedProjects(db *gorm.DB, projects []SeedProject) error {
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
