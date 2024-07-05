package entrypoints

import (
	"context"

	"github.com/spf13/cobra"
	_ "gorm.io/driver/postgres" // Required to import database driver.

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/config"
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime"
	"github.com/flyteorg/flyte/flyteadmin/pkg/server"
)

var parentMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "This command controls migration behavior for the Flyte admin database. Please choose a subcommand.",
}

// This runs all the migrations
var migrateCmd = &cobra.Command{
	Use:   "run",
	Short: "This command will run all the migrations for the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return server.Migrate(ctx)
	},
}

// Rollback the latest migration
var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "This command will rollback one migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		return server.Rollback(ctx)
	},
}

// This seeds the database with project values
var seedProjectsCmd = &cobra.Command{
	Use:   "seed-projects",
	Short: "Seed projects in the database.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		configuration := runtime.NewConfigurationProvider()
		projectConfig := configuration.ApplicationConfiguration().GetProjectsConfig()

		var projectsMap = make(map[string]config.Project)

		for _, project := range *projectConfig {
			projectsMap[project.Name] = config.Project{Name: project.Name, Description: project.Description}
		}

		var projects = make([]config.Project, 0, len(projectsMap))

		for _, project := range projectsMap {
			projects = append(projects, project)
		}

		return server.SeedProjects(ctx, projects)
	},
}

func init() {
	RootCmd.AddCommand(parentMigrateCmd)
	parentMigrateCmd.AddCommand(migrateCmd)
	parentMigrateCmd.AddCommand(rollbackCmd)
	parentMigrateCmd.AddCommand(seedProjectsCmd)
}
