package config

import "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"

// Database config. Contains values necessary to open a database connection.
type DbConfig struct {
	BaseConfig
	Host         string `json:"host"`
	Port         int    `json:"port"`
	DbName       string `json:"dbname"`
	User         string `json:"user"`
	Password     string `json:"password"`
	ExtraOptions string `json:"options"`
}

func NewDbConfig(dbConfigValues interfaces.DbConfig) DbConfig {
	return DbConfig{
		BaseConfig: BaseConfig{
			IsDebug: dbConfigValues.Debug,
		},
		Host:         dbConfigValues.Host,
		Port:         dbConfigValues.Port,
		DbName:       dbConfigValues.DbName,
		User:         dbConfigValues.User,
		Password:     dbConfigValues.Password,
		ExtraOptions: dbConfigValues.ExtraOptions,
	}
}
