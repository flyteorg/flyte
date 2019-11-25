package config

import (
	"fmt"

	config2 "github.com/lyft/flyteadmin/pkg/auth/config"
	"github.com/lyft/flytestdlib/config"
)

const SectionKey = "server"

//go:generate pflags ServerConfig --default-var=defaultServerConfig

type ServerConfig struct {
	HTTPPort             int                   `json:"httpPort" pflag:",On which http port to serve admin"`
	GrpcPort             int                   `json:"grpcPort" pflag:",On which grpc port to serve admin"`
	GrpcServerReflection bool                  `json:"grpcServerReflection" pflag:",Enable GRPC Server Reflection"`
	KubeConfig           string                `json:"kube-config" pflag:",Path to kubernetes client config file."`
	Master               string                `json:"master" pflag:",The address of the Kubernetes API server."`
	Security             ServerSecurityOptions `json:"security"`
}

type ServerSecurityOptions struct {
	Secure  bool                 `json:"secure"`
	Ssl     SslOptions           `json:"ssl"`
	UseAuth bool                 `json:"useAuth"`
	Oauth   config2.OAuthOptions `json:"oauth"`

	// These options are here to allow deployments where the Flyte UI (Console) is served from a different domain/port.
	// Note that CORS only applies to Admin's API endpoints. The health check endpoint for instance is unaffected.
	// Please obviously evaluate security concerns before turning this on.
	AllowCors bool `json:"allowCors"`
	// TODO: Go through the gorilla library and resolve singular vs plural. It should be singular, but what else is the library doing?
	AllowedOrigins []string `json:"allowedOrigins"`
	// These are the Access-Control-Request-Headers that the server will respond to
	AllowedHeaders []string `json:"allowedHeaders"`
}

type SslOptions struct {
	CertificateFile string `json:"certificateFile"`
	KeyFile         string `json:"keyFile"`
}

var defaultServerConfig = &ServerConfig{
	Security: ServerSecurityOptions{
		Oauth: config2.OAuthOptions{
			// Please see the comments in this struct's definition for more information
			HTTPAuthorizationHeader: "placeholder",
			GrpcAuthorizationHeader: "flyte-authorization",
		},
	},
}
var serverConfig = config.MustRegisterSection(SectionKey, defaultServerConfig)

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
	return fmt.Sprintf(":%d", s.GrpcPort)
}

func init() {
	SetConfig(&ServerConfig{})
}
