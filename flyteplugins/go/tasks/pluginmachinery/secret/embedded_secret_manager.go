package secret

import (
	"context"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	goCache "github.com/eko/gocache/lib/v4/cache"
	slices2 "golang.org/x/exp/slices"
	corev1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	stdlibErrors "github.com/flyteorg/flyte/v2/flytestdlib/errors"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

const (
	// Static name of the volume used for mounting secrets with file mount requirement.
	EmbeddedSecretsFileMountVolumeName = "embedded-secret-vol" // #nosec G101
	EmbeddedSecretsFileMountPath       = "/etc/flyte/secrets"  // #nosec G101

	// Name of the environment variable in the init container used for mounting secrets as files.
	// This environment variable is used to pass secret names and values to the init container.
	// The init container then reads its value and writes secrets to files.
	// Format of this environment variable's value:
	// 		secret_name1=base64_encoded_secret_value1
	// 		secret_name2=base64_encoded_secret_value2
	//		...
	EmbeddedSecretsFileMountInitContainerEnvVariableName = "SECRETS" // #nosec G101

	ProjectLabel      = "project"
	DomainLabel       = "domain"
	OrganizationLabel = "organization"
	EmptySecretScope  = ""

	// All of the namespace, group and key cannot contain '/' so it is safe to use '/' as a delimiter.
	NamespaceGroupKeyDelimiter = "/"
	rawK8sSecretsStorageFormat = valueFormatter + NamespaceGroupKeyDelimiter + valueFormatter + NamespaceGroupKeyDelimiter + valueFormatter

	SecretNotFoundErrorFormat                                   = "secret %v not found in the secret manager"                                // #nosec G101
	SecretReadFailureErrorFormat                                = "secret %v failed to be read from secret manager"                          // #nosec G101
	SecretNilErrorFormat                                        = "secret %v read as empty from the secret manager"                          // #nosec G101
	SecretRequirementsErrorFormat                               = "secret read requirements not met due to empty %v field in the pod labels" // #nosec G101
	SecretSecretNotFoundAcrossAllScopes                         = "secret not found across all scopes"                                       // #nosec G101
	ErrCodeSecretRequirementsError       stdlibErrors.ErrorCode = "SecretRequirementsError"                                                  // #nosec G101
	ErrCodeSecretNotFound                stdlibErrors.ErrorCode = "SecretNotFound"                                                           // #nosec G101
	ErrCodeSecretNotFoundAcrossAllScopes stdlibErrors.ErrorCode = "SecretNotFoundAcrossAllScopes"                                            // #nosec G101
	ErrCodeSecretReadFailure             stdlibErrors.ErrorCode = "SecretReadFailure"                                                        // #nosec G101
	ErrCodeSecretNil                     stdlibErrors.ErrorCode = "SecretNil"                                                                // #nosec G101
)

// AWSSecretManagerInjector allows injecting of secrets from AWS Secret Manager as environment variable. It uses AWS-provided SideCar
// as an init-container to download the secret and save it to a local volume shared with all other containers in the pod.
// It supports multiple secrets to be mounted but that will result into adding an init container for each secret.
// The role/serviceaccount used to run the Pod must have permissions to pull the secret from AWS Secret Manager.
// Otherwise, the Pod will fail with an init-error.
// Files will be mounted on /etc/flyte/secrets/<SecretGroup>/<SecretKey>
type EmbeddedSecretManagerInjector struct {
	cfg                config.EmbeddedSecretManagerConfig
	secretFetchers     []SecretFetcher
	k8sClient          client.Client
	referenceNamespace string
	secretCache        goCache.CacheInterface[SecretValue]
	parentCfg          *config.Config
}

func (i *EmbeddedSecretManagerInjector) Type() config.SecretManagerType {
	return config.SecretManagerTypeEmbedded
}

func GetSecretID(secretKey string, labels map[string]string) (string, error) {
	return EncodeSecretName(labels[OrganizationLabel], labels[DomainLabel], labels[ProjectLabel], secretKey), nil
}

func validateRequiredFieldsExist(labels map[string]string) error {
	if labels[OrganizationLabel] == "" {
		return stdlibErrors.Errorf(ErrCodeSecretRequirementsError, fmt.Sprintf(SecretRequirementsErrorFormat, OrganizationLabel))
	}
	if labels[ProjectLabel] == "" {
		return stdlibErrors.Errorf(ErrCodeSecretRequirementsError, fmt.Sprintf(SecretRequirementsErrorFormat, ProjectLabel))
	}
	if labels[DomainLabel] == "" {
		return stdlibErrors.Errorf(ErrCodeSecretRequirementsError, fmt.Sprintf(SecretRequirementsErrorFormat, DomainLabel))
	}
	return nil
}

func deriveSecretNameComponents(secret *core.Secret, pod *corev1.Pod) (*SecretNameComponents, error) {
	err := validateRequiredFieldsExist(pod.Labels)
	if err != nil {
		return nil, err
	}

	return &SecretNameComponents{
		Project: pod.Labels[ProjectLabel],
		Domain:  pod.Labels[DomainLabel],
		Org:     pod.Labels[OrganizationLabel],
		Name:    secret.Key,
	}, nil
}

func encodeSecretName(components *SecretNameComponents) string {
	return EncodeSecretName(components.Org, components.Domain, components.Project, components.Name)
}

func (i *EmbeddedSecretManagerInjector) lookUpSecret(ctx context.Context, components *SecretNameComponents) (*SecretValue, string, error) {
	componentsCopy := *components // Create a copy to avoid modifying the original

	// Attempt to lookup from the cache
	projectDomainScopedSecret := encodeSecretName(&componentsCopy)
	projectDomainScopedImagePullSecretName := ToImagePullK8sName(componentsCopy)
	if cachedValue, err := i.secretCache.Get(ctx, projectDomainScopedSecret); err == nil {
		logger.Debugf(ctx, "Found secret [%s] in cache.", projectDomainScopedSecret)
		return &cachedValue, projectDomainScopedImagePullSecretName, nil
	}

	componentsCopy.Project = EmptySecretScope
	domainScopedSecret := encodeSecretName(&componentsCopy)
	domainScopedImagePullSecretName := ToImagePullK8sName(componentsCopy)
	if cachedValue, err := i.secretCache.Get(ctx, domainScopedSecret); err == nil {
		logger.Debugf(ctx, "Found secret [%s] in cache.", domainScopedSecret)
		return &cachedValue, domainScopedImagePullSecretName, nil
	}

	componentsCopy.Domain = EmptySecretScope
	orgScopedSecret := encodeSecretName(&componentsCopy)
	orgScopedImagePullSecretName := ToImagePullK8sName(componentsCopy)
	if cachedValue, err := i.secretCache.Get(ctx, orgScopedSecret); err == nil {
		logger.Debugf(ctx, "Found secret [%s] in cache.", orgScopedSecret)
		return &cachedValue, orgScopedImagePullSecretName, nil
	}

	logger.Infof(ctx, "Secret [%s] not found in cache. Fetching from secret fetchers.", components)

	for _, secretFetcher := range i.secretFetchers {
		// Fetch project-domain scoped secret
		secretValue, err := secretFetcher.GetSecretValue(ctx, projectDomainScopedSecret)
		if err == nil {
			if err := i.secretCache.Set(ctx, projectDomainScopedSecret, *secretValue); err != nil {
				logger.Warnf(ctx, "Failed to cache secret [%s]: %v", projectDomainScopedSecret, err)
			}

			return secretValue, projectDomainScopedImagePullSecretName, nil
		}

		if !stdlibErrors.IsCausedBy(err, ErrCodeSecretNotFound) {
			return nil, "", err
		}

		// Fetch domain scoped secret
		secretValue, err = secretFetcher.GetSecretValue(ctx, domainScopedSecret)
		if err == nil {
			if err := i.secretCache.Set(ctx, domainScopedSecret, *secretValue); err != nil {
				logger.Warnf(ctx, "Failed to cache secret [%s]: %v", domainScopedSecret, err)
			}

			return secretValue, domainScopedImagePullSecretName, nil
		}

		if !stdlibErrors.IsCausedBy(err, ErrCodeSecretNotFound) {
			return nil, "", err
		}

		// Fetch organization scoped secret
		secretValue, err = secretFetcher.GetSecretValue(ctx, orgScopedSecret)
		if err == nil {
			if err := i.secretCache.Set(ctx, orgScopedSecret, *secretValue); err != nil {
				logger.Warnf(ctx, "Failed to cache secret [%s]: %v", orgScopedSecret, err)
			}

			return secretValue, orgScopedImagePullSecretName, err
		}

		if !stdlibErrors.IsCausedBy(err, ErrCodeSecretNotFound) {
			return nil, "", err
		}
	}

	return nil, "", stdlibErrors.Errorf(ErrCodeSecretNotFoundAcrossAllScopes, SecretSecretNotFoundAcrossAllScopes)
}

// addImagePullSecretToPod adds an image pull secret to a pod if it doesn't already exist
func (i *EmbeddedSecretManagerInjector) addImagePullSecretToPod(
	ctx context.Context,
	pod *corev1.Pod,
	secret *corev1.Secret,
) (*corev1.Pod, error) {
	// Check if there is a Secret in the same namespace of the Pod that has the same labels of the Secret parameter

	mirroredSecret := &corev1.Secret{}
	err := i.k8sClient.Get(ctx, types.NamespacedName{Name: secret.GetName(), Namespace: pod.GetNamespace()}, mirroredSecret)
	if err != nil && k8sError.IsNotFound(err) {
		// Can't deep copy existing secret because it is not in the same namespace and has fields not allowed in Create
		newSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:        secret.GetName(),
				Namespace:   pod.GetNamespace(),
				Labels:      secret.Labels,
				Annotations: secret.Annotations,
			},
			Type:       secret.Type,
			Data:       secret.Data,
			StringData: secret.StringData,
		}
		err = i.k8sClient.Create(ctx, newSecret)
		mirroredSecret = newSecret
		if err != nil {
			logger.Errorf(ctx, "Failed to create mirrored secret [%s] in namespace [%s]: %v", secret.GetName(), pod.GetNamespace(), err)
			return pod, fmt.Errorf("failed to create mirrored secret [%s] in namespace [%s]: %w", secret.GetName(), pod.GetNamespace(), err)
		}
	} else if err != nil {
		logger.Errorf(ctx, "Failed to check for mirrored secret [%s] in namespace [%s]: %v", secret.GetName(), pod.GetNamespace(), err)
		return pod, fmt.Errorf("failed to check for mirrored secret [%s] in namespace [%s]: %w", secret.GetName(), pod.GetNamespace(), err)
	}

	// Now add the image pull secret reference to the pod
	imagePullSecretRef := corev1.LocalObjectReference{
		Name: mirroredSecret.GetName(),
	}

	// Check if the secret reference already exists to avoid duplicates
	secretExists := false
	if pod.Spec.ImagePullSecrets != nil {
		for _, existingSecret := range pod.Spec.ImagePullSecrets {
			if existingSecret.Name == mirroredSecret.GetName() {
				secretExists = true
				break
			}
		}
	}

	// Add the image pull secret reference if it doesn't already exist
	if !secretExists {
		if pod.Spec.ImagePullSecrets == nil {
			pod.Spec.ImagePullSecrets = make([]corev1.LocalObjectReference, 0)
		}

		pod.Spec.ImagePullSecrets = append(pod.Spec.ImagePullSecrets, imagePullSecretRef)
		slices2.SortStableFunc(pod.Spec.ImagePullSecrets, func(e corev1.LocalObjectReference, e2 corev1.LocalObjectReference) int {
			return strings.Compare(e.Name, e2.Name)
		})

		logger.Infof(ctx, "Added image pull secret [%s] to pod [%s/%s]",
			mirroredSecret.GetName(), pod.Namespace, pod.Name)
		return pod, nil
	} else {
		logger.Debugf(ctx, "Image pull secret [%s] already exists in pod [%s/%s]",
			mirroredSecret.GetName(), pod.Namespace, pod.Name)
		return pod, nil
	}
}

