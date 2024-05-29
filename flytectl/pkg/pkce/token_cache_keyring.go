package pkce

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/cache"
	"github.com/flyteorg/flyte/flytestdlib/logger"

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
	condLocker  *cache.NoopLocker
	cond        *sync.Cond
}

func (t *TokenCacheKeyringProvider) PurgeIfEquals(existing *oauth2.Token) (bool, error) {
	if existingBytes, err := json.Marshal(existing); err != nil {
		return false, fmt.Errorf("unable to marshal token to save in cache due to %w", err)
	} else if tokenJSON, err := keyring.Get(t.ServiceName, t.ServiceUser); err != nil {
		logger.Warnf(context.Background(), "unable to read token from cache but not failing the purge as the token might not have been saved at all. Error: %v", err)
		return true, nil
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

// CondWait  adds the current go routine to the condition waitlist and waits for another go routine to notify using CondBroadcast
// The current usage is that one who was able to acquire the lock using TryLock is the one who gets a valid token and notifies all the waitlist requesters so that they can use the new valid token.
// It also locks the Locker in the condition variable as the semantics of Wait is that it unlocks the Locker after adding
// the consumer to the waitlist and before blocking on notification.
func (t *TokenCacheKeyringProvider) CondWait() {
	t.cond.L.Lock()
	t.cond.Wait()
	t.cond.L.Unlock()
}

// CondBroadcast broadcasts the condition.
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
	condLocker := &cache.NoopLocker{}
	return &TokenCacheKeyringProvider{
		mu:          &sync.Mutex{},
		condLocker:  condLocker,
		cond:        sync.NewCond(condLocker),
		ServiceName: serviceName,
		ServiceUser: serviceUser,
	}
}
