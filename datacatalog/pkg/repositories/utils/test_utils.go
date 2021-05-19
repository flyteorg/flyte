// Shared utils for postgresql tests.
package utils

import (
	"fmt"
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func GetDbForTest(t *testing.T) *gorm.DB {
	mocket.Catcher.Register()
	db, err := gorm.Open(postgres.New(postgres.Config{DriverName: mocket.DriverName}))
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to open mock db with err %v", err))
	}
	return db
}
