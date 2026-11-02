package config

import (
	"time"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/serviceclient"
	"k8s.io/apimachinery/pkg/api/resource"
)

const configSectionKey = "dataproxy"

//go:generate pflags DataProxyConfig --default-var=defaultConfig

var defaultConfig = &DataProxyConfig{
	Server: ServerConfig{
		Port: 8088,
		Host: "0.0.0.0",
	},
	Upload: DataProxyUploadConfig{
		MaxSize:               resource.MustParse("100Mi"),
		MaxExpiresIn:          config.Duration{Duration: 3600000000000}, // 1 hour
		DefaultFileNameLength: 20,
		StoragePrefix:         "uploads",
	},
	Download: DataProxyDownloadConfig{
		MaxExpiresIn: config.Duration{Duration: 3600000000000}, // 1 hour
	},
	RunService: serviceclient.ServiceConfig{URL: "http://localhost:8090"},
	Logs: LogsConfig{
		Persisted: PersistedLogsConfig{
			Type: PersistedLogsTypeNone,
			Loki: DefaultLokiConfig(),
		},
	},
}

var configSection = config.MustRegisterSection(configSectionKey, defaultConfig)

type DataProxyConfig struct {
	// Server configures the standalone DataProxy service HTTP listener.
	Server ServerConfig `json:"server"`

	// PprofPort is the port the pprof endpoint binds to. 0 disables profiling.
	PprofPort int `json:"pprofPort" pflag:",Port the pprof endpoint binds to; 0 disables profiling"`

	Upload   DataProxyUploadConfig   `json:"upload" pflag:",Defines data proxy upload configuration."`
	Download DataProxyDownloadConfig `json:"download" pflag:",Defines data proxy download configuration."`

	// RunService configures the Runs service client. The Runs component also
	// hosts the Task, Trigger, and Project APIs used by dataproxy.
	RunService serviceclient.ServiceConfig `json:"runService" pflag:",Runs service client configuration."`

	Logs LogsConfig `json:"logs"`
}

type LogsConfig struct {
	// Persisted is consulted once the Kubernetes pod no longer exists.
	Persisted PersistedLogsConfig `json:"persisted"`
}

type PersistedLogsType string

const (
	PersistedLogsTypeNone PersistedLogsType = ""
	PersistedLogsTypeLoki PersistedLogsType = "loki"
)

type PersistedLogsConfig struct {
	Type PersistedLogsType `json:"type" pflag:",Persisted log backend type: '' (disabled) or 'loki'."`
	Loki LokiConfig        `json:"loki"`
}

type LokiConfig struct {
	URL           string              `json:"url" pflag:",Loki base URL."`
	QueryTemplate string              `json:"queryTemplate" pflag:",Go template for the LogQL query; fields .Namespace .PodName .ContainerName."`
	TenantID      string              `json:"tenantId" pflag:",Loki tenant (X-Scope-OrgID header); empty for single-tenant."`
	BasicAuth     LokiBasicAuthConfig `json:"basicAuth"`
	BearerToken   string              `json:"bearerToken" pflag:",Bearer token sent in the Authorization header."`
	Timeout       config.Duration     `json:"timeout" pflag:",Per-request timeout."`
	PageSize      int                 `json:"pageSize" pflag:",Lines per Loki query_range request; must not exceed Loki's max_entries_limit_per_query."`
	MaxLines      int                 `json:"maxLines" pflag:",Maximum lines returned per attempt; 0 for unlimited."`
	TimePadding   config.Duration     `json:"timePadding" pflag:",Padding added around the container start/end window."`
	Lookback      config.Duration     `json:"lookback" pflag:",Window queried when no container start time is recorded."`
}

type LokiBasicAuthConfig struct {
	Username string `json:"username" pflag:",Basic auth username."`
	Password string `json:"password" pflag:",Basic auth password."`
}

// DefaultLokiConfig matches the labels Promtail and Alloy attach to pod logs.
func DefaultLokiConfig() LokiConfig {
	return LokiConfig{
		QueryTemplate: `{namespace="{{ .Namespace }}", pod="{{ .PodName }}", container="{{ .ContainerName }}"}`,
		Timeout:       config.Duration{Duration: 30 * time.Second},
		PageSize:      1000,
		TimePadding:   config.Duration{Duration: time.Minute},
		Lookback:      config.Duration{Duration: 7 * 24 * time.Hour},
	}
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port int    `json:"port" pflag:",Port to bind the HTTP server"`
	Host string `json:"host" pflag:",Host to bind the HTTP server"`
}

// GetConfig returns the parsed data proxy configuration
func GetConfig() *DataProxyConfig {
	return configSection.GetConfig().(*DataProxyConfig)
}

type DataProxyDownloadConfig struct {
	MaxExpiresIn config.Duration `json:"maxExpiresIn" pflag:",Maximum allowed expiration duration."`
}

type DataProxyUploadConfig struct {
	MaxSize               resource.Quantity `json:"maxSize" pflag:",Maximum allowed upload size."`
	MaxExpiresIn          config.Duration   `json:"maxExpiresIn" pflag:",Maximum allowed expiration duration."`
	DefaultFileNameLength int               `json:"defaultFileNameLength" pflag:",Default length for the generated file name if file name not provided in the request."`
	StoragePrefix         string            `json:"storagePrefix" pflag:",Storage prefix to use for all upload requests."`
}
