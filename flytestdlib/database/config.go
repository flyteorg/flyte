package database

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
)

const (
	database    = "database"
	postgresStr = "postgres"

	AuthTypePassword = "password"
	AuthTypeIAM      = "iam"
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
		AuthType:        AuthTypePassword,
	},
}
var configSection = config.MustRegisterSection(database, defaultConfig)

// DbConfig is used to for initiating the database connection with the store that holds registered
// entities (e.g. workflows, tasks, launch plans...)
type DbConfig struct {
	MaxIdleConnections int             `json:"maxIdleConnections" pflag:",maxIdleConnections sets the maximum number of connections in the idle connection pool."`
	MaxOpenConnections int             `json:"maxOpenConnections" pflag:",maxOpenConnections sets the maximum number of open connections to the database."`
	ConnMaxLifeTime    config.Duration `json:"connMaxLifeTime" pflag:",sets the maximum amount of time a connection may be reused"`
	Postgres           PostgresConfig  `json:"postgres,omitempty"`
}

// PostgresConfig includes specific config options for opening a connection to a postgres database.
type PostgresConfig struct {
	Host            string `json:"host" pflag:",The host name of the database server"`
	ReadReplicaHost string `json:"readReplicaHost" pflag:",The host name of the read replica database server"`
	Port            int    `json:"port" pflag:",The port name of the database server"`
	DbName          string `json:"dbname" pflag:",The database name"`
	User            string `json:"username" pflag:",The database user who is connecting to the server."`
	// With AuthType 'password', either Password or PasswordPath must be set. With 'iam', neither may be set.
	Password     string `json:"password" pflag:",The database password."`
	PasswordPath string `json:"passwordPath" pflag:",Points to the file containing the database password."`
	ExtraOptions string `json:"options" pflag:",See http://gorm.io/docs/connecting_to_the_database.html for available options passed, in addition to the above."`
	Debug        bool   `json:"debug" pflag:" Whether or not to start the database connection with debug mode enabled."`
	// With AuthType 'iam', Host and ReadReplicaHost must be RDS/Aurora endpoints: tokens are signed for
	// the configured host, and AWS rejects tokens signed for a custom DNS name such as a Route 53 CNAME.
	// Only the first host of a multi-host value gets a token.
	AuthType string `json:"authType" pflag:",Either 'password' (default) or 'iam'. With 'iam', each new connection authenticates with a fresh AWS RDS IAM auth token instead of Password/PasswordPath."`
	Region   string `json:"region" pflag:",The AWS region of the RDS instance, used to sign IAM auth tokens. Defaults to the region the AWS SDK resolves."`
}

var emptyPostgresConfig = PostgresConfig{}

func (s PostgresConfig) IsEmpty() bool {
	return s == emptyPostgresConfig
}

// IsIAMAuth reports whether connections authenticate with AWS RDS IAM auth tokens.
func (s PostgresConfig) IsIAMAuth() bool {
	return s.AuthType == AuthTypeIAM
}

// Validate rejects unknown auth types, and IAM auth combined with a static password,
// a password file or options that connect without SSL first (RDS only accepts IAM auth over SSL).
// An empty AuthType means password auth.
func (s PostgresConfig) Validate() error {
	switch s.AuthType {
	case "", AuthTypePassword:
		return nil
	case AuthTypeIAM:
		if s.Password != "" || s.PasswordPath != "" {
			return fmt.Errorf("database postgres authType %q cannot be combined with password or passwordPath "+
				"(password defaults to %q; set it to \"\" explicitly)", AuthTypeIAM, postgresStr)
		}
		// Parse options the way pgx will, so spacing and PGSSLMODE are honoured. A nil TLSConfig means the
		// first attempt is plaintext (sslmode=disable or allow). The placeholder TCP host stops pgconn
		// defaulting to a Unix socket, which never uses TLS.
		opts, err := pgconn.ParseConfig("host=rds.invalid " + s.ExtraOptions)
		if err != nil {
			return fmt.Errorf("parse database postgres options: %w", err)
		}
		if opts.TLSConfig == nil {
			return fmt.Errorf("database postgres authType %q requires SSL; set options to sslmode=require or stricter", AuthTypeIAM)
		}
		return nil
	default:
		return fmt.Errorf("unknown database postgres authType %q, expected %q or %q", s.AuthType, AuthTypePassword, AuthTypeIAM)
	}
}

func GetConfig() *DbConfig {
	databaseConfig := configSection.GetConfig().(*DbConfig)
	return databaseConfig
}
