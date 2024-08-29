package cache

import (
	"fmt"
	"sync"
	"sync/atomic"

	"golang.org/x/oauth2"
)

type TokenCacheInMemoryProvider struct {
	token      atomic.Value
	mu         *sync.Mutex
	condLocker *NoopLocker
	cond       *sync.Cond
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
// We use the condLocker which is noOp locker to get added to waitlist for notifications.
// The underlying notifcationList doesn't need to be guarded as it implementation is atomic and is thread safe
// Refer https://go.dev/src/runtime/sema.go
// Following is the function and its comments
// notifyListAdd adds the caller to a notify list such that it can receive
// notifications. The caller must eventually call notifyListWait to wait for
// such a notification, passing the returned ticket number.
//
//	func notifyListAdd(l *notifyList) uint32 {
//		// This may be called concurrently, for example, when called from
//		// sync.Cond.Wait while holding a RWMutex in read mode.
//		return l.wait.Add(1) - 1
//	}
func (t *TokenCacheInMemoryProvider) CondWait() {
	t.condLocker.Lock()
	t.cond.Wait()
	t.condLocker.Unlock()
}

// NoopLocker has empty implementation of Locker interface
type NoopLocker struct {
}

func (*NoopLocker) Lock() {

}
func (*NoopLocker) Unlock() {
}

// CondBroadcast signals the condition.
func (t *TokenCacheInMemoryProvider) CondBroadcast() {
	t.cond.Broadcast()
}

func NewTokenCacheInMemoryProvider() *TokenCacheInMemoryProvider {
	condLocker := &NoopLocker{}
	return &TokenCacheInMemoryProvider{
		mu:         &sync.Mutex{},
		token:      atomic.Value{},
		condLocker: condLocker,
		cond:       sync.NewCond(condLocker),
	}
}
