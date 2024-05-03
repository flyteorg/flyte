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
	return t.token.CompareAndSwap(existing, nil), nil
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

// CondSignal signals the condition.
func (t *TokenCacheInMemoryProvider) CondSignal() {
	t.cond.Signal()
}

func NewTokenCacheInMemoryProvider() *TokenCacheInMemoryProvider {
	return &TokenCacheInMemoryProvider{
		mu:    &sync.Mutex{},
		token: atomic.Value{},
		cond:  sync.NewCond(&sync.Mutex{}),
	}
}
