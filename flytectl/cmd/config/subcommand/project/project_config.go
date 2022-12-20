package project

import (
	"fmt"
	"io/ioutil"

	"github.com/flyteorg/flytectl/clierrors"

	"github.com/flyteorg/flytectl/pkg/filters"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"gopkg.in/yaml.v2"
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
func (c *ConfigProject) GetProjectSpec(id string) (*admin.Project, error) {
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
		if len(id) > 0 {
			projectSpec.Id = id
		}
		return &projectSpec, nil
	}

	projectSpec.Id = id
	projectSpec.Name = c.Name
	projectSpec.Description = c.Description
	projectSpec.Labels = &admin.Labels{
		Values: c.Labels,
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
