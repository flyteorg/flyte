package config

import (
	"os"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
)

//go:generate enumer --type=SecretManagerType --trimprefix=SecretManagerType -json -yaml
//go:generate enumer --type=KVVersion --trimprefix=KVVersion -json -yaml
//go:generate pflags Config --default-var=DefaultConfig

const (
	EmbeddedSecretsFileMountInitContainerName = "init-embedded-secret"
	DefaultSecretEnvVarPrefix                 = "_UNION_"
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
			ImagePullSecrets: ImagePullSecretsConfig{
				Enabled: false,
			},
		},
		ImageBuilderConfig: ImageBuilderConfig{
			ExcludedContainerNames: []string{EmbeddedSecretsFileMountInitContainerName},
		},
		WebhookTimeout:                     30, // default timeout for webhook calls in seconds
		DisableCreateMutatingWebhookConfig: false,
		KubeClientConfig: KubeClientConfig{
			QPS:     100,
			Burst:   25,
			Timeout: config.Duration{Duration: 30 * time.Second},
		},
		SecretEnvVarPrefix: DefaultSecretEnvVarPrefix,
	}

	configSection = config.MustRegisterSection("webhook", DefaultConfig)
	initOnce      sync.Once
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
	ImageBuilderConfig                 ImageBuilderConfig `json:"imageBuilderConfig,omitempty" pflag:"-,"`
	WebhookTimeout                     int32              `json:"webhookTimeout" pflag:",Timeout for webhook calls in seconds. Defaults to 30 seconds."`
	DisableCreateMutatingWebhookConfig bool               `json:"disableCreateMutatingWebhookConfig"`
	KubeClientConfig                   KubeClientConfig   `json:"kubeClientConfig" pflag:",Configuration to control the Kubernetes client used by the webhook"`
	SecretEnvVarPrefix                 string             `json:"secretEnvVarPrefix" pflag:",The prefix for secret environment variables. Used by K8s, Global, and Embedded secret managers. Defaults to _UNION_"`
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
	ImagePullSecrets       ImagePullSecretsConfig       `json:"imagePullSecrets" pflag:",Whether to enable image pull secrets for the webhook pod."`
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
	Namespace        string           `json:"namespace" pflag:",K8s namespace to be used for storing union secrets"`
	KubeClientConfig KubeClientConfig `json:"kubeClientConfig" pflag:",Configuration to control the Kubernetes client used by the secret fetcher for K8s embedded secret manager. Falls back to webhook.kubeClientConfig if not set."`
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

type ImagePullSecretsConfig struct {
	Enabled bool `json:"enabled" pflag:",Whether to enable image pull secrets for the webhook pod."`
}

// KubeClientConfig contains the configuration used by the webhook to configure its internal Kubernetes Client.
type KubeClientConfig struct {
	// QPS indicates the maximum QPS to the master from this client.
	QPS int32 `json:"qps" pflag:",Max QPS to the master for requests to KubeAPI. 0 defaults to 5."`
	// Maximum burst for throttle.
	Burst int `json:"burst" pflag:",Max burst rate for throttle. 0 defaults to 10"`
	// The maximum length of time to wait before giving up on a server request.
	Timeout config.Duration `json:"timeout" pflag:",Max duration allowed for every request to KubeAPI before giving up. 0 implies no timeout."`
}

// ResolveKubeClientConfigs initializes KubeClientConfig with fallback values
// For K8s secretfetcher: if k8sConfig.kubeClientConfig is not set, use webhook-level kubeClientConfig as fallback
func (c *Config) ResolveKubeClientConfigs() {
	// Resolve K8s secret fetcher config with fallback to webhook-level config
	if c.EmbeddedSecretManagerConfig.K8sConfig.KubeClientConfig.QPS == 0 {
		c.EmbeddedSecretManagerConfig.K8sConfig.KubeClientConfig.QPS = c.KubeClientConfig.QPS
	}
	if c.EmbeddedSecretManagerConfig.K8sConfig.KubeClientConfig.Burst == 0 {
		c.EmbeddedSecretManagerConfig.K8sConfig.KubeClientConfig.Burst = c.KubeClientConfig.Burst
	}
	if c.EmbeddedSecretManagerConfig.K8sConfig.KubeClientConfig.Timeout.Duration == 0 {
		c.EmbeddedSecretManagerConfig.K8sConfig.KubeClientConfig.Timeout = c.KubeClientConfig.Timeout
	}
}

func GetConfig() *Config {
	cfg := configSection.GetConfig().(*Config)
	// Ensure initialization happens exactly once, thread-safe
	initOnce.Do(func() {
		cfg.ResolveKubeClientConfigs()
	})
	return cfg
}

func MustRegisterSubsection(key config.SectionKey, section config.Config) config.Section {
	return configSection.MustRegisterSection(key, section)
}
