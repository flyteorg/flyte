package storage

import (
	"fmt"

	"github.com/lyft/flytestdlib/promutils"

	"github.com/graymeta/stow"
	"github.com/graymeta/stow/local"
)

func getLocalStowConfigMap(cfg *Config) stow.ConfigMap {
	stowConfig := stow.ConfigMap{}
	if endpoint := cfg.Connection.Endpoint.String(); endpoint != "" {
		stowConfig[local.ConfigKeyPath] = endpoint
	}

	return stowConfig
}

// Creates a Data store backed by Stow-S3 raw store.
func newLocalRawStore(cfg *Config, metricsScope promutils.Scope) (RawStore, error) {
	if cfg.InitContainer == "" {
		return nil, fmt.Errorf("initContainer is required")
	}

	loc, err := stow.Dial(local.Kind, getLocalStowConfigMap(cfg))

	if err != nil {
		return emptyStore, fmt.Errorf("unable to configure the storage for local. Error: %v", err)
	}

	c, err := loc.Container(cfg.InitContainer)
	if err != nil {
		if IsNotFound(err) {
			c, err = loc.CreateContainer(cfg.InitContainer)
			if err != nil && !IsExists(err) {
				return emptyStore, fmt.Errorf("unable to initialize container [%v]. Error: %v", cfg.InitContainer, err)
			}

			return NewStowRawStore(DataReference(c.Name()), c, metricsScope)
		}

		return emptyStore, err
	}

	return NewStowRawStore(DataReference(c.Name()), c, metricsScope)
}
