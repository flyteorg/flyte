package secretmanager

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	coreIdl "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytestdlib/logger"
)

// Env Var Lookup based on Prefix + SecretGroup + _ + SecretKey
const envVarLookupFormatter = "%s%s_%s"

// FileEnvSecretManager allows retrieving secrets mounted to this process through Env Vars or Files.
type FileEnvSecretManager struct {
	secretPath string
	envPrefix  string
}

func (f FileEnvSecretManager) Get(ctx context.Context, key string) (string, error) {
	envVar := fmt.Sprintf("%s%s", f.envPrefix, key)
	v, ok := os.LookupEnv(envVar)
	if ok {
		logger.Debugf(ctx, "Secret found %s", v)
		return v, nil
	}

	secretFile := filepath.Join(f.secretPath, key)
	if _, err := os.Stat(secretFile); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("secrets not found - file [%s], Env [%s]", secretFile, envVar)
		}
		return "", err
	}

	logger.Debugf(ctx, "reading secrets from filePath [%s]", secretFile)
	b, err := ioutil.ReadFile(secretFile)
	if err != nil {
		return "", err
	}
	return string(b), err
}

// GetForSecret retrieves a secret from the environment of the running process. To lookup secret, both secret's key and
// group must be non-empty. GetForSecret will first lookup env variables using  the configured
// Prefix+SecretGroup+_+SecretKey. If the secret is not found in environment, it'll lookup the secret from files using
// the configured SecretPath / SecretGroup / SecretKey.
func (f FileEnvSecretManager) GetForSecret(ctx context.Context, secret *coreIdl.Secret) (string, error) {
	if len(secret.Group) == 0 || len(secret.Key) == 0 {
		return "", fmt.Errorf("both key and group are required parameters. Secret: [%v]", secret.String())
	}

	envVar := fmt.Sprintf(envVarLookupFormatter, f.envPrefix, strings.ToUpper(secret.Group), strings.ToUpper(secret.Key))
	v, ok := os.LookupEnv(envVar)
	if ok {
		logger.Debugf(ctx, "Secret found %s", v)
		return v, nil
	}

	secretFile := filepath.Join(f.secretPath, filepath.Join(secret.Group, secret.Key))
	if _, err := os.Stat(secretFile); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("secrets not found - Env [%s], file [%s]", envVar, secretFile)
		}

		return "", err
	}

	logger.Debugf(ctx, "reading secrets from filePath [%s]", secretFile)
	b, err := ioutil.ReadFile(secretFile)
	if err != nil {
		return "", err
	}

	return string(b), err
}

func NewFileEnvSecretManager(cfg *Config) FileEnvSecretManager {
	return FileEnvSecretManager{
		secretPath: cfg.SecretFilePrefix,
		envPrefix:  cfg.EnvironmentPrefix,
	}
}
