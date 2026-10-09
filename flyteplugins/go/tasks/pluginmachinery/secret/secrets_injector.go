package secret

import (
	"context"
	"fmt"
	"slices"

	corev1 "k8s.io/api/core/v1"
	k8sRuntime "k8s.io/apimachinery/pkg/runtime"
	ctrlcache "sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secretmanager"
	stdlibCache "github.com/flyteorg/flyte/v2/flytestdlib/cache"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

type SecretsInjector interface {
	Type() config.SecretManagerType
	Inject(ctx context.Context, secrets *core.Secret, p *corev1.Pod) (newP *corev1.Pod, injected bool, err error)
	InvalidateCache(ctx context.Context, org, domain, project, secretName string)
}

func newSecretsInjector(
	ctx context.Context,
	secretManagerType config.SecretManagerType,
	webhookConfig *config.Config,
	globalSecretManagerConfig *secretmanager.Config,
	podNamespace string,
	limitNamespace string,
	scope promutils.Scope,
) (SecretsInjector, error) {
	switch secretManagerType {
	case config.SecretManagerTypeGlobal:
		return NewGlobalSecrets(secretmanager.NewFileEnvSecretManager(globalSecretManagerConfig), webhookConfig), nil
	case config.SecretManagerTypeK8s:
		return NewK8sSecretsInjector(webhookConfig), nil
	case config.SecretManagerTypeAWS:
		return NewAWSSecretManagerInjector(webhookConfig.AWSSecretManagerConfig), nil
	case config.SecretManagerTypeGCP:
		return NewGCPSecretManagerInjector(webhookConfig.GCPSecretManagerConfig), nil
	case config.SecretManagerTypeVault:
		return NewVaultSecretManagerInjector(webhookConfig.VaultSecretManagerConfig), nil
	case config.SecretManagerTypeEmbedded:
		kubeConfig, err := resolveKubeConfig(ctx)
		if err != nil {
			logger.Errorf(ctx, "Failed to get kubernetes config: %v", err)
			return nil, fmt.Errorf("failed to start secret manager service due to %v", err)
		}
		if webhookConfig.KubeClientConfig.QPS > 0 {
			kubeConfig.QPS = float32(webhookConfig.KubeClientConfig.QPS)
		}
		if webhookConfig.KubeClientConfig.Burst > 0 {
			kubeConfig.Burst = webhookConfig.KubeClientConfig.Burst
		}
		if webhookConfig.KubeClientConfig.Timeout.Duration > 0 {
			kubeConfig.Timeout = webhookConfig.KubeClientConfig.Timeout.Duration
		}
		// Initialize controller-runtime client
		ctrlRuntimeScheme := k8sRuntime.NewScheme()
		if err := corev1.AddToScheme(ctrlRuntimeScheme); err != nil {
			logger.Errorf(ctx, "Failed to add core v1 to scheme: %v", err)
			return nil, fmt.Errorf("failed to add core v1 to scheme: %w", err)
		}

		// The k8s client backs the image-pull-secret path (reference secret lookup + mirroring
		// into the pod namespace) and, for the K8s embedded type, the stored-secret check. Both
		// are on the admission hot path, so reads go through a Secret informer — but only one
		// scoped to the namespaces those reads hit. When neither feature is in use the client
		// is direct and no informer (and no cluster-wide Secret watch) is started.
		clientOpts := client.Options{Scheme: ctrlRuntimeScheme}
		if useInformer, namespaces := secretInformerNamespaces(webhookConfig, podNamespace, limitNamespace); useInformer {
			cacheOpts := ctrlcache.Options{Scheme: ctrlRuntimeScheme}
			if len(namespaces) > 0 {
				cacheOpts.DefaultNamespaces = make(map[string]ctrlcache.Config, len(namespaces))
				for _, ns := range namespaces {
					cacheOpts.DefaultNamespaces[ns] = ctrlcache.Config{}
				}
			}
			logger.Infof(ctx, "Starting Secret informer for namespaces %v (empty = all)", namespaces)
			secretInformerCache, err := ctrlcache.New(kubeConfig, cacheOpts)
			if err != nil {
				return nil, fmt.Errorf("failed to create informer cache: %w", err)
			}

			// Explicitly register the Secret informer so the cache only watches Secrets and so
			// WaitForCacheSync below actually blocks on the initial Secret list.
			if _, err := secretInformerCache.GetInformer(ctx, &corev1.Secret{}); err != nil {
				return nil, fmt.Errorf("failed to register Secret informer: %w", err)
			}

			go func() {
				if err := secretInformerCache.Start(ctx); err != nil {
					logger.Errorf(ctx, "secret informer cache stopped: %v", err)
				}
			}()
			if !secretInformerCache.WaitForCacheSync(ctx) {
				return nil, fmt.Errorf("secret informer cache failed to sync")
			}

			clientOpts.Cache = &client.CacheOptions{Reader: secretInformerCache}
		}

		ctrlRuntimeClient, err := client.New(kubeConfig, clientOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to create controller-runtime client: %w", err)
		}

		var secretFetchers []SecretFetcher
		secretFetcher, err := NewSecretFetcher(ctx, webhookConfig.EmbeddedSecretManagerConfig)
		if err != nil {
			return nil, err
		}

		secretFetchers = append(secretFetchers, secretFetcher)

		cacheConfig := stdlibCache.GetConfig()
		cacheFactory, err := stdlibCache.NewFactory(ctx, cacheConfig,
			nil, scope.NewSubScope("secret_cache"))
		if err != nil {
			logger.Errorf(ctx, "Failed to create cache factory: %v", err)
			return nil, fmt.Errorf("failed to create cache factory: %w", err)
		}

		secretCache, err := cacheFactory.New[SecretValue]("secret_cache", cacheConfig.Type, nil, scope.NewSubScope("secret_value"))
		if err != nil {
			logger.Errorf(ctx, "Failed to create secret cache: %v", err)
			return nil, fmt.Errorf("failed to create secret cache: %w", err)
		}

		return NewEmbeddedSecretManagerInjector(webhookConfig.EmbeddedSecretManagerConfig, secretFetchers,
			ctrlRuntimeClient, podNamespace, secretCache, webhookConfig), nil
	case config.SecretManagerTypeAzure:
		return NewAzureSecretManagerInjector(webhookConfig.AzureSecretManagerConfig), nil
	default:
		return nil, fmt.Errorf("unrecognized secret manager type [%v]", secretManagerType)
	}
}

// AllNamespaces is the limitNamespace value meaning "no namespace limit". An empty
// limitNamespace means the same.
const AllNamespaces = "all"

// secretInformerNamespaces decides whether the embedded secret manager's k8s client should be
// backed by a Secret informer and, if so, which namespaces it must watch. An empty namespace list
// means all namespaces.
//
//   - Image pull secrets read the reference secret from podNamespace and get/create the mirrored
//     secret in each task pod's namespace. Pods are confined to limitNamespace when it is set, so
//     the informer watches {limitNamespace, podNamespace}; otherwise it is cluster-wide.
//   - The K8s embedded type checks the stored secret in K8sConfig.Namespace only.
//
// With neither feature in use, no informer is needed.
func secretInformerNamespaces(webhookConfig *config.Config, podNamespace, limitNamespace string) (bool, []string) {
	embeddedCfg := webhookConfig.EmbeddedSecretManagerConfig
	imagePull := embeddedCfg.ImagePullSecrets.Enabled
	k8sType := embeddedCfg.Type == config.EmbeddedSecretManagerTypeK8s
	if !imagePull && !k8sType {
		return false, nil
	}

	unlimited := limitNamespace == "" || limitNamespace == AllNamespaces
	if imagePull && unlimited {
		return true, nil
	}

	var namespaces []string
	add := func(ns string) {
		if ns != "" && !slices.Contains(namespaces, ns) {
			namespaces = append(namespaces, ns)
		}
	}
	if imagePull {
		add(limitNamespace)
		add(podNamespace)
	}
	if k8sType {
		add(embeddedCfg.K8sConfig.Namespace)
	}
	if len(namespaces) == 0 {
		// K8s type with no namespace configured: fall back to a direct client rather than
		// silently watching every Secret in the cluster.
		return false, nil
	}
	return true, namespaces
}
