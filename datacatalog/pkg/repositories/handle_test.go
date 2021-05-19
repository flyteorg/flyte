package repositories

import (
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/stretchr/testify/assert"

	"database/sql/driver"

	"github.com/flyteorg/datacatalog/pkg/repositories/utils"
)

func TestCreateDB(t *testing.T) {
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	checkExists := false
	GlobalMock.NewMock().WithQuery(
		`SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)%!(EXTRA string=testDB)`).WithCallback(
		func(s string, values []driver.NamedValue) {
			checkExists = true
		},
	).WithReply([]map[string]interface{}{
		{"exists": false},
	})

	createdDB := false

	// NOTE: unfortunately mocket does not support checking CREATE statements, but let's match the suffix
	GlobalMock.NewMock().WithQuery(
		`DATABASE testDB`).WithCallback(
		func(s string, values []driver.NamedValue) {
			assert.Equal(t, "CREATE DATABASE testDB", s)
			createdDB = true
		},
	)

	db := utils.GetDbForTest(t)
	dbHandle := &DBHandle{
		db: db,
	}
	_ = dbHandle.CreateDB("testDB")
	assert.True(t, checkExists)
	assert.True(t, createdDB)
}

func TestDBAlreadyExists(t *testing.T) {
	GlobalMock := mocket.Catcher.Reset()
	GlobalMock.Logging = true

	checkExists := false
	GlobalMock.NewMock().WithQuery(
		`SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)%!(EXTRA string=testDB)`).WithCallback(
		func(s string, values []driver.NamedValue) {
			checkExists = true
		},
	).WithReply([]map[string]interface{}{
		{"exists": true},
	})

	createdDB := false
	GlobalMock.NewMock().WithQuery(
		`DATABASE testDB`).WithCallback(
		func(s string, values []driver.NamedValue) {
			createdDB = false
		},
	)

	db := utils.GetDbForTest(t)
	dbHandle := &DBHandle{
		db: db,
	}
	err := dbHandle.CreateDB("testDB")
	assert.NoError(t, err)
	assert.True(t, checkExists)
	assert.False(t, createdDB)
}
