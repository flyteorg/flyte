package interfaces

//go:generate mockery -name ExternalResourceConfiguration -output=mocks -case=underscore

// Connection store the secret and config information required to connect to an external service.
type Connection struct {
	TaskType string            `json:"taskType"`
	Secrets  map[string]string `json:"secrets"`
	Configs  map[string]string `json:"configs"`
}

type ExternalResourceConfig struct {
	Connections map[string]Connection `json:"connections"`
}

type ExternalResourceConfiguration interface {
	GetConnections() map[string]Connection
}
