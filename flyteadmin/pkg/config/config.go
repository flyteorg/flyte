package config

import (
	"fmt"

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
}

type GrpcConfig struct {
	Port                int  `json:"port" pflag:",On which grpc port to serve admin"`
	ServerReflection    bool `json:"serverReflection" pflag:",Enable GRPC Server Reflection"`
	MaxMessageSizeBytes int  `json:"maxMessageSizeBytes" pflag:",The max size in bytes for incoming gRPC messages"`
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
	HTTPPort:   8088,
	KubeConfig: "$HOME/.kube/config",
	Security: ServerSecurityOptions{
		AllowCors:      true,
		AllowedHeaders: []string{"Content-Type", "flyte-authorization"},
		AllowedOrigins: []string{"*"},
	},
	GrpcConfig: GrpcConfig{
		Port:             8089,
		ServerReflection: true,
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
	return fmt.Sprintf(":%d", s.HTTPPort)
}

func (s ServerConfig) GetGrpcHostAddress() string {
	if s.GrpcConfig.Port >= 0 {
		return fmt.Sprintf(":%d", s.GrpcConfig.Port)
	}
	return fmt.Sprintf(":%d", s.GrpcPort)
}

func init() {
	SetConfig(&ServerConfig{})
}