// findAndAddImagePullSecret finds image pull secrets for the given secret and adds them to the pod
func (i *EmbeddedSecretManagerInjector) findAndAddImagePullSecret(
	ctx context.Context,
	k8sSecretName string,
	pod *corev1.Pod,
) (*corev1.Pod, error) {
	referenceSecret := &corev1.Secret{}
	// Pass an empty slice of client.GetOption instead of nil
	err := i.k8sClient.Get(ctx, types.NamespacedName{Name: k8sSecretName, Namespace: i.referenceNamespace}, referenceSecret)
	if err != nil && !k8sError.IsNotFound(err) {
		return pod, fmt.Errorf("failed to get reference secret [%s]: %w", k8sSecretName, err)
	}

	if k8sError.IsNotFound(err) {
		logger.Debugf(ctx, "No reference Kubernetes secret found [%s]. Assuming not image pull secret.", k8sSecretName)
		return pod, nil
	}

	return i.addImagePullSecretToPod(ctx, pod, referenceSecret)
}

func (i *EmbeddedSecretManagerInjector) Inject(
	ctx context.Context,
	secret *core.Secret,
	pod *corev1.Pod,
) (*corev1.Pod, bool /*injected*/, error) {
	if len(secret.Key) == 0 {
		return pod, false, fmt.Errorf("EmbeddedSecretManager requires key to be set. Secret: [%v]", secret)
	}

	secretNameComponents, err := deriveSecretNameComponents(secret, pod)
	if err != nil {
		return pod, false, err
	}

	secretValue, k8sImagePullSecretName, err := i.lookUpSecret(ctx, secretNameComponents)
	if err != nil {
		return pod, false, err
	}

	switch secret.MountRequirement {
	case core.Secret_ANY:
		fallthrough
	case core.Secret_ENV_VAR:
		var stringValue string
		if secretValue.StringValue != "" {
			stringValue = secretValue.StringValue
		} else {
			// GCP secrets store values as binary only. This means a secret could be
			// defined as a file, but mounted as an environment variable.
			// We could fail this path for AWS, but for consistent behaviour between
			// AWS and GCP we will allow this path for AWS as well.
			if !utf8.Valid(secretValue.BinaryValue) {
				return pod, false, fmt.Errorf(
					"secret %q is attempted to be mounted as an environment variable, "+
						"but has a binary value that is not a valid UTF-8 string; mount "+
						"as a file instead", secret.Key)
			}
			stringValue = string(secretValue.BinaryValue)
		}
		i.injectAsEnvVar(secret, stringValue, pod)
	case core.Secret_FILE:
		secretBytes := secretValue.BinaryValue
		if len(secretValue.StringValue) > 0 {
			secretBytes = []byte(secretValue.StringValue)
		}
		i.injectAsFile(secret, secretBytes, pod)
	default:
		err := fmt.Errorf("unrecognized mount requirement [%v] for secret [%v]", secret.MountRequirement.String(), secret.Key)
		logger.Error(ctx, err)
		return pod, false, err
	}

	if i.cfg.ImagePullSecrets.Enabled {
		pod, err = i.findAndAddImagePullSecret(ctx, k8sImagePullSecretName, pod)
		if err != nil {
			logger.Errorf(ctx, "Failed to add image pull secret [%s]: %v", secretNameComponents.Name, err)
			return pod, false, err
		}
	}

	return pod, true, nil
}

