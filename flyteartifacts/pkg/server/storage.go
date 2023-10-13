package server

import (
	"github.com/flyteorg/flyte/flytestdlib/database"
)

type RDSStorage struct {
	config database.DbConfig
}

func NewStorage() StorageInterface {
	dbCfg := database.GetConfig()

	return &RDSStorage{
		config: *dbCfg,
	}
}
