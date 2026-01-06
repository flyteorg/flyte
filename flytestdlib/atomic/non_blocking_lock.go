package atomic

// Lock that provides TryLock method instead of blocking lock
type NonBlockingLock interface {
	TryLock() bool
	Release()
}

func NewNonBlockingLock() NonBlockingLock {
	return &nonBlockingLock{lock: NewBool(false)}
}

type nonBlockingLock struct {
	lock Bool
}

func (n *nonBlockingLock) TryLock() bool {
	return n.lock.CompareAndSwap(false, true)
}

func (n *nonBlockingLock) Release() {
	n.lock.Store(false)
}
