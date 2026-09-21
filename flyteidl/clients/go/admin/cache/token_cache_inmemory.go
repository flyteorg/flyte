package cache

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/oauth2"
)

// condWaitTimeout bounds CondWait. Cond-style notification has no memory: a
// broadcast that fires between a waiter's decision to wait and the wait itself
// is lost, and a refresher that exits without broadcasting never wakes anyone.
// The bound turns both cases into a short delay followed by the caller's
// retry instead of a permanent hang of the calling goroutine.
// Var, not const, so tests can shorten it.
var condWaitTimeout = 30 * time.Second

type TokenCacheInMemoryProvider struct {
	token atomic.Value
	mu    *sync.Mutex

	// notify is closed and replaced on every CondBroadcast; waiters select on
	// the channel they observed, so a broadcast wakes exactly the waiters that
	// started waiting before it.
	notify atomic.Pointer[chan struct{}]
}

func (t *TokenCacheInMemoryProvider) SaveToken(token *oauth2.Token) error {
	t.token.Store(token)
	return nil
}

func (t *TokenCacheInMemoryProvider) GetToken() (*oauth2.Token, error) {
	tkn := t.token.Load()
	if tkn == nil {
		return nil, fmt.Errorf("cannot find token in cache")
	}
	return tkn.(*oauth2.Token), nil
}

func (t *TokenCacheInMemoryProvider) PurgeIfEquals(existing *oauth2.Token) (bool, error) {
	// Add an empty token since we can't mark it nil using Compare and swap
	return t.token.CompareAndSwap(existing, &oauth2.Token{}), nil
}

func (t *TokenCacheInMemoryProvider) Lock() {
	t.mu.Lock()
}

func (t *TokenCacheInMemoryProvider) TryLock() bool {
	return t.mu.TryLock()
}

func (t *TokenCacheInMemoryProvider) Unlock() {
	t.mu.Unlock()
}

// CondWait blocks until another goroutine calls CondBroadcast, or until
// condWaitTimeout elapses — whichever comes first. The current usage is that
// the goroutine that acquired the lock via TryLock refreshes the token and
// broadcasts so waiters can retry with the new token. The timeout guarantees a
// waiter is never parked forever when the broadcast is missed (raced) or never
// sent (refresh failed); after waking it simply retries and, if the token is
// still invalid, fails or refreshes on its own.
func (t *TokenCacheInMemoryProvider) CondWait() {
	ch := *t.notify.Load()

	select {
	case <-ch:
	case <-time.After(condWaitTimeout):
	}
}

// NoopLocker has empty implementation of Locker interface
type NoopLocker struct {
}

func (*NoopLocker) Lock() {

}
func (*NoopLocker) Unlock() {
}

// CondBroadcast wakes every goroutine currently blocked in CondWait by closing
// the notification channel and installing a fresh one for future waiters.
// Swap guarantees each concurrent broadcaster closes a distinct channel, so a
// channel is never closed twice.
func (t *TokenCacheInMemoryProvider) CondBroadcast() {
	newCh := make(chan struct{})
	old := t.notify.Swap(&newCh)
	close(*old)
}

func NewTokenCacheInMemoryProvider() *TokenCacheInMemoryProvider {
	t := &TokenCacheInMemoryProvider{
		mu:    &sync.Mutex{},
		token: atomic.Value{},
	}
	ch := make(chan struct{})
	t.notify.Store(&ch)
	return t
}
