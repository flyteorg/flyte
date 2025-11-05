package ioutils

import (
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/config"
)

//go:generate pflags Config --default-var=defaultConfig

var cfgSection = config.MustRegisterSubSection("ioutils", defaultConfig)

type Config struct {
	RemoteFileOutputPathsConfig RemoteFileOutputPathsConfig `json:"remoteFileOutputPaths" pflag:",Config for remote file output paths."`
}

type RemoteFileOutputPathsConfig struct {
	DeckFilename string `json:"deckFilename" pflag:",Filename to use for the deck file."`
}

var defaultConfig = &Config{
	RemoteFileOutputPathsConfig: RemoteFileOutputPathsConfig{
		DeckFilename: "deck.html",
	},
}

func GetConfig() *Config {
	return cfgSection.GetConfig().(*Config)
}
