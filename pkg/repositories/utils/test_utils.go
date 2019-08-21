// Shared utils for postgresql tests.
package utils

import (
	"fmt"
	"testing"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/jinzhu/gorm"
)

func GetDbForTest(t *testing.T) *gorm.DB {
	mocket.Catcher.Register()
	db, err := gorm.Open(mocket.DriverName, "fake args")
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to open mock db with err %v", err))
	}
	return db
}
