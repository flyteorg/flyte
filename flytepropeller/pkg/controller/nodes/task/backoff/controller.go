package backoff

import (
	"context"
	"fmt"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	stdAtomic "github.com/flyteorg/flytestdlib/atomic"

	"github.com/flyteorg/flytestdlib/logger"

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
			BackOffExponent:    stdAtomic.NewUint32(0),
			NextEligibleTime:   NewAtomicTime(m.Clock.Now()),
			MaxBackOffDuration: maxBackOffDuration,
		}, ComputeResourceCeilings: &ComputeResourceCeilings{
			computeResourceCeilings: NewSyncResourceList(),
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

func ComposeResourceKey(o client.Object) string {
	return fmt.Sprintf("%v,%v", o.GetObjectKind().GroupVersionKind().String(), o.GetNamespace())
}

func NewController(ctx context.Context) *Controller {
	logger.Infof(ctx, "Initializing the back-off controller.\n")
	return &Controller{
		Clock:             clock.RealClock{},
		backOffHandlerMap: HandlerMap{},
	}
}
