package database

import (
	"time"

	"github.com/flyteorg/flyte/flytestdlib/config"
)

const (
	database    = "database"
	postgresStr = "postgres"
)

//go:generate pflags DbConfig --default-var=defaultConfig

var defaultConfig = &DbConfig{
	MaxIdleConnections: 10,
	MaxOpenConnections: 100,
	ConnMaxLifeTime:    config.Duration{Duration: time.Hour},
	Postgres: PostgresConfig{
		// These values are suitable for local sandbox development
		Host:            "localhost",
		ReadReplicaHost: "localhost",
		Port:            30001,
		DbName:          postgresStr,
		User:            postgresStr,
		Password:        postgresStr,
		ExtraOptions:    "sslmode=disable",
	},
}
var configSection = config.MustRegisterSection(database, defaultConfig)

// DbConfig is used to for initiating the database connection with the store that holds registered
// entities (e.g. workflows, tasks, launch plans...)
type DbConfig struct {
	EnableForeignKeyConstraintWhenMigrating bool            `json:"enableForeignKeyConstraintWhenMigrating" pflag:",Whether to enable gorm foreign keys when migrating the db"`
	MaxIdleConnections                      int             `json:"maxIdleConnections" pflag:",maxIdleConnections sets the maximum number of connections in the idle connection pool."`
	MaxOpenConnections                      int             `json:"maxOpenConnections" pflag:",maxOpenConnections sets the maximum number of open connections to the database."`
	ConnMaxLifeTime                         config.Duration `json:"connMaxLifeTime" pflag:",sets the maximum amount of time a connection may be reused"`
	Postgres                                PostgresConfig  `json:"postgres,omitempty"`
	SQLite                                  SQLiteConfig    `json:"sqlite,omitempty"`
}

// SQLiteConfig can be used to configure
type SQLiteConfig struct {
	File string `json:"file" pflag:",The path to the file (existing or new) where the DB should be created / stored. If existing, then this will be reused, else a new will be created"`
}

// PostgresConfig includes specific config options for opening a connection to a postgres database.
type PostgresConfig struct {
	Host            string `json:"host" pflag:",The host name of the database server"`
	ReadReplicaHost string `json:"readReplicaHost" pflag:",The host name of the read replica database server"`
	Port            int    `json:"port" pflag:",The port name of the database server"`
	DbName          string `json:"dbname" pflag:",The database name"`
	User            string `json:"username" pflag:",The database user who is connecting to the server."`
	// Either Password or PasswordPath must be set.
	Password     string `json:"password" pflag:",The database password."`
	PasswordPath string `json:"passwordPath" pflag:",Points to the file containing the database password."`
	ExtraOptions string `json:"options" pflag:",See http://gorm.io/docs/connecting_to_the_database.html for available options passed, in addition to the above."`
	Debug        bool   `json:"debug" pflag:" Whether or not to start the database connection with debug mode enabled."`
}

var emptySQLiteCfg = SQLiteConfig{}
var emptyPostgresConfig = PostgresConfig{}

func (s SQLiteConfig) IsEmpty() bool {
	return s == emptySQLiteCfg
}

func (s PostgresConfig) IsEmpty() bool {
	return s == emptyPostgresConfig
}

func GetConfig() *DbConfig {
	return configSection.GetConfig().(*DbConfig)
}
