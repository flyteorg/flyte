package client

import (
	"context"
	"net/http"
	"net/url"

	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/presto/config"
)

const (
	httpRequestTimeoutSecs = 30
)

type noopPrestoClient struct {
	client      *http.Client
	environment *url.URL
}

func (p noopPrestoClient) ExecuteCommand(
	ctx context.Context,
	queryStr string,
	executeArgs PrestoExecuteArgs) (PrestoExecuteResponse, error) {

	return PrestoExecuteResponse{}, nil
}

func (p noopPrestoClient) KillCommand(ctx context.Context, commandID string) error {
	return nil
}

func (p noopPrestoClient) GetCommandStatus(ctx context.Context, commandID string) (PrestoStatus, error) {
	return PrestoStatusUnknown, nil
}

func NewNoopPrestoClient(cfg *config.Config) PrestoClient {
	return &noopPrestoClient{
		client:      &http.Client{Timeout: httpRequestTimeoutSecs * time.Second},
		environment: cfg.Environment.ResolveReference(&cfg.Environment.URL),
	}
}
