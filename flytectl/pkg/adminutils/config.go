package adminutils

import "github.com/flyteorg/flytestdlib/config"

//go:generate pflags Config --default-var DefaultConfig --bind-default-var

var (
	DefaultConfig = &Config{
		MaxRecords: 500,
		BatchSize:  100,
	}
	section = config.MustRegisterSection("adminutils", DefaultConfig)
)

type Config struct {
	MaxRecords int `json:"maxRecords" pflag:",Maximum number of records to retrieve."`
	BatchSize  int `json:"batchSize" pflag:",Maximum number of records to retrieve per call."`
}

func GetConfig() *Config {
	return section.GetConfig().(*Config)
}
