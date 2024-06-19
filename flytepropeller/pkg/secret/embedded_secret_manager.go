package secret

import (
	"context"
	"fmt"
	"strings"

	gcpsm "cloud.google.com/go/secretmanager/apiv1"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	awssm "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	corev1 "k8s.io/api/core/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	stdlibErrors "github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

const (
	UnionSecretEnvVarPrefix                                     = "_UNION_"
	SecretFieldSeparator                                        = "__"
	ValueFormatter                                              = "%s"
	SecretsStorageUnionPrefix                                   = "u"
	SecretsStorageOrgPrefixFormat                               = SecretsStorageUnionPrefix + SecretFieldSeparator + "org" + SecretFieldSeparator + ValueFormatter
	SecretsStorageDomainPrefixFormat                            = SecretsStorageOrgPrefixFormat + SecretFieldSeparator + "domain" + SecretFieldSeparator + ValueFormatter
	SecretsStorageProjectPrefixFormat                           = SecretsStorageDomainPrefixFormat + SecretFieldSeparator + "project" + SecretFieldSeparator + ValueFormatter
	SecretsStorageFormat                                        = SecretsStorageProjectPrefixFormat + SecretFieldSeparator + "key" + SecretFieldSeparator + ValueFormatter
	ProjectLabel                                                = "project"
	DomainLabel                                                 = "domain"
	OrganizationLabel                                           = "organization"
	EmptySecretScope                                            = ""
	AWSSecretLatesVersion                                       = "AWSCURRENT"
	GCPSecretNameFormat                                         = "projects/%s/secrets/%s/versions/latest"                                   // #nosec G101
	SecretNotFoundErrorFormat                                   = "secret %v not found in the secret manager"                                // #nosec G101
	SecretReadFailureErrorFormat                                = "secret %v failed to be read from secret manager"                          // #nosec G101
	SecretNilErrorFormat                                        = "secret %v read as empty from the secret manager"                          // #nosec G101
	SecretRequirementsErrorFormat                               = "secret read requirements not met due to empty %v field in the pod labels" // #nosec G101
	SecretSecretNotFoundAcrossAllScopes                         = "secret not found across all scope"                                        // #nosec G101
	ErrCodeSecretRequirementsError       stdlibErrors.ErrorCode = "SecretRequirementsError"                                                  // #nosec G101
	ErrCodeSecretNotFound                stdlibErrors.ErrorCode = "SecretNotFound"                                                           // #nosec G101
	ErrCodeSecretNotFoundAcrossAllScopes stdlibErrors.ErrorCode = "SecretNotFoundAcrossAllScopes"                                            // #nosec G101
	ErrCodeSecretReadFailure             stdlibErrors.ErrorCode = "SecretReadFailure"                                                        // #nosec G101
	ErrCodeSecretNil                     stdlibErrors.ErrorCode = "SecretNil"                                                                // #nosec G101
)

//go:generate mockery --output=./mocks --case=underscore -name=SecretFetcher
type SecretFetcher interface {
	Get(ctx context.Context, key string) (string, error)
}

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

func (i EmbeddedSecretManagerInjector) lookUpSecret(ctx context.Context, secret *core.Secret, labels map[string]string) (string, error) {
	// Fetch the secret from configured secrets manager
	err := validateRequiredFieldsExist(labels)
	if err != nil {
		return "", err
	}
	// Fetch project-domain scoped secret
	projectDomainScopedSecret := fmt.Sprintf(SecretsStorageFormat, labels[OrganizationLabel], labels[DomainLabel], labels[ProjectLabel], secret.Key)
	secretValue, err := i.secretFetcher.Get(ctx, projectDomainScopedSecret)
	if err != nil && !stdlibErrors.IsCausedBy(err, ErrCodeSecretNotFound) {
		return "", err
	}
	if len(secretValue) > 0 {
		return secretValue, nil
	}

	// Fetch domain scoped secret
	domainScopedSecret := fmt.Sprintf(SecretsStorageFormat, labels[OrganizationLabel], labels[DomainLabel], EmptySecretScope, secret.Key)
	secretValue, err = i.secretFetcher.Get(ctx, domainScopedSecret)
	if err != nil && !stdlibErrors.IsCausedBy(err, ErrCodeSecretNotFound) {
		return "", err
	}
	if len(secretValue) > 0 {
		return secretValue, nil
	}

	// Fetch organization scoped secret
	orgScopedSecret := fmt.Sprintf(SecretsStorageFormat, labels[OrganizationLabel], EmptySecretScope, EmptySecretScope, secret.Key)
	secretValue, err = i.secretFetcher.Get(ctx, orgScopedSecret)
	if err != nil && !stdlibErrors.IsCausedBy(err, ErrCodeSecretNotFound) {
		return "", err
	}
	if len(secretValue) > 0 {
		return secretValue, nil
	}

	return "", stdlibErrors.Errorf(ErrCodeSecretNotFoundAcrossAllScopes, SecretSecretNotFoundAcrossAllScopes)
}
func (i EmbeddedSecretManagerInjector) Inject(ctx context.Context, secret *core.Secret, p *corev1.Pod) (newP *corev1.Pod, injected bool, err error) {
	if len(secret.Key) == 0 {
		return p, false, fmt.Errorf("EmbeddedSecretManager requires key to be set. "+
			"Secret: [%v]", secret)
	}

	switch secret.MountRequirement {
	case core.Secret_ANY:
		fallthrough
	case core.Secret_ENV_VAR:
		// Fetch the secret from secrets manager
		secretValue, err := i.lookUpSecret(ctx, secret, p.Labels)
		if err != nil {
			return p, false, err
		}

		prefixEnvVar := corev1.EnvVar{
			Name:  SecretEnvVarPrefix,
			Value: UnionSecretEnvVarPrefix,
		}
		// Inject secret-inject webhook annotations to mount the secret in a predictable location.
		envVars := []corev1.EnvVar{
			prefixEnvVar,
			// Set environment variable to let the container know where to find the mounted files.
			{
				Name:  UnionSecretEnvVarPrefix + strings.ToUpper(secret.Key),
				Value: secretValue,
			},
		}

		for _, envVar := range envVars {
			p.Spec.InitContainers = AppendEnvVars(p.Spec.InitContainers, envVar)
			p.Spec.Containers = AppendEnvVars(p.Spec.Containers, envVar)
		}

	case core.Secret_FILE:
		err := fmt.Errorf("secret [%v] requirement is not supported for secret [%v]", secret.MountRequirement.String(), secret.Key)
		logger.Error(ctx, err)
		return p, false, err
	default:
		err := fmt.Errorf("unrecognized mount requirement [%v] for secret [%v]", secret.MountRequirement.String(), secret.Key)
		logger.Error(ctx, err)
		return p, false, err
	}

	return p, true, nil
}

func NewEmbeddedSecretManagerInjector(cfg config.EmbeddedSecretManagerConfig, secretFetcher SecretFetcher) SecretsInjector {
	return EmbeddedSecretManagerInjector{
		cfg:           cfg,
		secretFetcher: secretFetcher,
	}
}

func NewSecretFetcherManager(ctx context.Context, cfg config.EmbeddedSecretManagerConfig) (SecretFetcher, error) {
	switch cfg.Type {
	case config.EmbeddedSecretManagerTypeAWS:
		awsCfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(cfg.AWSConfig.Region))
		if err != nil {
			logger.Errorf(ctx, "failed to start secret manager service due to %v", err)
			return nil, fmt.Errorf("failed to start secret manager service due to %v", err)
		}
		return NewAWSSecretFetcher(cfg.AWSConfig, awssm.NewFromConfig(awsCfg)), nil
	case config.EmbeddedSecretManagerTypeGCP:
		gcpSmClient, err := gcpsm.NewClient(ctx)
		if err != nil {
			logger.Errorf(ctx, "failed to start secret manager service due to %v", err)
			return nil, fmt.Errorf("failed to start secret manager service due to %v", err)
		}
		return NewGCPSecretFetcher(cfg.GCPConfig, gcpSmClient), nil
	}
	return nil, fmt.Errorf("failed to start secret fetcher service due to unsupported type %v. Only supported for aws and gcp right now", cfg.Type)
}
