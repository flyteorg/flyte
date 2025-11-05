package client

import "context"

// Contains information needed to execute a Presto query
type PrestoExecuteArgs struct {
	RoutingGroup string `json:"routingGroup,omitempty"`
	Catalog      string `json:"catalog,omitempty"`
	Schema       string `json:"schema,omitempty"`
	Source       string `json:"source,omitempty"`
	User         string `json:"user,omitempty"`
}

// Representation of a response after submitting a query to Presto
type PrestoExecuteResponse struct {
	ID      string `json:"id,omitempty"`
	NextURI string `json:"nextUri,omitempty"`
}

//go:generate mockery -all -case=snake

// Interface to interact with PrestoClient for Presto tasks
type PrestoClient interface {
	// Submits a query to Presto
	ExecuteCommand(ctx context.Context, commandStr string, executeArgs PrestoExecuteArgs) (PrestoExecuteResponse, error)

	// Cancels a currently running Presto query
	KillCommand(ctx context.Context, commandID string) error

	// Gets the status of a Presto query
	GetCommandStatus(ctx context.Context, commandID string) (PrestoStatus, error)
}
