package cache

import "golang.org/x/oauth2"

//go:generate mockery -all -case=underscore

// TokenCache defines the interface needed to cache and retrieve oauth tokens.
type TokenCache interface {
	// SaveToken saves the token securely to cache.
	SaveToken(token *oauth2.Token) error

	// GetToken retrieves the token from the cache.
	GetToken() (*oauth2.Token, error)

	// Purge the token from the cache.
	Purge()

	// Lock the cache.
	Lock()

	// Unlock the cache.
	Unlock()
}