func (i *EmbeddedSecretManagerInjector) injectAsEnvVar(secret *core.Secret, secretValue string, pod *corev1.Pod) {
	envVars := []corev1.EnvVar{
		{
			Name:  SecretEnvVarPrefix,
			Value: i.parentCfg.SecretEnvVarPrefix,
		},
		{
			Name:  i.parentCfg.SecretEnvVarPrefix + strings.ToUpper(secret.GetKey()),
			Value: secretValue,
		},
	}
	if secret.GetEnvVar() != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  secret.EnvVar,
			Value: secretValue,
		})
	}
	pod.Spec.InitContainers = AppendEnvVars(pod.Spec.InitContainers, envVars...)
	pod.Spec.Containers = AppendEnvVars(pod.Spec.Containers, envVars...)
}

func (i *EmbeddedSecretManagerInjector) injectAsFile(secret *core.Secret, secretValue []byte, pod *corev1.Pod) {
	initContainer, exists := i.getOrAppendFileMountInitContainer(pod)
	appendSecretToFileMountInitContainer(initContainer, secret.GetKey(), secretValue)

	if exists {
		return
	}

	volume := corev1.Volume{
		Name: EmbeddedSecretsFileMountVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium: corev1.StorageMediumMemory,
			},
		},
	}
	pod.Spec.Volumes = appendVolumeIfNotExists(pod.Spec.Volumes, volume)

	volumeMount := corev1.VolumeMount{
		Name:      EmbeddedSecretsFileMountVolumeName,
		ReadOnly:  true,
		MountPath: EmbeddedSecretsFileMountPath,
	}
	pod.Spec.InitContainers = AppendVolumeMounts(pod.Spec.InitContainers, volumeMount)
	pod.Spec.Containers = AppendVolumeMounts(pod.Spec.Containers, volumeMount)

	envVars := []corev1.EnvVar{
		// Set environment variable to let the containers know where to find the mounted files.
		{
			Name:  SecretPathDefaultDirEnvVar,
			Value: EmbeddedSecretsFileMountPath,
		},
		// Sets an empty prefix to let the containers know the file names will match the secret keys as-is.
		{
			Name:  SecretPathFilePrefixEnvVar,
			Value: "",
		},
	}
	if secret.GetEnvVar() != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name:  secret.GetEnvVar(),
			Value: filepath.Join(EmbeddedSecretsFileMountPath, strings.ToLower(secret.GetKey())),
		})
	}
	pod.Spec.InitContainers = AppendEnvVars(pod.Spec.InitContainers, envVars...)
	pod.Spec.Containers = AppendEnvVars(pod.Spec.Containers, envVars...)
}

