package config

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
)

var (
	DefaultConfig = &Config{
		SecretName:        "flyte-pod-webhook",
		ServiceName:       "flyte-pod-webhook",
		ServicePort:       443,
		MetricsPrefix:     "flyte:",
		CertDir:           "/etc/webhook/certs",
		LocalCert:         false,
		ListenPort:        9443,
		SecretManagerType: SecretManagerTypeK8s,
		AWSSecretManagerConfig: AWSSecretManagerConfig{
			SidecarImage: "docker.io/amazon/aws-secrets-manager-secret-sidecar:v0.1.4",
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("500Mi"),
					corev1.ResourceCPU:    resource.MustParse("200m"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("500Mi"),
					corev1.ResourceCPU:    resource.MustParse("200m"),
				},
			},
		},
		GCPSecretManagerConfig: GCPSecretManagerConfig{
			SidecarImage: "gcr.io/google.com/cloudsdktool/cloud-sdk:alpine",
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("500Mi"),
					corev1.ResourceCPU:    resource.MustParse("200m"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceMemory: resource.MustParse("500Mi"),
					corev1.ResourceCPU:    resource.MustParse("200m"),
				},
			},
		},
		VaultSecretManagerConfig: VaultSecretManagerConfig{
			Role:      "flyte",
			KVVersion: KVVersion2,
		},
	}

	configSection = config.MustRegisterSection("webhook", DefaultConfig)
)

// SecretManagerType defines which secret manager to use.
type SecretManagerType int

const (
	// SecretManagerTypeGlobal reads env vars and mounted secrets from the webhook pod.
	SecretManagerTypeGlobal SecretManagerType = iota

	// SecretManagerTypeK8s injects K8s volume mounts to mount K8s secrets.
	SecretManagerTypeK8s

	// SecretManagerTypeAWS injects a sidecar to pull secrets from AWS Secret Manager.
	SecretManagerTypeAWS

	// SecretManagerTypeGCP injects a sidecar to pull secrets from GCP Secret Manager.
	SecretManagerTypeGCP

	// SecretManagerTypeVault pulls secrets from Hashicorp Vault.
	SecretManagerTypeVault
)

// KVVersion defines which KV Engine Version to use with VaultSecretManager.
type KVVersion int

const (
	// KVVersion1 refers to unversioned secrets
	KVVersion1 KVVersion = iota
	// KVVersion2 refers to versioned secrets
	KVVersion2
)

type Config struct {
	MetricsPrefix            string                   `json:"metrics-prefix" pflag:",An optional prefix for all published metrics."`
	CertDir                  string                   `json:"certDir" pflag:",Certificate directory to use to write generated certs."`
	LocalCert                bool                     `json:"localCert" pflag:",write certs locally."`
	ListenPort               int                      `json:"listenPort" pflag:",The port to use to listen to webhook calls."`
	ServiceName              string                   `json:"serviceName" pflag:",The name of the webhook service."`
	ServicePort              int32                    `json:"servicePort" pflag:",The port on the service that hosting webhook."`
	SecretName               string                   `json:"secretName" pflag:",Secret name to write generated certs to."`
	SecretManagerType        SecretManagerType        `json:"secretManagerType" pflag:"-,Secret manager type to use if secrets are not found in global secrets."`
	AWSSecretManagerConfig   AWSSecretManagerConfig   `json:"awsSecretManager" pflag:",AWS Secret Manager config."`
	GCPSecretManagerConfig   GCPSecretManagerConfig   `json:"gcpSecretManager" pflag:",GCP Secret Manager config."`
	VaultSecretManagerConfig VaultSecretManagerConfig `json:"vaultSecretManager" pflag:",Vault Secret Manager config."`
}

func (c Config) ExpandCertDir() string {
	return os.ExpandEnv(c.CertDir)
}

type AWSSecretManagerConfig struct {
	SidecarImage string                      `json:"sidecarImage" pflag:",Specifies the sidecar docker image to use"`
	Resources    corev1.ResourceRequirements `json:"resources" pflag:"-,Specifies resource requirements for the init container."`
}

type GCPSecretManagerConfig struct {
	SidecarImage string                      `json:"sidecarImage" pflag:",Specifies the sidecar docker image to use"`
	Resources    corev1.ResourceRequirements `json:"resources" pflag:"-,Specifies resource requirements for the init container."`
}

type VaultSecretManagerConfig struct {
	Role        string            `json:"role" pflag:",Specifies the vault role to use"`
	KVVersion   KVVersion         `json:"kvVersion" pflag:"-,The KV Engine Version. Defaults to 2."`
	Annotations map[string]string `json:"annotations" pflag:"-,Annotations to add to user task pod."`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
