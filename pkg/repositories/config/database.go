package config

import "gorm.io/gorm/logger"

//go:generate pflags DbConfigSection

// This struct corresponds to the  database section of in the config
type DbConfigSection struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	DbName string `json:"dbname"`
	User   string `json:"username"`
	// Either Password or PasswordPath must be set.
	Password     string `json:"password"`
	PasswordPath string `json:"passwordPath"`
	// See http://gorm.io/docs/connecting_to_the_database.html for available options passed, in addition to the above.
	ExtraOptions string          `json:"options"`
	LogLevel     logger.LogLevel `json:"log_level"`
}

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
