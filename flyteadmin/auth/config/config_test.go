package config

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/config/viper"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"

	"github.com/ory/fosite"
)

func TestHashFlyteClientSecret(t *testing.T) {
	hasher := &fosite.BCrypt{WorkFactor: 6}
	res, err := hasher.Hash(context.Background(), []byte("foobar"))
	assert.NoError(t, err)
	t.Log(string(res))
}

type Secret struct {
	Myval []byte `json:"myval"`
}

func TestFromYaml(t *testing.T) {
	raw, err := ioutil.ReadFile(filepath.Join("testdata", "secret.yaml"))
	assert.NoError(t, err)

	s := Secret{}
	assert.NoError(t, yaml.Unmarshal(raw, &s))
	assert.Equal(t, "$2a$06$pxs1AkG81Kvrhpml1QiLSOQaTk9eePrU/7Yab9y07h3x0TglbaoT6", string(s.Myval))
}

func TestParseClientSecretConfig(t *testing.T) {
	assert.NoError(t, logger.SetConfig(&logger.Config{IncludeSourceCode: true}))

	accessor := viper.NewAccessor(config.Options{
		RootSection: cfgSection,
		SearchPaths: []string{filepath.Join("testdata", "config.yaml")},
	})

	assert.NoError(t, accessor.UpdateConfig(context.Background()))
	assert.Equal(t, "my-client", GetConfig().AppAuth.SelfAuthServer.StaticClients["my-client"].ID)
}

func TestDefaultConfig(t *testing.T) {
	assert.Equal(t, len(DefaultConfig.AppAuth.SelfAuthServer.StaticClients), 3)
	assert.Equal(t, DefaultConfig.AppAuth.SelfAuthServer.StaticClients["flyte-cli"].ID, "flyte-cli")
}

func TestCompare(t *testing.T) {
	hasher := &fosite.BCrypt{WorkFactor: 6}
	err := hasher.Compare(context.Background(), DefaultConfig.AppAuth.SelfAuthServer.StaticClients["flytepropeller"].Secret, []byte("foobar"))
	assert.Error(t, err)
}
