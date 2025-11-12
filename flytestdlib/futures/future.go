// Package futures implements a simple Async Futures for golang
package futures

import (
	"context"
	"fmt"
	"sync"
)

// Provides a Future API for asynchronous completion of tasks
type Future interface {
	// Returns true if the Future is ready and either a value or error is available. Once Ready returns True, Get should return immediately
	Ready() bool
	// Get is a potentially blocking call, that returns the asynchronously computed value or an error
	// If Get is called before Ready() returns True, then it will block till the future has been completed
	Get(ctx context.Context) (interface{}, error)
}

// This is a synchronous future, where the values are available immediately on construction. This is used to maintain a synonymous API with both
// Async and Sync tasks
type SyncFuture struct {
	// The actual value
	val interface{}
	// OR an error
	err error
}

// Always returns true
func (s SyncFuture) Ready() bool {
	return true
}

// returns the previously provided value / error
func (s *SyncFuture) Get(_ context.Context) (interface{}, error) {
	return s.val, s.err
}

// Creates a new "completed" future that matches the async computation api
func NewSyncFuture(val interface{}, err error) *SyncFuture {
	return &SyncFuture{
		val: val,
		err: err,
	}
}

// ErrAsyncFutureCanceled is returned when the async future is cancelled by invoking the cancel function on the context
var ErrAsyncFutureCanceled = fmt.Errorf("async future was canceled")

// An asynchronously completing future
type AsyncFuture struct {
	sync.Mutex
	doneChannel chan bool
	cancelFn    context.CancelFunc
	// The actual value
	val interface{}
	// Or an error
	err   error
	ready bool
}

func (f *AsyncFuture) set(val interface{}, err error) {
	f.Lock()
	defer f.Unlock()
	f.val = val
	f.err = err
	f.ready = true
	f.doneChannel <- true
}

func (f *AsyncFuture) get() (interface{}, error) {
	f.Lock()
	defer f.Unlock()
	return f.val, f.err
}

// returns whether the future is completed
func (f *AsyncFuture) Ready() bool {
	f.Lock()
	defer f.Unlock()
	return f.ready
}

// Returns results (interface{} or an error) OR blocks till the results are available.
// If context is cancelled while waiting for results, an ErrAsyncFutureCanceled is returned
func (f *AsyncFuture) Get(ctx context.Context) (interface{}, error) {
	select {
	case <-ctx.Done():
		f.cancelFn()
		return nil, ErrAsyncFutureCanceled
	case <-f.doneChannel:
		return f.get()
	}
}

// Creates a new Async future, that will call the method `closure` and return the results from the execution of
// this method
func NewAsyncFuture(ctx context.Context, closure func(context.Context) (interface{}, error)) *AsyncFuture {
	childCtx, cancel := context.WithCancel(ctx)
	f := &AsyncFuture{
		doneChannel: make(chan bool, 1),
		cancelFn:    cancel,
	}

	go func(ctx2 context.Context, fut *AsyncFuture) {
		val, err := closure(ctx2)
		fut.set(val, err)
	}(childCtx, f)
	return f
}
