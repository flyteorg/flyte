package configuration

import (
	"github.com/flyteorg/flyte/flytestdlib/config"
)

const artifactsServer = "artifactsServer"

type PostgresConfig struct {
	Host   string `json:"host" pflag:",The host name of the database server"`
	Port   int    `json:"port" pflag:",The port name of the database server"`
	DbName string `json:"dbname" pflag:",The database name"`
	User   string `json:"username" pflag:",The database user who is connecting to the server."`
	// Either Password or PasswordPath must be set.
	Password     string `json:"password" pflag:",The database password."`
	PasswordPath string `json:"passwordPath" pflag:",Points to the file containing the database password."`
	ExtraOptions string `json:"options" pflag:",See http://gorm.io/docs/connecting_to_the_database.html for available options passed, in addition to the above."`
	Debug        bool   `json:"debug" pflag:" Whether or not to start the database connection with debug mode enabled."`
}

type DbConfig struct {
	Postgres PostgresConfig `json:"postgres" pflag:",Postgres database configuration"`
}

type ApplicationConfiguration struct {
	Database DbConfig `json:"database" pflag:",Database configuration"`
}

var defaultApplicationConfiguration = ApplicationConfiguration{
	Database: DbConfig{
		Postgres: PostgresConfig{
			// These values are suitable for local sandbox development
			Host:     "localhost",
			Port:     30001,
			DbName:   "artifacts",
			User:     "postgres",
			Password: "postgres",
		},
	},
}

var ApplicationConfig = config.MustRegisterSection(artifactsServer, &defaultApplicationConfiguration)
