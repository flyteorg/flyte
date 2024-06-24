package secret

import (
	"context"
	"encoding/base64"
	"fmt"
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
	EmbeddedSecretsVolumeName = "embedded-secret-vol" // #nosec G101
	EmbeddedSecretsMountPath  = "/etc/flyte/secrets"  // #nosec G101

	SecretFieldSeparator              = "__"
	ValueFormatter                    = "%s"
	SecretsStorageUnionPrefix         = "u"
	SecretsStorageOrgPrefixFormat     = SecretsStorageUnionPrefix + SecretFieldSeparator + "org" + SecretFieldSeparator + ValueFormatter
	SecretsStorageDomainPrefixFormat  = SecretsStorageOrgPrefixFormat + SecretFieldSeparator + "domain" + SecretFieldSeparator + ValueFormatter
	SecretsStorageProjectPrefixFormat = SecretsStorageDomainPrefixFormat + SecretFieldSeparator + "project" + SecretFieldSeparator + ValueFormatter
	SecretsStorageFormat              = SecretsStorageProjectPrefixFormat + SecretFieldSeparator + "key" + SecretFieldSeparator + ValueFormatter
	ProjectLabel                      = "project"
	DomainLabel                       = "domain"
	OrganizationLabel                 = "organization"
	EmptySecretScope                  = ""

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
		return fmt.Sprintf(SecretsStorageFormat, labels[OrganizationLabel], labels[DomainLabel], labels[ProjectLabel], secretKey), nil
	}

	if source == admin.AttributesSource_DOMAIN {
		return fmt.Sprintf(SecretsStorageFormat, labels[OrganizationLabel], labels[DomainLabel], EmptySecretScope, secretKey), nil
	}

	orgScopedSecret := fmt.Sprintf(SecretsStorageFormat, labels[OrganizationLabel], EmptySecretScope, EmptySecretScope, secretKey)
	return orgScopedSecret, nil
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
	projectDomainScopedSecret := fmt.Sprintf(SecretsStorageFormat, labels[OrganizationLabel], labels[DomainLabel], labels[ProjectLabel], secret.Key)
	secretValue, err := i.secretFetcher.GetSecretValue(ctx, projectDomainScopedSecret)
	if err == nil {
		return secretValue, nil
	}
	if !stdlibErrors.IsCausedBy(err, ErrCodeSecretNotFound) {
		return nil, err
	}

	// Fetch domain scoped secret
	domainScopedSecret := fmt.Sprintf(SecretsStorageFormat, labels[OrganizationLabel], labels[DomainLabel], EmptySecretScope, secret.Key)
	secretValue, err = i.secretFetcher.GetSecretValue(ctx, domainScopedSecret)
	if err == nil {
		return secretValue, nil
	}
	if !stdlibErrors.IsCausedBy(err, ErrCodeSecretNotFound) {
		return nil, err
	}

	// Fetch organization scoped secret
	orgScopedSecret := fmt.Sprintf(SecretsStorageFormat, labels[OrganizationLabel], EmptySecretScope, EmptySecretScope, secret.Key)
	secretValue, err = i.secretFetcher.GetSecretValue(ctx, orgScopedSecret)
	if err != nil {
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
		i.injectAsEnvVar(secret.Key, stringValue, pod)
	case core.Secret_FILE:
		if secretValue.BinaryValue == nil {
			return pod, false, fmt.Errorf(
				"secret %q is attempted to be mounted as a file, but has no binary "+
					"value; mount as an environment variable instead", secret.Key)
		}
		i.injectAsFile(secret.Key, secretValue.BinaryValue, pod)
	default:
		err := fmt.Errorf("unrecognized mount requirement [%v] for secret [%v]", secret.MountRequirement.String(), secret.Key)
		logger.Error(ctx, err)
		return pod, false, err
	}

	return pod, true, nil
}

func (i EmbeddedSecretManagerInjector) injectAsEnvVar(secretKey string, secretValue string, pod *corev1.Pod) {
	envVars := []corev1.EnvVar{
		{
			Name:  SecretEnvVarPrefix,
			Value: UnionSecretEnvVarPrefix,
		},
		{
			Name:  UnionSecretEnvVarPrefix + strings.ToUpper(secretKey),
			Value: secretValue,
		},
	}
	pod.Spec.InitContainers = AppendEnvVars(pod.Spec.InitContainers, envVars...)
	pod.Spec.Containers = AppendEnvVars(pod.Spec.Containers, envVars...)
}

func (i EmbeddedSecretManagerInjector) injectAsFile(secretKey string, secretValue []byte, pod *corev1.Pod) {
	// A volume with a static name so that if we try to inject multiple secrets, we won't mount multiple volumes.
	volume := corev1.Volume{
		Name: EmbeddedSecretsVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium: corev1.StorageMediumMemory,
			},
		},
	}
	pod.Spec.Volumes = appendVolumeIfNotExists(pod.Spec.Volumes, volume)

	secretFilePath := EmbeddedSecretsMountPath + "/" + secretKey
	secretInitContainer := corev1.Container{
		Name:  "init-embedded-secret-" + secretKey,
		Image: i.cfg.FileMountInitContainer.Image,
		Env: []corev1.EnvVar{
			{
				Name:  "SECRET_VALUE",
				Value: base64.StdEncoding.EncodeToString(secretValue),
			},
		},
		Command: []string{
			"sh",
			"-c",
			fmt.Sprintf("printf \"%%s\" \"$SECRET_VALUE\" | base64 -d > \"%s\"", secretFilePath),
		},
		Resources: i.cfg.FileMountInitContainer.Resources,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      EmbeddedSecretsVolumeName,
				ReadOnly:  false,
				MountPath: EmbeddedSecretsMountPath,
			},
		},
	}
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, secretInitContainer)

	secretVolumeMount := corev1.VolumeMount{
		Name:      EmbeddedSecretsVolumeName,
		ReadOnly:  true,
		MountPath: EmbeddedSecretsMountPath,
	}
	pod.Spec.InitContainers = AppendVolumeMounts(pod.Spec.InitContainers, secretVolumeMount)
	pod.Spec.Containers = AppendVolumeMounts(pod.Spec.Containers, secretVolumeMount)

	// Inject AWS secret-inject webhook annotations to mount the secret in a predictable location.
	envVars := []corev1.EnvVar{
		// Set environment variable to let the containers know where to find the mounted files.
		{
			Name:  SecretPathDefaultDirEnvVar,
			Value: EmbeddedSecretsMountPath,
		},
		// Sets an empty prefix to let the containers know the file names will match the secret keys as-is.
		{
			Name:  SecretPathFilePrefixEnvVar,
			Value: "",
		},
	}
	pod.Spec.InitContainers = AppendEnvVars(pod.Spec.InitContainers, envVars...)
	pod.Spec.Containers = AppendEnvVars(pod.Spec.Containers, envVars...)
}

func NewEmbeddedSecretManagerInjector(cfg config.EmbeddedSecretManagerConfig, secretFetcher SecretFetcher) SecretsInjector {
	return EmbeddedSecretManagerInjector{
		cfg:           cfg,
		secretFetcher: secretFetcher,
	}
}
