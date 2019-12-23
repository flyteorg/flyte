package backoff

import (
	"context"
	"fmt"
	"time"

	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/k8s"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/clock"
)

// Controller is a name-spaced collection of back-off handlers
type Controller struct {
	// Controller.Clock allows the use of fake clock when testing
	Clock             clock.Clock
	backOffHandlerMap HandlerMap
}

func (m *Controller) GetOrCreateHandler(ctx context.Context, key string, backOffBaseSecond int, maxBackOffDuration time.Duration) *ComputeResourceAwareBackOffHandler {
	h, loaded := m.backOffHandlerMap.LoadOrStore(key, &ComputeResourceAwareBackOffHandler{
		SimpleBackOffBlocker: &SimpleBackOffBlocker{
			Clock:              m.Clock,
			BackOffBaseSecond:  backOffBaseSecond,
			BackOffExponent:    0,
			NextEligibleTime:   m.Clock.Now(),
			MaxBackOffDuration: maxBackOffDuration,
		}, ComputeResourceCeilings: &ComputeResourceCeilings{
			computeResourceCeilings: v1.ResourceList{},
		},
	})

	if loaded {
		logger.Infof(ctx, "The back-off handler for [%v] has been loaded.\n", key)
	} else {
		logger.Infof(ctx, "The back-off handler for [%v] has been created.\n", key)
	}
	if ret, casted := h.(*ComputeResourceAwareBackOffHandler); casted {
		return ret
	}
	return nil
}

func (m *Controller) GetBackOffHandler(key string) (*ComputeResourceAwareBackOffHandler, bool) {
	return m.backOffHandlerMap.Get(key)
}

func (m *Controller) CreateBackOffHandler(ctx context.Context, key string, backOffBaseSecond int, maxBackOffDuration time.Duration) *ComputeResourceAwareBackOffHandler {
	m.backOffHandlerMap.Set(key, &ComputeResourceAwareBackOffHandler{
		SimpleBackOffBlocker: &SimpleBackOffBlocker{
			Clock:              m.Clock,
			BackOffBaseSecond:  backOffBaseSecond,
			BackOffExponent:    0,
			NextEligibleTime:   m.Clock.Now(),
			MaxBackOffDuration: maxBackOffDuration,
		},
		ComputeResourceCeilings: &ComputeResourceCeilings{
			computeResourceCeilings: v1.ResourceList{},
		},
	})
	h, _ := m.backOffHandlerMap.Get(key)
	h.reset()
	logger.Infof(ctx, "The back-off handler for [%v] has been created.\n", key)
	return h
}

func ComposeResourceKey(o k8s.Resource) string {
	return fmt.Sprintf("%v,%v", o.GroupVersionKind().String(), o.GetNamespace())
}

func NewController(ctx context.Context) *Controller {
	logger.Infof(ctx, "Initializing the back-off controller.\n")
	return &Controller{
		Clock:             clock.RealClock{},
		backOffHandlerMap: HandlerMap{},
	}
}