func (i *EmbeddedSecretManagerInjector) getOrAppendFileMountInitContainer(pod *corev1.Pod) (*corev1.Container, bool /*exists*/) {
	index := slices.IndexFunc(
		pod.Spec.InitContainers,
		func(c corev1.Container) bool {
			return c.Name == i.cfg.FileMountInitContainer.ContainerName
		})
	if index != -1 {
		return &pod.Spec.InitContainers[index], true
	}

	pod.Spec.InitContainers = append(pod.Spec.InitContainers, corev1.Container{
		Name:  i.cfg.FileMountInitContainer.ContainerName,
		Image: i.cfg.FileMountInitContainer.Image,
		Command: []string{
			"sh",
			"-c",
			// Script below expects an environment variable EmbeddedSecretsFileMountInitContainerEnvVariableName
			// with contents in the following format:
			// 		secret_name1=base64_encoded_secret_value1
			// 		secret_name2=base64_encoded_secret_value2
			//		...
			//
			// It base64-decodes each secret value and writes it to a separate
			// file in the EmbeddedSecretsFileMountPath directory.
			fmt.Sprintf(`
				printf "%%s" "$%s" \
				| awk '/^.+=/ {
					i = index($0, "=");
					name = substr($0, 0, i - 1);
					value = substr($0, i + 1);
					output_file = "%s/" name;
					print value | "base64 -d > " output_file;
				}'
				`,
				EmbeddedSecretsFileMountInitContainerEnvVariableName,
				EmbeddedSecretsFileMountPath),
		},
		Resources: i.cfg.FileMountInitContainer.Resources,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      EmbeddedSecretsFileMountVolumeName,
				ReadOnly:  false,
				MountPath: EmbeddedSecretsFileMountPath,
			},
		},
	})

	return &pod.Spec.InitContainers[len(pod.Spec.InitContainers)-1], false
}

