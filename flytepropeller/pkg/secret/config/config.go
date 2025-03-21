package config

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flytestdlib/config"
)

//go:generate enumer --type=SecretManagerType --trimprefix=SecretManagerType -json -yaml
//go:generate enumer --type=KVVersion --trimprefix=KVVersion -json -yaml
//go:generate pflags Config --default-var=DefaultConfig

const (
	EmbeddedSecretsFileMountInitContainerName = "init-embedded-secret"
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
		AzureSecretManagerConfig: AzureSecretManagerConfig{
			SidecarImage: "mcr.microsoft.com/azure-cli:cbl-mariner2.0",
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
		EmbeddedSecretManagerConfig: EmbeddedSecretManagerConfig{
			FileMountInitContainer: FileMountInitContainerConfig{
				Image: "busybox:1.28",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("100Mi"),
						corev1.ResourceCPU:    resource.MustParse("100m"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceMemory: resource.MustParse("100Mi"),
						corev1.ResourceCPU:    resource.MustParse("100m"),
					},
				},
				ContainerName: EmbeddedSecretsFileMountInitContainerName,
			},
		},
		ImageBuilderConfig: ImageBuilderConfig{
			ExcludedContainerNames: []string{EmbeddedSecretsFileMountInitContainerName},
		},
	}

	configSection = config.MustRegisterSection("webhook", DefaultConfig)
)

// SecretManagerType defines which secret manager to use.
type SecretManagerType int

const (
	// SecretManagerTypeGlobal defines a global secret manager that can read env vars and mounted secrets to the webhook
	// pod.
	SecretManagerTypeGlobal SecretManagerType = iota

	// SecretManagerTypeK8s defines a secret manager webhook that injects K8s volume mounts to mount K8s secrets.
	SecretManagerTypeK8s

	// SecretManagerTypeAWS defines a secret manager webhook that injects a side car to pull secrets from AWS Secret
	// Manager and mount them to a local file system (in memory) and share that mount with other containers in the pod.
	SecretManagerTypeAWS

	// SecretManagerTypeGCP defines a secret manager webhook that injects a side car to pull secrets from GCP Secret
	// Manager and mount them to a local file system (in memory) and share that mount with other containers in the pod.
	SecretManagerTypeGCP

	// SecretManagerTypeVault defines a secret manager webhook that pulls secrets from Hashicorp Vault.
	SecretManagerTypeVault

	// SecretManagerTypeEmbedded defines an embedded secret manager webhook that pulls secrets from the configured secrets manager.
	// Without using sidecar. This type directly calls into the secrets manager for the configured provider directly.
	// Currently supported only for AWS.
	SecretManagerTypeEmbedded

	// SecretManagerTypeAzure defines a secret manager webhook that injects a side car to pull secrets from Azure Key Vault
	SecretManagerTypeAzure
)

// Defines with KV Engine Version to use with VaultSecretManager - https://www.vaultproject.io/docs/secrets/kv#kv-secrets-engine
type KVVersion int

const (
	// KV v1 refers to unversioned secrets
	KVVersion1 KVVersion = iota
	// KV v2 refers to versioned secrets
	KVVersion2
)

type Config struct {
	MetricsPrefix string `json:"metrics-prefix" pflag:",An optional prefix for all published metrics."`
	CertDir       string `json:"certDir" pflag:",Certificate directory to use to write generated certs. Defaults to /etc/webhook/certs/"`
	LocalCert     bool   `json:"localCert" pflag:",write certs locally. Defaults to false"`
	ListenPort    int    `json:"listenPort" pflag:",The port to use to listen to webhook calls. Defaults to 9443"`
	ServiceName   string `json:"serviceName" pflag:",The name of the webhook service."`
	ServicePort   int32  `json:"servicePort" pflag:",The port on the service that hosting webhook."`
	SecretName    string `json:"secretName" pflag:",Secret name to write generated certs to."`
	// Deprecated: use SecretManagerTypes instead.
	SecretManagerType           SecretManagerType           `json:"secretManagerType" pflag:"-,Deprecated. Secret manager type to use if secrets are not found in global secrets. Ignored if secretManagerTypes is set."`
	SecretManagerTypes          []SecretManagerType         `json:"secretManagerTypes" pflag:"-,List of secret manager types to use if secrets are not found in global secrets. In order of preference. Overrides secretManagerType if set."`
	AWSSecretManagerConfig      AWSSecretManagerConfig      `json:"awsSecretManager" pflag:",AWS Secret Manager config."`
	GCPSecretManagerConfig      GCPSecretManagerConfig      `json:"gcpSecretManager" pflag:",GCP Secret Manager config."`
	VaultSecretManagerConfig    VaultSecretManagerConfig    `json:"vaultSecretManager" pflag:",Vault Secret Manager config."`
	EmbeddedSecretManagerConfig EmbeddedSecretManagerConfig `json:"embeddedSecretManagerConfig" pflag:",Embedded Secret Manager config without sidecar and which calls into the supported providers directly."`
	AzureSecretManagerConfig    AzureSecretManagerConfig    `json:"azureSecretManager" pflag:",Azure Secret Manager config."`

	// Ignore PFlag for Image Builder
	ImageBuilderConfig ImageBuilderConfig `json:"imageBuilderConfig,omitempty" pflag:"-,"`
}

