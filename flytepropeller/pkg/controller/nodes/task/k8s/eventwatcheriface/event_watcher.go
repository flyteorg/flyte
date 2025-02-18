package eventwatcheriface

import (
	"time"

	"k8s.io/apimachinery/pkg/types"
)

//go:generate mockery-v2 --output=mocks --all --with-expecter
type EventWatcher interface {
	List(objectNsName types.NamespacedName, createdAfter time.Time) []*EventInfo
}

// EventInfo stores detail about the event and the timestamp of the first occurrence.
// All other fields are thrown away to conserve space.
type EventInfo struct {
	Reason     string
	Note       string
	CreatedAt  time.Time
	RecordedAt time.Time
}
