package v1

import "context"

type PluginLoaderFn func(ctx context.Context) error

var loaders []PluginLoaderFn

// Registers a plugin loader to be called when it's safe to perform plugin initialization logic. This function is NOT
// thread-safe and is expected to be called in an init() function (which runs in a single thread).
func RegisterLoader(fn PluginLoaderFn) {
	loaders = append(loaders, fn)
}

// Runs all plugin loader functions and errors out if any of the loaders fails to finish successfully.
func RunAllLoaders(ctx context.Context) error {
	for _, fn := range loaders {
		if err := fn(ctx); err != nil {
			return err
		}
	}

	return nil
}
