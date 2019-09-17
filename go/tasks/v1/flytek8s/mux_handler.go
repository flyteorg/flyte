package flytek8s

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils/labeled"

	"github.com/lyft/flytestdlib/promutils"

	"github.com/lyft/flytestdlib/logger"

	"sigs.k8s.io/controller-runtime/pkg/cache/informertest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	ctrlHandler "sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/apimachinery/pkg/runtime"
)

var instance *flytek8s
var once sync.Once

type flytek8s struct {
	watchNamespace string
	kubeClient     client.Client
	informersCache cache.Cache
}

func (f *flytek8s) InjectClient(c client.Client) error {
	f.kubeClient = c
	return nil
}

func InjectClient(c client.Client) error {
	if instance == nil {
		return fmt.Errorf("instance not initialized")
	}

	return instance.InjectClient(c)
}

func (f *flytek8s) InjectCache(c cache.Cache) error {
	f.informersCache = c
	return nil
}

func InjectCache(c cache.Cache) error {
	if instance == nil {
		return fmt.Errorf("instance not initialized")
	}

	return instance.InjectCache(c)
}

func InitializeFake() client.Client {
	once.Do(func() {
		instance = &flytek8s{
			watchNamespace: "",
		}

		instance.kubeClient = fake.NewFakeClient()
		instance.informersCache = &informertest.FakeInformers{}
	})

	return instance.kubeClient
}

func Initialize(ctx context.Context, watchNamespace string, resyncPeriod time.Duration) (err error) {
	once.Do(func() {
		instance = &flytek8s{
			watchNamespace: watchNamespace,
		}

		kubeConfig := config.GetConfigOrDie()
		instance.kubeClient, err = client.New(kubeConfig, client.Options{})
		if err != nil {
			return
		}

		instance.informersCache, err = cache.New(kubeConfig, cache.Options{
			Namespace: watchNamespace,
			Resync:    &resyncPeriod,
		})

		if err == nil {
			go func() {
				logger.Infof(ctx, "Starting informers cache.")
				err = instance.informersCache.Start(ctx.Done())
				if err != nil {
					logger.Panicf(ctx, "Failed to start informers cache. Error: %v", err)
				}
			}()
		}
	})

	if err != nil {
		return err
	}

	if watchNamespace != instance.watchNamespace {
		return fmt.Errorf("flytek8s is supposed to be used under single namespace."+
			" configured-for: %v, requested-for: %v", instance.watchNamespace, watchNamespace)
	}

	return nil
}

func RegisterResource(_ context.Context, resourceToWatch runtime.Object, handler Handler, metricsScope promutils.Scope) error {
	if instance == nil {
		return fmt.Errorf("instance not initialized")
	}

	if handler == nil {
		return fmt.Errorf("nil Handler for resource %s", resourceToWatch.GetObjectKind())
	}

	src := source.Kind{
		Type: resourceToWatch,
	}

	if _, err := inject.CacheInto(instance.informersCache, &src); err != nil {
		return err
	}

	// TODO: a more unique workqueue name
	q := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(),
		resourceToWatch.GetObjectKind().GroupVersionKind().Kind)

	updateCount := labeled.NewCounter("informer_update", "Update events from informer", metricsScope)
	droppedUpdateCount := labeled.NewCounter("informer_update_dropped", "Update events from informer that have the same resource version", metricsScope)

	err := src.Start(ctrlHandler.Funcs{
		CreateFunc: func(evt event.CreateEvent, q2 workqueue.RateLimitingInterface) {
		},
		UpdateFunc: func(evt event.UpdateEvent, q2 workqueue.RateLimitingInterface) {
			if evt.MetaNew == nil {
				logger.Warn(context.Background(), "Received an Update event with nil MetaNew.")
			} else if evt.MetaOld == nil || evt.MetaOld.GetResourceVersion() != evt.MetaNew.GetResourceVersion() {
				newCtx := contextutils.WithNamespace(context.Background(), evt.MetaNew.GetNamespace())
				updateCount.Inc(newCtx)

				logger.Debugf(newCtx, "Enqueueing owner for updated object [%v/%v]", evt.MetaNew.GetNamespace(), evt.MetaNew.GetName())
				err := handler.Handle(newCtx, evt.ObjectNew)
				if err != nil {
					logger.Warnf(newCtx, "Failed to handle Update event for object [%v]", evt.ObjectNew)
				}
			} else {
				newCtx := contextutils.WithNamespace(context.Background(), evt.MetaNew.GetNamespace())
				droppedUpdateCount.Inc(newCtx)
			}
		},
		DeleteFunc: func(evt event.DeleteEvent, q2 workqueue.RateLimitingInterface) {
		},
		GenericFunc: func(evt event.GenericEvent, q2 workqueue.RateLimitingInterface) {
		},
	}, q)

	return err
}
