package qubole

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flyteplugins/go/tasks/v1/qubole/config"
)

type secretsManager struct {
	// Memoize the key
	quboleKey string
}

type SecretsManager interface {
	GetToken() (string, error)
}

func NewSecretsManager() SecretsManager {
	return &secretsManager{}
}

func (s *secretsManager) GetToken() (string, error) {
	if s.quboleKey != "" {
		return s.quboleKey, nil
	}

	// If the environment variable has been defined, then just use the value of the environment
	// variable.  This is primarily for local use/testing purposes.  Otherwise we expect the token
	// to exist in a file.
	if key := os.Getenv("QUBOLE_API_KEY"); key != "" {
		return key, nil
	}

	// Assume that secrets have been mounted somehow
	fileLocation := config.GetQuboleConfig().QuboleTokenPath

	b, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		logger.Errorf(context.Background(), "Could not read entry at %s", fileLocation)
		return "", errors.New(fmt.Sprintf("Bad Qubole token file, could not read file at [%s]", fileLocation))
	}
	s.quboleKey = strings.TrimSpace(string(b))
	if s.quboleKey == "" {
		logger.Errorf(context.Background(), "Qubole token was empty")
		return "", errors.New("bad Qubole token - file was read but was empty")
	}

	return s.quboleKey, nil
}
