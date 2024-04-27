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

func (t *TokenCacheInMemoryProvider) Purge() {
	t.token.Store(nil)
}

func (t *TokenCacheInMemoryProvider) Lock() {
	t.mu.Lock()
}

func (t *TokenCacheInMemoryProvider) Unlock() {
	t.mu.Unlock()
}

func NewTokenCacheInMemoryProvider() *TokenCacheInMemoryProvider {
	return &TokenCacheInMemoryProvider{
		mu:    &sync.Mutex{},
		token: atomic.Value{},
	}
}
