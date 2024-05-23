package cache

import (
	"fmt"
	"sync"
	"sync/atomic"

	"golang.org/x/oauth2"
)

type TokenCacheInMemoryProvider struct {
	token atomic.Value
	mu    *sync.Mutex
	cond  *sync.Cond
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

// CondWait  adds the current go routine to the condition waitlist and waits for another go routine to notify using CondBroadcast
// The current usage is that one who was able to acquire the lock using TryLock is the one who gets a valid token and notifies all the waitlist requesters so that they can use the new valid token.
// It also locks the Locker in the condition variable as the semantics of Wait is that it unlocks the Locker after adding
// the consumer to the waitlist and before blocking on notification.
func (t *TokenCacheInMemoryProvider) CondWait() {
	t.cond.L.Lock()
	t.cond.Wait()
	t.cond.L.Unlock()
}

// CondBroadcast signals the condition.
func (t *TokenCacheInMemoryProvider) CondBroadcast() {
	t.cond.Broadcast()
}

func NewTokenCacheInMemoryProvider() *TokenCacheInMemoryProvider {
	return &TokenCacheInMemoryProvider{
		mu:    &sync.Mutex{},
		token: atomic.Value{},
		cond:  sync.NewCond(&sync.Mutex{}),
	}
}
