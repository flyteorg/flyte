package secret

import (
	"context"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	corev1 "k8s.io/api/core/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	stdlibErrors "github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

const (
	UnionSecretEnvVarPrefix = "_UNION_"
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
	cfg           config.EmbeddedSecretManagerConfig
	secretFetcher SecretFetcher
}

func (i EmbeddedSecretManagerInjector) Type() config.SecretManagerType {
	return config.SecretManagerTypeEmbedded
}

func GetSecretID(secretKey string, source admin.AttributesSource, labels map[string]string) (string, error) {
	err := validateRequiredFieldsExist(labels)
	if err != nil {
		return "", err
	}
	if source == admin.AttributesSource_PROJECT_DOMAIN {
		return EncodeSecretName(labels[OrganizationLabel], labels[DomainLabel], labels[ProjectLabel], secretKey), nil
	}

	if source == admin.AttributesSource_DOMAIN {
		return EncodeSecretName(labels[OrganizationLabel], labels[DomainLabel], EmptySecretScope, secretKey), nil
	}

	return EncodeSecretName(labels[OrganizationLabel], EmptySecretScope, EmptySecretScope, secretKey), nil
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

func (i EmbeddedSecretManagerInjector) lookUpSecret(ctx context.Context, secret *core.Secret, labels map[string]string) (*SecretValue, error) {
	// Fetch the secret from configured secrets manager
	err := validateRequiredFieldsExist(labels)
	if err != nil {
		return nil, err
	}
	// Fetch project-domain scoped secret
	projectDomainScopedSecret := EncodeSecretName(labels[OrganizationLabel], labels[DomainLabel], labels[ProjectLabel], secret.Key)
	secretValue, err := i.secretFetcher.GetSecretValue(ctx, projectDomainScopedSecret)
	if err == nil {
		return secretValue, nil
	}
	if !stdlibErrors.IsCausedBy(err, ErrCodeSecretNotFound) {
		return nil, err
	}

	// Fetch domain scoped secret
	domainScopedSecret := EncodeSecretName(labels[OrganizationLabel], labels[DomainLabel], EmptySecretScope, secret.Key)
	secretValue, err = i.secretFetcher.GetSecretValue(ctx, domainScopedSecret)
	if err == nil {
		return secretValue, nil
	}
	if !stdlibErrors.IsCausedBy(err, ErrCodeSecretNotFound) {
		return nil, err
	}

	// Fetch organization scoped secret
	orgScopedSecret := EncodeSecretName(labels[OrganizationLabel], EmptySecretScope, EmptySecretScope, secret.Key)
	secretValue, err = i.secretFetcher.GetSecretValue(ctx, orgScopedSecret)
	if err == nil {
		return secretValue, err
	}
	if !stdlibErrors.IsCausedBy(err, ErrCodeSecretNotFound) {
		return nil, err
	}

	return nil, stdlibErrors.Errorf(ErrCodeSecretNotFoundAcrossAllScopes, SecretSecretNotFoundAcrossAllScopes)
}

func (i EmbeddedSecretManagerInjector) Inject(
	ctx context.Context,
	secret *core.Secret,
	pod *corev1.Pod,
) (*corev1.Pod, bool /*injected*/, error) {
	if len(secret.Key) == 0 {
		return pod, false, fmt.Errorf("EmbeddedSecretManager requires key to be set. Secret: [%v]", secret)
	}

	secretValue, err := i.lookUpSecret(ctx, secret, pod.Labels)
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
		if secretValue.BinaryValue == nil {
			return pod, false, fmt.Errorf(
				"secret %q is attempted to be mounted as a file, but has no binary "+
					"value; mount as an environment variable instead", secret.Key)
		}
		i.injectAsFile(secret, secretValue.BinaryValue, pod)
	default:
		err := fmt.Errorf("unrecognized mount requirement [%v] for secret [%v]", secret.MountRequirement.String(), secret.Key)
		logger.Error(ctx, err)
		return pod, false, err
	}

	return pod, true, nil
}

func (i EmbeddedSecretManagerInjector) injectAsEnvVar(secret *core.Secret, secretValue string, pod *corev1.Pod) {
	envVars := []corev1.EnvVar{
		{
			Name:  SecretEnvVarPrefix,
			Value: UnionSecretEnvVarPrefix,
		},
		{
			Name:  UnionSecretEnvVarPrefix + strings.ToUpper(secret.GetKey()),
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

func (i EmbeddedSecretManagerInjector) injectAsFile(secret *core.Secret, secretValue []byte, pod *corev1.Pod) {
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

func (i EmbeddedSecretManagerInjector) getOrAppendFileMountInitContainer(pod *corev1.Pod) (*corev1.Container, bool /*exists*/) {
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

func NewEmbeddedSecretManagerInjector(cfg config.EmbeddedSecretManagerConfig, secretFetcher SecretFetcher) SecretsInjector {
	return EmbeddedSecretManagerInjector{
		cfg:           cfg,
		secretFetcher: secretFetcher,
	}
}