func appendSecretToFileMountInitContainer(initContainer *corev1.Container, secretKey string, secretValue []byte) {
	var envVar *corev1.EnvVar
	index := slices.IndexFunc(
		initContainer.Env,
		func(env corev1.EnvVar) bool { return env.Name == EmbeddedSecretsFileMountInitContainerEnvVariableName })
	if index != -1 {
		envVar = &initContainer.Env[index]
	} else {
		initContainer.Env = append(initContainer.Env, corev1.EnvVar{
			Name: EmbeddedSecretsFileMountInitContainerEnvVariableName,
		})
		envVar = &initContainer.Env[len(initContainer.Env)-1]
	}

	envVar.Value += fmt.Sprintf("%s=%s\n", secretKey, base64.StdEncoding.EncodeToString(secretValue))
}

func NewEmbeddedSecretManagerInjector(
	cfg config.EmbeddedSecretManagerConfig,
	secretFetchers []SecretFetcher,
	k8sClient client.Client,
	referenceNamespace string,
	secretCache goCache.CacheInterface[SecretValue],
	parentCfg *config.Config,
) SecretsInjector {
	return &EmbeddedSecretManagerInjector{
		cfg:                cfg,
		secretFetchers:     secretFetchers,
		k8sClient:          k8sClient,
		referenceNamespace: referenceNamespace,
		secretCache:        secretCache,
		parentCfg:          parentCfg,
	}
}

//go:generate mockery -name=MockableControllerRuntimeClient -output=./mocks -case=underscore

type MockableControllerRuntimeClient interface {
	client.Client
}
