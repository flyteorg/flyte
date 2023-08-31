package models

import (
	"sync"
	"time"

	"gorm.io/gorm/schema"
	"k8s.io/apimachinery/pkg/util/sets"
)

// This is the base model definition every flyteadmin model embeds.
// This is nearly identical to http://doc.gorm.io/models.html#conventions except that flyteadmin models define their
// own primary keys rather than use the ID as the primary key
type BaseModel struct {
	ID        uint `gorm:"index;autoIncrement"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index"`
}

func modelColumns(v any) sets.String {
	s, err := schema.Parse(v, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		panic(err)
	}
	return sets.NewString(s.DBNames...)
}
