// Shared utils for postgresql tests.
package gormimpl

import (
	"fmt"
	"testing"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"

	mocket "github.com/Selvatico/go-mocket"
	"github.com/jinzhu/gorm"
	"github.com/lyft/flyteadmin/pkg/common"
)

const project = "project"
const domain = "domain"
const name = "name"
const description = "description"
const resourceType = core.ResourceType_WORKFLOW
const version = "XYZ"

func GetDbForTest(t *testing.T) *gorm.DB {
	mocket.Catcher.Register()
	db, err := gorm.Open(mocket.DriverName, "fake args")
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to open mock db with err %v", err))
	}
	return db
}

func getEqualityFilter(entity common.Entity, field string, value interface{}) common.InlineFilter {
	filter, _ := common.NewSingleValueFilter(entity, common.Equal, field, value)
	return filter
}
