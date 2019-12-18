package backoff_manager

import (
	"context"
	"fmt"
	"time"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/k8s"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/clock"
)

var (
	backOffBaseSecond  int // b in b^n
	maxBackOffDuration time.Duration
)

type BackOffManager struct {
	Clock             clock.Clock
	backOffHandlerMap BackOffHandlerMap
	defaultIsActive   bool
}

func (m *BackOffManager) GetBackOffHandler(key string) (*ComputeResourceAwareBackOffHandler, bool) {
	return m.backOffHandlerMap.Get(key)
}

func (m *BackOffManager) CreateBackOffHandler(key string, backOffBaseSecond int, maxBackOffDuration time.Duration) *ComputeResourceAwareBackOffHandler {
	m.backOffHandlerMap.Set(key, &ComputeResourceAwareBackOffHandler{
		SimpleBackOffBlocker: &SimpleBackOffBlocker{
			Clock:              m.Clock,
			BackOffBaseSecond:  backOffBaseSecond,
			BackOffExponent:    0,
			NextEligibleTime:   m.Clock.Now(),
			MaxBackOffDuration: maxBackOffDuration,
		},
		// TODO changhong: initialize this field with proper value
		ComputeResourceCeilings: &ComputeResourceCeilings{
			computeResourceCeilings: v1.ResourceList{},
		},
	})
	h, _ := m.backOffHandlerMap.Get(key)
	h.ComputeResourceCeilings.resetAll()
	h.SimpleBackOffBlocker.reset()
	return h
}

func ComposeResourceKey(o k8s.Resource) string {
	return fmt.Sprintf("%v,%v", o.GroupVersionKind().String(), o.GetNamespace())
}

func NewBackOffManager(ctx context.Context, baseSecond int, maxDuration time.Duration) *BackOffManager {
	backOffBaseSecond = baseSecond
	maxBackOffDuration = maxDuration

	return &BackOffManager{
		Clock:             clock.RealClock{},
		backOffHandlerMap: BackOffHandlerMap{},
	}
}
