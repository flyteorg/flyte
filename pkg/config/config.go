package config

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	authConfig "github.com/flyteorg/flyteadmin/auth/config"
	"github.com/flyteorg/flytestdlib/config"
)

const SectionKey = "server"

//go:generate pflags ServerConfig --default-var=defaultServerConfig

type ServerConfig struct {
	HTTPPort             int                   `json:"httpPort" pflag:",On which http port to serve admin"`
	GrpcPort             int                   `json:"grpcPort" pflag:",deprecated"`
	GrpcServerReflection bool                  `json:"grpcServerReflection" pflag:",deprecated"`
	KubeConfig           string                `json:"kube-config" pflag:",Path to kubernetes client config file, default is empty, useful for incluster config."`
	Master               string                `json:"master" pflag:",The address of the Kubernetes API server."`
	Security             ServerSecurityOptions `json:"security"`
	GrpcConfig           GrpcConfig            `json:"grpc"`
	// Deprecated: please use auth.AppAuth.ThirdPartyConfig instead.
	DeprecatedThirdPartyConfig authConfig.ThirdPartyConfigOptions `json:"thirdPartyConfig" pflag:",Deprecated please use auth.appAuth.thirdPartyConfig instead."`

	DataProxy                DataProxyConfig  `json:"dataProxy" pflag:",Defines data proxy configuration."`
	ReadHeaderTimeoutSeconds int              `json:"readHeaderTimeoutSeconds" pflag:",The amount of time allowed to read request headers."`
	KubeClientConfig         KubeClientConfig `json:"kubeClientConfig" pflag:",Configuration to control the Kubernetes client"`
}

type DataProxyConfig struct {
	Upload   DataProxyUploadConfig   `json:"upload" pflag:",Defines data proxy upload configuration."`
	Download DataProxyDownloadConfig `json:"download" pflag:",Defines data proxy download configuration."`
}

type DataProxyDownloadConfig struct {
	MaxExpiresIn config.Duration `json:"maxExpiresIn" pflag:",Maximum allowed expiration duration."`
}

type DataProxyUploadConfig struct {
	MaxSize               resource.Quantity `json:"maxSize" pflag:",Maximum allowed upload size."`
	MaxExpiresIn          config.Duration   `json:"maxExpiresIn" pflag:",Maximum allowed expiration duration."`
	DefaultFileNameLength int               `json:"defaultFileNameLength" pflag:",Default length for the generated file name if not provided in the request."`
	StoragePrefix         string            `json:"storagePrefix" pflag:",Storage prefix to use for all upload requests."`
}

type GrpcConfig struct {
	Port                int  `json:"port" pflag:",On which grpc port to serve admin"`
	ServerReflection    bool `json:"serverReflection" pflag:",Enable GRPC Server Reflection"`
	MaxMessageSizeBytes int  `json:"maxMessageSizeBytes" pflag:",The max size in bytes for incoming gRPC messages"`
}

// KubeClientConfig contains the configuration used by flyteadmin to configure its internal Kubernetes Client.
type KubeClientConfig struct {
	// QPS indicates the maximum QPS to the master from this client.
	// If it's zero, the created RESTClient will use DefaultQPS: 5
	QPS int32 `json:"qps" pflag:",Max QPS to the master for requests to KubeAPI. 0 defaults to 5."`
	// Maximum burst for throttle.
	// If it's zero, the created RESTClient will use DefaultBurst: 10.
	Burst int `json:"burst" pflag:",Max burst rate for throttle. 0 defaults to 10"`
	// The maximum length of time to wait before giving up on a server request. A value of zero means no timeout.
	Timeout config.Duration `json:"timeout" pflag:",Max duration allowed for every request to KubeAPI before giving up. 0 implies no timeout."`
}

type ServerSecurityOptions struct {
	Secure      bool       `json:"secure"`
	Ssl         SslOptions `json:"ssl"`
	UseAuth     bool       `json:"useAuth"`
	AuditAccess bool       `json:"auditAccess"`

	// These options are here to allow deployments where the Flyte UI (Console) is served from a different domain/port.
	// Note that CORS only applies to Admin's API endpoints. The health check endpoint for instance is unaffected.
	// Please obviously evaluate security concerns before turning this on.
	AllowCors bool `json:"allowCors"`
	// Defines origins which are allowed to make CORS requests. This list should _not_ contain "*", as that
	// will make CORS header responses incompatible with the `withCredentials=true` setting.
	AllowedOrigins []string `json:"allowedOrigins"`
	// These are the Access-Control-Request-Headers that the server will respond to.
	// By default, the server will allow Accept, Accept-Language, Content-Language, and Content-Type.
	// DeprecatedUser this setting to add any additional headers which are needed
	AllowedHeaders []string `json:"allowedHeaders"`
}

type SslOptions struct {
	CertificateFile string `json:"certificateFile"`
	KeyFile         string `json:"keyFile"`
}

var defaultServerConfig = &ServerConfig{
	HTTPPort: 8088,
	Security: ServerSecurityOptions{
		AllowCors:      true,
		AllowedHeaders: []string{"Content-Type", "flyte-authorization"},
		AllowedOrigins: []string{"*"},
	},
	GrpcConfig: GrpcConfig{
		Port:             8089,
		ServerReflection: true,
	},
	DataProxy: DataProxyConfig{
		Upload: DataProxyUploadConfig{
			MaxSize:               resource.MustParse("6Mi"),
			MaxExpiresIn:          config.Duration{Duration: time.Hour},
			DefaultFileNameLength: 20,
		},
		Download: DataProxyDownloadConfig{
			MaxExpiresIn: config.Duration{Duration: time.Hour},
		},
	},
	ReadHeaderTimeoutSeconds: 32, // just shy of requestTimeoutUpperBound
	KubeClientConfig: KubeClientConfig{
		QPS:     100,
		Burst:   25,
		Timeout: config.Duration{Duration: 30 * time.Second},
	},
}
var serverConfig = config.MustRegisterSection(SectionKey, defaultServerConfig)

func MustRegisterSubsection(key config.SectionKey, configSection config.Config) config.Section {
	return serverConfig.MustRegisterSection(key, configSection)
}

func GetConfig() *ServerConfig {
	return serverConfig.GetConfig().(*ServerConfig)
}

func SetConfig(s *ServerConfig) {
	if err := serverConfig.SetConfig(s); err != nil {
		panic(err)
	}
}

func (s ServerConfig) GetHostAddress() string {
	return fmt.Sprintf("0.0.0.0:%d", s.HTTPPort)
}

func (s ServerConfig) GetGrpcHostAddress() string {
	if s.GrpcConfig.Port >= 0 {
		return fmt.Sprintf("0.0.0.0:%d", s.GrpcConfig.Port)
	}
	return fmt.Sprintf("0.0.0.0:%d", s.GrpcPort)
}

func init() {
	SetConfig(&ServerConfig{})
}
