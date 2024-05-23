package cache

import (
	"fmt"
	"slices"
	"strings"

	"golang.org/x/oauth2"
)

//go:generate mockery -all -case=underscore

// TokenCacheType defines the type of token cache implementation.
type TokenCacheType string

const (
	// TokenCacheTypeKeyring represents the token cache implementation using the OS's keyring.
	TokenCacheTypeKeyring TokenCacheType = "keyring"
	// TokenCacheTypeInMemory represents the token cache implementation using an in-memory cache.
	TokenCacheTypeInMemory = "inmemory"
	// TokenCacheTypeFilesystem represents the token cache implementation using the local filesystem.
	TokenCacheTypeFilesystem = "filesystem"
)

var AllTokenCacheTypes = []TokenCacheType{TokenCacheTypeKeyring, TokenCacheTypeInMemory, TokenCacheTypeFilesystem}

// String implements pflag.Value interface.
func (t *TokenCacheType) String() string {
	if t == nil {
		return ""
	}
	return string(*t)
}

// Set implements pflag.Value interface.
func (t *TokenCacheType) Set(value string) error {
	if slices.Contains(AllTokenCacheTypes, TokenCacheType(strings.ToLower(value))) {
		*t = TokenCacheType(value)
		return nil
	}

	return fmt.Errorf("%s is an unrecognized token cache type (supported types %v)", value, AllTokenCacheTypes)
}

// Type implements pflag.Value interface.
func (t *TokenCacheType) Type() string {
	return "token-cache-type"
}

// TokenCache defines the interface needed to cache and retrieve oauth tokens.
type TokenCache interface {
	// SaveToken saves the token securely to cache.
	SaveToken(token *oauth2.Token) error

	// Retrieves the token from the cache.
	GetToken() (*oauth2.Token, error)
}
