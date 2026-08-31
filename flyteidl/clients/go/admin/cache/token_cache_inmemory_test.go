package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCondBroadcastWakesAllWaiters(t *testing.T) {
	c := NewTokenCacheInMemoryProvider()

	const waiters = 5
	var wg sync.WaitGroup
	wg.Add(waiters)
	for i := 0; i < waiters; i++ {
		go func() {
			defer wg.Done()
			c.CondWait()
		}()
	}

	// Give the waiters a moment to park before broadcasting.
	time.Sleep(50 * time.Millisecond)
	c.CondBroadcast()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("waiters were not woken by CondBroadcast")
	}
}

func TestCondWaitTimesOutWithoutBroadcast(t *testing.T) {
	prev := condWaitTimeout
	condWaitTimeout = 50 * time.Millisecond
	defer func() { condWaitTimeout = prev }()

	c := NewTokenCacheInMemoryProvider()

	start := time.Now()
	c.CondWait() // no broadcast ever comes
	assert.Less(t, time.Since(start), 5*time.Second, "CondWait must return once the timeout elapses")
}

func TestConcurrentCondBroadcastDoesNotPanic(t *testing.T) {
	c := NewTokenCacheInMemoryProvider()

	// Concurrent broadcasters must each close a distinct channel (Swap
	// semantics) — a shared close would panic with "close of closed channel".
	const broadcasters = 16
	var wg sync.WaitGroup
	wg.Add(broadcasters)
	for i := 0; i < broadcasters; i++ {
		go func() {
			defer wg.Done()
			c.CondBroadcast()
		}()
	}
	wg.Wait()
}

func TestCondWaitAfterMissedBroadcastTimesOut(t *testing.T) {
	prev := condWaitTimeout
	condWaitTimeout = 50 * time.Millisecond
	defer func() { condWaitTimeout = prev }()

	c := NewTokenCacheInMemoryProvider()

	// The broadcast fires before the waiter starts waiting — the classic lost
	// wakeup. The waiter must still return via the timeout.
	c.CondBroadcast()

	start := time.Now()
	c.CondWait()
	assert.Less(t, time.Since(start), 5*time.Second, "CondWait must not hang on a missed broadcast")
}
