package cache

import (
	"fmt"

	"golang.org/x/oauth2"
)

type TokenCacheInMemoryProvider struct {
	token *oauth2.Token
}

func (t *TokenCacheInMemoryProvider) SaveToken(token *oauth2.Token) error {
	t.token = token
	return nil
}

func (t TokenCacheInMemoryProvider) GetToken() (*oauth2.Token, error) {
	if t.token == nil {
		return nil, fmt.Errorf("cannot find token in cache")
	}

	return t.token, nil
}
