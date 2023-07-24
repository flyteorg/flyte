// Shared utils for postgresql tests.
package gormimpl

import (
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	mocket "github.com/Selvatico/go-mocket"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const project = "project"
const domain = "domain"
const name = "name"
const description = "description"
const resourceType = core.ResourceType_WORKFLOW
const version = "XYZ"

func GetDbForTest(t *testing.T) *gorm.DB {
	mocket.Catcher.Register()
	db, err := gorm.Open(postgres.New(postgres.Config{DriverName: mocket.DriverName}))
	if err != nil {
		t.Fatalf("Failed to open mock db with err %v", err)
	}
	return db
}

func getEqualityFilter(entity common.Entity, field string, value interface{}) common.InlineFilter {
	filter, _ := common.NewSingleValueFilter(entity, common.Equal, field, value)
	return filter
}
