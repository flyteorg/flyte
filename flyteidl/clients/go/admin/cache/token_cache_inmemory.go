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

// CondWait waits for the condition to be true.
func (t *TokenCacheInMemoryProvider) CondWait() {
	t.cond.Wait()
}

// CondBroadcast signals the condition.
func (t *TokenCacheInMemoryProvider) CondBroadcast() {
	t.cond.Broadcast()
}

func NewTokenCacheInMemoryProvider() *TokenCacheInMemoryProvider {
	condMutex := &sync.Mutex{}
	return &TokenCacheInMemoryProvider{
		mu:    condMutex,
		token: atomic.Value{},
		cond:  sync.NewCond(condMutex),
	}
}
