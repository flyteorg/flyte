package io

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/storage"
)

//go:generate mockery -all -case=underscore

// InputFilePaths contains the different ways available for downstream systems to retrieve inputs.
// If using Files for IO with tasks, then the input will be written to this path. All the files are always created in a
// sandbox per execution
type InputFilePaths interface {
	// GetInputPrefixPath returns the inputs file path, minus the protobuf file name.
	GetInputPrefixPath() storage.DataReference
	// GetInputPath returns a path for where the protobuf encoded inputs of type `core.LiteralMap` can be found. The returned value is an URN in the configured storage backend
	GetInputPath() storage.DataReference
}

// InputReader provides a method to access the inputs for a task execution within the plugin's Task Context
type InputReader interface {
	InputFilePaths
	// Get the inputs for this task as a literal map, an error is returned only in case of systemic errors.
	// No outputs or void is indicated using *core.LiteralMap -> nil
	Get(ctx context.Context) (*core.LiteralMap, error)
}

// OutputReader provides an abstracted OutputReader interface. The plugins are responsible to provide
// the implementations for the interface. Some helper implementations can be found in ioutils
type OutputReader interface {
	// IsError returns true if an error was detected when reading the output and false if no error was detected
	IsError(ctx context.Context) (bool, error)
	// ReadError returns the error as type ExecutionError
	ReadError(ctx context.Context) (ExecutionError, error)
	// IsFile returns true if the outputs are using the OutputFilePaths specified files. If so it allows the system to
	// optimize the reads of the files
	IsFile(ctx context.Context) bool
	// Exists returns true if the output exists false otherwise
	Exists(ctx context.Context) (bool, error)
	// Read returns the output -> *core.LiteralMap (nil if void), *ExecutionError if user error when reading the output and error to indicate system problems
	Read(ctx context.Context) (*core.LiteralMap, *ExecutionError, error)
	// DeckExists checks if the deck file has been generated.
	DeckExists(ctx context.Context) (bool, error)
}

// CheckpointPaths provides the paths / keys to input Checkpoints directory and an output checkpoints directory.
type CheckpointPaths interface {
	// GetPreviousCheckpointsPrefix returns the storage prefix for checkpoints for the previous iteration / attempt.
	// It is optional and can be an empty string in some cases
	GetPreviousCheckpointsPrefix() storage.DataReference
	// GetCheckpointPrefix returns the storage prefix that should be used to store checkpoints for the current attempt
	// The path is not accessible to Flyte backend and are stored in the users raw path
	GetCheckpointPrefix() storage.DataReference
}

// RawOutputPaths is the actual path where the data produced by a task can be placed. It is completely optional. The advantage
// of using this path is to provide exactly once semantics. It is guaranteed that this path is unique for every new execution
// of a task (across retries etc) and is constant for a specific execution.
// As of 02/20/2020 Flytekit generates this path randomly for S3. This structure proposes migration of this logic to
// FlytePluginMachinery so that it can be used more universally outside of Flytekit.
type RawOutputPaths interface {
	// GetRawOutputPrefix is the prefix (blob store prefix or directory) where all data produced can be stored.
	GetRawOutputPrefix() storage.DataReference
}

// OutputFilePaths contains and provides all paths where various meta outputs produced by the task can be placed,
// such that the framework can directly access them. Every path is represented using storage.DataReference ->
// an URN for the configured storage backend
type OutputFilePaths interface {
	// RawOutputPaths are available with OutputFilePaths
	RawOutputPaths

	// CheckpointPaths that can be optionally used to checkpoint
	CheckpointPaths

	// GetOutputPrefixPath returns a path to a directory or prefix that contains all execution metadata for this execution
	GetOutputPrefixPath() storage.DataReference
	// GetOutputPath returns a fully qualified path (URN) to where the framework expects the output to exist in the configured storage backend
	GetOutputPath() storage.DataReference
	// GetDeckPath returns a fully qualified path (URN) to where the framework expects the deck.html to exist in the configured storage backend
	GetDeckPath() storage.DataReference
	// GetErrorPath returns a fully qualified path (URN) where the error information should be placed as a protobuf core.ErrorDocument. It is not directly
	// used by the framework, but could be used in the future
	GetErrorPath() storage.DataReference
}

// OutputWriter provides an interface to write back the outputs to the engine.
type OutputWriter interface {
	OutputFilePaths
	// Put Once the task completes, use this method to indicate the output accessor to the framework
	Put(ctx context.Context, reader OutputReader) error
}

// ExecutionError Indicates any error in executing the task
type ExecutionError struct {
	// Core error structure
	*core.ExecutionError
	// Indicates if this error is recoverable
	IsRecoverable bool
}
