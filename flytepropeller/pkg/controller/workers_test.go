package controller

import (
	"context"
	"sync"
	"testing"

	"github.com/flyteorg/flytepropeller/pkg/controller/config"

	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

var testLocalScope2 = promutils.NewScope("worker_pool")

type testHandler struct {
	InitCb   func(ctx context.Context) error
	HandleCb func(ctx context.Context, namespace, key string) error
}

func (t *testHandler) Initialize(ctx context.Context) error {
	return t.InitCb(ctx)
}

func (t *testHandler) Handle(ctx context.Context, namespace, key string) error {
	return t.HandleCb(ctx, namespace, key)
}

func simpleWorkQ(ctx context.Context, t *testing.T, testScope promutils.Scope) CompositeWorkQueue {
	cfg := config.CompositeQueueConfig{}
	q, err := NewCompositeWorkQueue(ctx, cfg, testScope)
	assert.NoError(t, err)
	assert.NotNil(t, q)
	return q
}

func TestWorkerPool_Run(t *testing.T) {
	ctx := context.TODO()
	l := testLocalScope2.NewSubScope("new")
	h := &testHandler{}
	q := simpleWorkQ(ctx, t, l)
	w := NewWorkerPool(ctx, l, q, h)
	assert.NotNil(t, w)

	t.Run("initcalled", func(t *testing.T) {

		initCalled := false
		h.InitCb = func(ctx context.Context) error {
			initCalled = true
			return nil
		}

		assert.NoError(t, w.Initialize(ctx))
		assert.True(t, initCalled)
	})

	// Bad TEST :(. We create 2 waitgroups, one will wait for the Run function to exit (called wg)
	// Other is called handleReceived, waits for receiving a handle
	// The flow is,
	// - start the poll loop
	// - add a key `x`
	// - wait for `x` to be handled
	// - cancel the loop
	// - wait for loop to exit
	t.Run("run", func(t *testing.T) {
		childCtx, cancel := context.WithCancel(ctx)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			assert.NoError(t, w.Run(childCtx, 1, func() bool {
				return true
			}))
			wg.Done()
		}()

		handleReceived := sync.WaitGroup{}
		handleReceived.Add(1)

		h.HandleCb = func(ctx context.Context, namespace, key string) error {
			if key == "x" {
				handleReceived.Done()
			} else {
				assert.FailNow(t, "x expected")
			}
			return nil
		}
		q.Add("x")
		handleReceived.Wait()

		cancel()
		wg.Wait()
	})
}
