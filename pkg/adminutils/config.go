package adminutils

import "github.com/lyft/flytestdlib/config"

//go:generate pflags Config

var (
	defaultConfig = &Config{
		MaxRecords: 500,
		BatchSize:  100,
	}
	section = config.MustRegisterSection("adminutils", defaultConfig)
)

type Config struct {
	MaxRecords int `json:"maxRecords" pflag:",Maximum number of records to retrieve."`
	BatchSize  int `json:"batchSize" pflag:",Maximum number of records to retrieve per call."`
}

func GetConfig() *Config {
	return section.GetConfig().(*Config)
}