//go:generate enumer --type=EmbeddedSecretManagerType -json -yaml -trimprefix=EmbeddedSecretManagerType
type EmbeddedSecretManagerType uint8

const (
	EmbeddedSecretManagerTypeAWS EmbeddedSecretManagerType = iota
	EmbeddedSecretManagerTypeGCP
	EmbeddedSecretManagerTypeAzure
	EmbeddedSecretManagerTypeK8s
)

type EmbeddedSecretManagerConfig struct {
	Type                   EmbeddedSecretManagerType    `json:"type" pflags:"-,Type of embedded secret manager to initialize"`
	AWSConfig              AWSConfig                    `json:"awsConfig" pflag:",Config for AWS settings"`
	GCPConfig              GCPConfig                    `json:"gcpConfig" pflag:",Config for GCP settings"`
	AzureConfig            AzureConfig                  `json:"azureConfig" pflag:",Config for Azure settings"`
	K8sConfig              K8sConfig                    `json:"k8sConfig" pflag:",Config for K8s settings"`
	FileMountInitContainer FileMountInitContainerConfig `json:"fileMountInitContainer" pflag:",Init container configuration to use for mounting secrets as files."`
}

type AWSConfig struct {
	Region string `json:"region" pflag:",AWS region"`
}

type GCPConfig struct {
	Project string `json:"project" pflag:",GCP project to be used for secret manager"`
}

type AzureConfig struct {
	VaultURI string `json:"vaultURI" pflag:",Azure Vault URI"`
}

type K8sConfig struct {
	Namespace string `json:"namespace" pflag:",K8s namespace to be used for storing union secrets"`
}

type FileMountInitContainerConfig struct {
	Image         string                      `json:"image" pflag:",Specifies init container image to use for mounting secrets as files."`
	Resources     corev1.ResourceRequirements `json:"resources" pflag:"-,Specifies resource requirements for the init container."`
	ContainerName string                      `json:"containerName" pflag:",Specifies the name of the init container that mounts secrets as files."`
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

type AzureSecretManagerConfig struct {
	SidecarImage string                      `json:"sidecarImage" pflag:",Specifies the sidecar docker image to use"`
	Resources    corev1.ResourceRequirements `json:"resources" pflag:"-,Specifies resource requirements for the init container."`
}

type VaultSecretManagerConfig struct {
	Role        string            `json:"role" pflag:",Specifies the vault role to use"`
	KVVersion   KVVersion         `json:"kvVersion" pflag:"-,DEPRECATED! Use the GroupVersion field of the Secret request instead. The KV Engine Version. Defaults to 2. Use 1 for unversioned secrets. Refer to - https://www.vaultproject.io/docs/secrets/kv#kv-secrets-engine."`
	Annotations map[string]string `json:"annotations" pflag:"-,Annotation to be added to user task pod. The annotation can also be used to override default annotations added by Flyte. Useful to customize Vault integration (https://developer.hashicorp.com/vault/docs/platform/k8s/injector/annotations)"`
}

type HostnameReplacement struct {
	Existing            string `json:"existing" pflag:",The existing hostname to replace"`
	Replacement         string `json:"replacement" pflag:",The replacement hostname"`
	DisableVerification bool   `json:"disableVerification" pflag:",Allow disabling URI verification for development environments"`
}

type ImageBuilderConfig struct {
	Enabled                bool                 `json:"enabled"`
	HostnameReplacement    HostnameReplacement  `json:"hostnameReplacement"`
	LabelSelector          metav1.LabelSelector `json:"labelSelector"`
	ExcludedContainerNames []string             `json:"excludedContainerNames"`
	ExcludedImagePrefixes  []string             `json:"excludedImagePrefixes"`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
