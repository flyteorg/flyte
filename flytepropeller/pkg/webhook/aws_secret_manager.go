package webhook

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/ratelimit"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	corev1 "k8s.io/api/core/v1"
	"strings"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/webhook/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

const (
	// AWSSecretNameEnvVar defines the environment variable name to use when pulling the secret from AWS secrets manager.
	AWSSecretNameEnvVar     = "SECRET_NAME"
	UnionSecretEnvVarPrefix = "_UNION_"
)

// AWSSecretManagerInjector allows injecting of secrets from AWS Secret Manager as environment variable. It uses AWS-provided SideCar
// as an init-container to download the secret and save it to a local volume shared with all other containers in the pod.
// It supports multiple secrets to be mounted but that will result into adding an init container for each secret.
// The role/serviceaccount used to run the Pod must have permissions to pull the secret from AWS Secret Manager.
// Otherwise, the Pod will fail with an init-error.
// Files will be mounted on /etc/flyte/secrets/<SecretGroup>/<SecretKey>
type AWSSecretManagerInjector struct {
	cfg       config.EmbeddedSecretManagerConfig
	awsClient *secretsmanager.Client
}

func formatAWSSecretName(group, key string) string {
	return fmt.Sprintf("%v_%v", group, key)
}

func (i AWSSecretManagerInjector) Type() config.SecretManagerType {
	return config.SecretManagerTypeEmbedded
}

func lookUpSecret(ctx context.Context, secret *core.Secret, awsClient *secretsmanager.Client) (*string, error) {
	// Fetch the AWS secret from secrets manager
	secretName := formatAWSSecretName(secret.Group, secret.Key)
	logger.Infof(ctx, "Got fetch secret Request %v!", secretName)
	resp, err := awsClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"),
	})
	if err != nil {
		err := fmt.Errorf("failed to get secret %v due to %v", secretName, err)
		logger.Error(ctx, err)
		return nil, err
	}
	logger.Infof(ctx, "Response %v ", resp)
	return resp.SecretString, nil
}
func (i AWSSecretManagerInjector) Inject(ctx context.Context, secret *core.Secret, p *corev1.Pod) (newP *corev1.Pod, injected bool, err error) {
	if len(secret.Key) == 0 {
		return nil, false, fmt.Errorf("AWS Secrets Webhook require both key to be set. "+
			"Secret: [%v]", secret)
	}

	switch secret.MountRequirement {
	case core.Secret_ANY:
		fallthrough
	case core.Secret_ENV_VAR:
		fallthrough
	case core.Secret_FILE:
		// Fetch the AWS secret from secrets manager
		// Only use the group for lookup
		secret.Group = "union-global"
		secretName, err := lookUpSecret(ctx, secret, i.awsClient)
		for _, c := range p.Spec.Containers {
			logger.Infof(ctx, "env variables on the pod", c.Env)
		}
		if err != nil {
			return p, false, err
		}

		prefixEnvVar := corev1.EnvVar{
			Name:  SecretEnvVarPrefix,
			Value: UnionSecretEnvVarPrefix,
		}
		// Inject AWS secret-inject webhook annotations to mount the secret in a predictable location.
		envVars := []corev1.EnvVar{
			prefixEnvVar,
			// Set environment variable to let the container know where to find the mounted files.
			{
				Name:  UnionSecretEnvVarPrefix + strings.ToUpper(secret.Key),
				Value: *secretName,
			},
		}

		for _, envVar := range envVars {
			p.Spec.InitContainers = AppendEnvVars(p.Spec.InitContainers, envVar)
			p.Spec.Containers = AppendEnvVars(p.Spec.Containers, envVar)
		}

	default:
		err := fmt.Errorf("unrecognized mount requirement [%v] for secret [%v]", secret.MountRequirement.String(), secret.Key)
		logger.Error(ctx, err)
		return p, false, err
	}

	return p, true, nil
}

// NewAWSSecretManagerInjector creates a SecretInjector that's able to mount secrets from AWS Secret Manager.
func NewAWSSecretManagerInjector(ctx context.Context, cfg config.EmbeddedSecretManagerConfig) (AWSSecretManagerInjector, error) {
	if cfg.Type != "AWS" {
		return AWSSecretManagerInjector{}, fmt.Errorf("failed to start secret manager service due to unsupported type %v. Only supported for aws right now", cfg.Type)
	}
	awsCfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(cfg.AWSConfig.Region),
		awsConfig.WithRetryer(func() aws.Retryer {
			return retry.NewStandard(func(options *retry.StandardOptions) {
				options.MaxAttempts = cfg.AWSConfig.MaxRetries
			})
		}))
	if err != nil {
		logger.Errorf(ctx, "failed to start secret manager service due to %v", err)
		return AWSSecretManagerInjector{}, fmt.Errorf("failed to start secret manager service due to %v", err)
	}
	awsCfg.RetryMode = aws.RetryModeStandard
	awsCfg.RetryMaxAttempts = cfg.AWSConfig.MaxRetries
	awsCfg.Retryer = func() aws.Retryer {
		return retry.NewStandard(func(o *retry.StandardOptions) {
			o.MaxAttempts = cfg.AWSConfig.MaxRetries
			o.RateLimiter = ratelimit.NewTokenRateLimit(cfg.AWSConfig.RateLimiterTokens)
		})
	}
	return AWSSecretManagerInjector{
		cfg:       cfg,
		awsClient: secretsmanager.NewFromConfig(awsCfg),
	}, nil
}
