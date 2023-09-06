package viper

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/flyteorg/flytestdlib/errors"

	"github.com/flyteorg/flytestdlib/logger"

	viperLib "github.com/spf13/viper"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
)

type Viper interface {
	BindPFlags(flags *pflag.FlagSet) error
	BindEnv(input ...string) error
	AutomaticEnv()
	ReadInConfig() error
	OnConfigChange(run func(in fsnotify.Event))
	WatchConfig()
	AllSettings() map[string]interface{}
	ConfigFileUsed() string
	MergeConfig(in io.Reader) error
}

// A proxy object for a collection of Viper instances.
type CollectionProxy struct {
	underlying   []Viper
	pflags       *pflag.FlagSet
	envVars      [][]string
	automaticEnv bool
}

func (c *CollectionProxy) BindPFlags(flags *pflag.FlagSet) error {
	err := errors.ErrorCollection{}
	for _, v := range c.underlying {
		err.Append(v.BindPFlags(flags))
	}

	c.pflags = flags

	return err.ErrorOrDefault()
}

func (c *CollectionProxy) BindEnv(input ...string) error {
	err := errors.ErrorCollection{}
	for _, v := range c.underlying {
		err.Append(v.BindEnv(input...))
	}

	if c.envVars == nil {
		c.envVars = make([][]string, 0, 1)
	}

	c.envVars = append(c.envVars, input)

	return err.ErrorOrDefault()
}

func (c *CollectionProxy) AutomaticEnv() {
	for _, v := range c.underlying {
		v.AutomaticEnv()
	}

	c.automaticEnv = true
}

func (c CollectionProxy) ReadInConfig() error {
	err := errors.ErrorCollection{}
	for _, v := range c.underlying {
		err.Append(v.ReadInConfig())
	}

	return err.ErrorOrDefault()
}

func (c CollectionProxy) OnConfigChange(run func(in fsnotify.Event)) {
	for _, v := range c.underlying {
		v.OnConfigChange(run)
	}
}

func (c CollectionProxy) WatchConfig() {
	for _, v := range c.underlying {
		v.WatchConfig()
	}
}

func (c CollectionProxy) AllSettings() map[string]interface{} {
	finalRes := map[string]interface{}{}
	if len(c.underlying) == 0 {
		return finalRes
	}

	combinedConfig, err := c.MergeAllConfigs()
	if err != nil {
		logger.Warnf(context.TODO(), "Failed to merge config. Error: %v", err)
		return finalRes
	}

	return combinedConfig.AllSettings()
}

func (c CollectionProxy) ConfigFileUsed() string {
	return fmt.Sprintf("[%v]", strings.Join(c.ConfigFilesUsed(), ","))
}

func (c CollectionProxy) MergeConfig(in io.Reader) error {
	panic("Not yet implemented.")
}

func (c CollectionProxy) MergeAllConfigs() (all Viper, err error) {
	combinedConfig := viperLib.New()
	if c.envVars != nil {
		for _, envConfig := range c.envVars {
			err = combinedConfig.BindEnv(envConfig...)
			if err != nil {
				return nil, err
			}
		}
	}

	if c.automaticEnv {
		combinedConfig.AutomaticEnv()
	}

	if c.pflags != nil {
		err = combinedConfig.BindPFlags(c.pflags)
		if err != nil {
			return nil, err
		}
	}

	for _, v := range c.underlying {
		if _, isCollection := v.(*CollectionProxy); isCollection {
			return nil, fmt.Errorf("merging nested CollectionProxies is not yet supported")
		}

		if len(v.ConfigFileUsed()) == 0 {
			continue
		}

		combinedConfig.SetConfigFile(v.ConfigFileUsed())

		reader, err := os.Open(v.ConfigFileUsed())
		if err != nil {
			return nil, err
		}

		err = combinedConfig.MergeConfig(reader)
		if err != nil {
			return nil, err
		}
	}

	return combinedConfig, nil
}

func (c CollectionProxy) ConfigFilesUsed() []string {
	res := make([]string, 0, len(c.underlying))
	for _, v := range c.underlying {
		filePath := v.ConfigFileUsed()
		if len(filePath) > 0 {
			res = append(res, filePath)
		}
	}

	return res
}
