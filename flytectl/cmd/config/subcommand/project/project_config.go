package project

import (
	"fmt"
	"io/ioutil"

	"github.com/flyteorg/flytectl/clierrors"
	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/pkg/filters"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"gopkg.in/yaml.v3"
)

//go:generate pflags Config --default-var DefaultConfig --bind-default-var
var (
	DefaultConfig = &Config{
		Filter: filters.DefaultFilter,
	}
)

// Config holds the flag for get project
type Config struct {
	Filter filters.Filters `json:"filter" pflag:","`
}

//go:generate pflags ConfigProject --default-var DefaultProjectConfig --bind-default-var

// ConfigProject hold configuration for project update flags.
type ConfigProject struct {
	ID              string            `json:"id" pflag:",id for the project specified as argument."`
	ActivateProject bool              `json:"activateProject" pflag:",(Deprecated) Activates the project specified as argument. Only used in update"`
	ArchiveProject  bool              `json:"archiveProject" pflag:",(Deprecated) Archives the project specified as argument. Only used in update"`
	Activate        bool              `json:"activate" pflag:",Activates the project specified as argument. Only used in update"`
	Archive         bool              `json:"archive" pflag:",Archives the project specified as argument. Only used in update"`
	Force           bool              `json:"force" pflag:",Skips asking for an acknowledgement during an update operation. Only used in update"`
	Name            string            `json:"name" pflag:",name for the project specified as argument."`
	DryRun          bool              `json:"dryRun" pflag:",execute command without making any modifications."`
	Description     string            `json:"description" pflag:",description for the project specified as argument."`
	Labels          map[string]string `json:"labels" pflag:",labels for the project specified as argument."`
	File            string            `json:"file" pflag:",file for the project definition."`
}

var DefaultProjectConfig = &ConfigProject{
	Description: "",
	Labels:      map[string]string{},
}

// GetProjectSpec return project spec from a file/flags
func (c *ConfigProject) GetProjectSpec(cf *config.Config) (*admin.Project, error) {
	projectSpec := admin.Project{}

	if len(c.File) > 0 {
		yamlFile, err := ioutil.ReadFile(c.File)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(yamlFile, &projectSpec)
		if err != nil {
			return nil, err
		}
	} else {
		projectSpec.Id = c.ID
		projectSpec.Name = c.Name
		projectSpec.Description = c.Description
		projectSpec.Labels = &admin.Labels{
			Values: c.Labels,
		}
		projectState, err := c.MapToAdminState()
		if err != nil {
			return nil, err
		}
		projectSpec.State = projectState
	}

	project := cf.Project
	if len(projectSpec.Id) == 0 && len(project) == 0 {
		err := fmt.Errorf(clierrors.ErrProjectNotPassed)
		return nil, err
	}

	if len(projectSpec.Id) > 0 && len(project) > 0 {
		err := fmt.Errorf(clierrors.ErrProjectIDBothPassed)
		return nil, err
	}

	// Get projectId from file, if not provided, fall back to project
	if len(projectSpec.Id) == 0 {
		projectSpec.Id = project
	}
	return &projectSpec, nil
}

// MapToAdminState return project spec from a file/flags
func (c *ConfigProject) MapToAdminState() (admin.Project_ProjectState, error) {
	if c.ActivateProject {
		c.Activate = c.ActivateProject
	}
	if c.ArchiveProject {
		c.Archive = c.ArchiveProject
	}

	activate := c.Activate
	archive := c.Archive

	if activate || archive {
		if activate == archive {
			return admin.Project_ACTIVE, fmt.Errorf(clierrors.ErrInvalidStateUpdate)
		}
		if archive {
			return admin.Project_ARCHIVED, nil
		}
	}
	return admin.Project_ACTIVE, nil
}
