package k8s

import "sync"

// subscription delivers child updates to one WatchForUpdates stream without dropping any.
// The SDK informer never re-lists, so a dropped terminal update hangs the parent
// (https://github.com/flyteorg/flyte/issues/8104).
type subscription struct {
	// out is the channel returned by Subscribe. Only pump sends on it.
	out chan *ActionUpdate

	mu sync.Mutex
	// pending holds the latest undelivered update per child action name.
	// A newer update replaces a parked one, so memory is bounded by the number of children.
	pending map[string]*ActionUpdate
	// order holds pending's keys in the order they were first parked.
	order []string

	// wake signals pump that pending may be non-empty.
	wake chan struct{}
	// done is closed to stop pump.
	done chan struct{}
}

func newSubscription(bufferSize int) *subscription {
	s := &subscription{
		out:     make(chan *ActionUpdate, bufferSize),
		pending: make(map[string]*ActionUpdate),
		wake:    make(chan struct{}, 1),
		done:    make(chan struct{}),
	}
	go s.pump()
	return s
}

// notify parks an update for pump to deliver. It never blocks.
func (s *subscription) notify(update *ActionUpdate) {
	key := update.ActionID.GetName()

	s.mu.Lock()
	if _, parked := s.pending[key]; !parked {
		s.order = append(s.order, key)
	}
	s.pending[key] = update
	s.mu.Unlock()

	select {
	case s.wake <- struct{}{}:
	default:
	}
}

// close stops pump, which then closes out.
func (s *subscription) close() {
	close(s.done)
}

// pump moves parked updates into out until close is called, then closes out.
func (s *subscription) pump() {
	defer close(s.out)
	for {
		select {
		case <-s.done:
			return
		case <-s.wake:
		}
		for update := s.next(); update != nil; update = s.next() {
			select {
			case s.out <- update:
			case <-s.done:
				return
			}
		}
	}
}

// next pops the oldest parked update, or returns nil if none are parked.
func (s *subscription) next() *ActionUpdate {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.order) == 0 {
		return nil
	}
	key := s.order[0]
	s.order[0] = ""
	s.order = s.order[1:]
	update := s.pending[key]
	delete(s.pending, key)
	return update
}
