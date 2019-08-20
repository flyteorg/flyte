// This package contains various task plugins that offer an extensibility point into how propeller executes tasks.
// Check the documentation of either version for examples on how to develop and register plugins.
package tasks

import (
	"context"

	"github.com/pkg/errors"

	v1 "github.com/lyft/flyteplugins/go/tasks/v1"

	// This is a temporary solution to invoke init() methods on all plugins. Ideally this step should happen dynamically
	// based on a config.
	_ "github.com/lyft/flyteplugins/go/tasks/v1/k8splugins"
	_ "github.com/lyft/flyteplugins/go/tasks/v1/qubole"
)

func Load(ctx context.Context) error {
	if err := v1.RunAllLoaders(ctx); err != nil {
		return err
	}

	if len(v1.ListAllTaskExecutors()) == 0 {
		return errors.Errorf("No Task Executor defined.")
	}

	return nil
}
