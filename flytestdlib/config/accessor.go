// A strongly-typed config library to parse configs from PFlags, Env Vars and Config files.
// Config package enables consumers to access (readonly for now) strongly typed configs without worrying about mismatching
// keys or casting to the wrong type. It supports basic types (e.g. int, string) as well as more complex structures through
// json encoding/decoding.
//
// Config package introduces the concept of Sections. Each section should be given a unique section key. The binary will
// not load if there is a conflict. Each section should be represented as a Go struct and registered at startup before
// config is loaded/parsed.
//
// Sections can be nested too. A new config section can be registered as a sub-section of an existing one. This allows
// dynamic grouping of sections while continuing to enforce strong-typed parsing of configs.
//
// Config data can be parsed from supported config file(s) (yaml, prop, toml), env vars, PFlags or a combination of these
// Precedence is (flags,  env vars, config file, defaults). When data is read from config files, a file watcher is started
// to monitor for changes in those files. If the registrant of a section subscribes to changes then a handler is called
// when the relevant section has been updated. Sections within a single config file will be invoked after all sections
// from that particular config file are parsed. It follows that if there are inter-dependent sections (e.g. changing one
// MUST be followed by a change in another), then make sure those sections are placed in the same config file.
//
// A convenience tool is also provided in cli package (pflags) that generates an implementation for PFlagProvider interface
// based on json names of the fields.
package config

import (
	"context"
	"flag"

	"github.com/spf13/pflag"
)

// Provides a simple config parser interface.
type Accessor interface {
	// Gets a friendly identifier for the accessor.
	ID() string

	// Initializes the config parser with golang's default flagset.
	InitializeFlags(cmdFlags *flag.FlagSet)

	// Initializes the config parser with pflag's flagset.
	InitializePflags(cmdFlags *pflag.FlagSet)

	// Parses and validates config file(s) discovered then updates the underlying config section with the results.
	// Exercise caution when calling this because multiple invocations will overwrite each other's results.
	UpdateConfig(ctx context.Context) error

	// Gets path(s) to the config file(s) used.
	ConfigFilesUsed() []string
}

// Options used to initialize a Config Accessor
type Options struct {
	// Instructs parser to fail if any section/key in the config file read do not have a corresponding registered section.
	StrictMode bool

	// Search paths to look for config file(s). If not specified, it searches for config.yaml under current directory as well
	// as /etc/flyte/config directories.
	SearchPaths []string

	// Defines the root section to use with the accessor.
	RootSection Section
}
