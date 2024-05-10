package interfaces

//go:generate mockery -name ClusterPoolAssignmentConfiguration -output=mocks -case=underscore

type Connection struct {
	Secrets map[string]string `json:"secrets"`
	Configs map[string]string `json:"configs"`
}

type ExternalResource struct {
	Connections map[string]Connection `json:"connections"`
}

type ExternalResourceConfig struct {
	ExternalResource ExternalResource `json:"externalResource"`
}

type ExternalResourceConfiguration interface {
	GetExternalResource() ExternalResource
}
