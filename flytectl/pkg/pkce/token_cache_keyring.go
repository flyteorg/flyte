package pkce

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/cache"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

const (
	KeyRingServiceUser = "flytectl-user"
	KeyRingServiceName = "flytectl"
)

// TokenCacheKeyringProvider wraps the logic to save and retrieve tokens from the OS's keyring implementation.
type TokenCacheKeyringProvider struct {
	ServiceName string
	ServiceUser string
	mu          *sync.Mutex
	cond        *sync.Cond
}

func (t *TokenCacheKeyringProvider) PurgeIfEquals(existing *oauth2.Token) (bool, error) {
	if existingBytes, err := json.Marshal(existing); err != nil {
		return false, fmt.Errorf("unable to marshal token to save in cache due to %w", err)
	} else if tokenJSON, err := keyring.Get(t.ServiceName, t.ServiceUser); err != nil {
		if err.Error() == "secret not found in keyring" {
			return false, fmt.Errorf("unable to read token from cache. Error: %w", cache.ErrNotFound)
		}

		return false, fmt.Errorf("unable to read token from cache. Error: %w", err)
	} else if tokenJSON != string(existingBytes) {
		return false, nil
	}

	_ = keyring.Delete(t.ServiceName, t.ServiceUser)
	return true, nil
}

func (t *TokenCacheKeyringProvider) Lock() {
	t.mu.Lock()
}

func (t *TokenCacheKeyringProvider) Unlock() {
	t.mu.Unlock()
}

// TryLock the cache.
func (t *TokenCacheKeyringProvider) TryLock() bool {
	return t.mu.TryLock()
}

// CondWait waits for the condition to be true.
func (t *TokenCacheKeyringProvider) CondWait() {
	t.cond.Wait()
}

// CondBroadcast signals the condition.
func (t *TokenCacheKeyringProvider) CondBroadcast() {
	t.cond.Broadcast()
}

func (t *TokenCacheKeyringProvider) SaveToken(token *oauth2.Token) error {
	var tokenBytes []byte
	if token.AccessToken == "" {
		return fmt.Errorf("cannot save empty token with expiration %v", token.Expiry)
	}

	var err error
	if tokenBytes, err = json.Marshal(token); err != nil {
		return fmt.Errorf("unable to marshal token to save in cache due to %w", err)
	}

	// set token in keyring
	if err = keyring.Set(t.ServiceName, t.ServiceUser, string(tokenBytes)); err != nil {
		return fmt.Errorf("unable to save token. Error: %w", err)
	}

	return nil
}

func (t *TokenCacheKeyringProvider) GetToken() (*oauth2.Token, error) {
	// get saved token
	tokenJSON, err := keyring.Get(t.ServiceName, t.ServiceUser)
	if len(tokenJSON) == 0 {
		return nil, fmt.Errorf("no token found in the cache")
	}

	if err != nil {
		return nil, err
	}

	token := oauth2.Token{}
	if err = json.Unmarshal([]byte(tokenJSON), &token); err != nil {
		return nil, fmt.Errorf("unmarshalling error for saved token. Error: %w", err)
	}

	return &token, nil
}

func NewTokenCacheKeyringProvider(serviceName, serviceUser string) *TokenCacheKeyringProvider {
	return &TokenCacheKeyringProvider{
		mu:          &sync.Mutex{},
		cond:        sync.NewCond(&sync.Mutex{}),
		ServiceName: serviceName,
		ServiceUser: serviceUser,
	}
}
