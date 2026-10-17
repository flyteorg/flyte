package k8s

import "sync"

// subscription delivers one WatchForUpdates stream's child updates without ever
// dropping one.
//
// A parent's SDK informer only learns a child has finished from that child's
// terminal update, and it never re-lists, so a dropped terminal update hangs the
// parent forever (https://github.com/flyteorg/flyte/issues/8104). Instead of a
// bounded channel that drops on overflow, notify parks updates in a map keyed by
// child action name and a pump goroutine feeds them into out. While an update is
// parked, a newer one for the same child replaces it: the informer only needs each
// child's latest state, and memory is bounded by the number of children rather
// than by how fast the stream drains.
type subscription struct {
	out chan *ActionUpdate

	mu sync.Mutex
	// pending holds the latest undelivered update per child action name.
	pending map[string]*ActionUpdate
	// order lists pending's keys in the order they were first parked, so children
	// are delivered fairly while one keeps updating.
	order []string

	// wake has capacity 1: a pending signal means "pending may be non-empty".
	wake chan struct{}
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

// notify parks update for delivery. It never blocks.
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

// close stops the pump, which then closes out.
func (s *subscription) close() {
	close(s.done)
}

// pump moves parked updates into out until close is called. It owns out and
// closes it on exit, so a send can never race the close.
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

// next takes the oldest parked update, or returns nil when none are parked.
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
