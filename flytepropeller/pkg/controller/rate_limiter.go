package controller

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"k8s.io/client-go/util/workqueue"

	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/interfaces"
)

// limiterAdapter adapts rate.NewLimiter to use the Reservation interface so that it can be used in unittests.
type limiterAdapter struct {
	limiter *rate.Limiter
}

func NewLimiter(r rate.Limit, b int) interfaces.Limiter {
	return &limiterAdapter{rate.NewLimiter(r, b)}
}

func (l *limiterAdapter) Allow() bool {
	return l.limiter.Allow()
}

func (l *limiterAdapter) AllowN(t time.Time, n int) bool {
	return l.limiter.AllowN(t, n)
}

func (l *limiterAdapter) Burst() int {
	return l.limiter.Burst()
}

func (l *limiterAdapter) Limit() rate.Limit {
	return l.limiter.Limit()
}

func (l *limiterAdapter) Reserve() interfaces.Reservation {
	return l.limiter.Reserve()
}

func (l *limiterAdapter) ReserveN(t time.Time, n int) interfaces.Reservation {
	return l.limiter.ReserveN(t, n)
}
func (l *limiterAdapter) SetBurst(newBurst int) {
	l.limiter.SetBurst(newBurst)
}

func (l *limiterAdapter) SetBurstAt(t time.Time, newBurst int) {
	l.limiter.SetBurstAt(t, newBurst)
}

func (l *limiterAdapter) SetLimit(newLimit rate.Limit) {
	l.limiter.SetLimit(newLimit)
}

func (l *limiterAdapter) SetLimitAt(t time.Time, newLimit rate.Limit) {
	l.limiter.SetLimitAt(t, newLimit)
}

func (l *limiterAdapter) Tokens() float64 {
	return l.limiter.Tokens()
}

func (l *limiterAdapter) TokensAt(t time.Time) float64 {
	return l.limiter.TokensAt(t)
}

func (l *limiterAdapter) Wait(ctx context.Context) (err error) {
	return l.limiter.Wait(ctx)
}

func (l *limiterAdapter) WaitN(ctx context.Context, n int) (err error) {
	return l.limiter.WaitN(ctx, n)
}

// Similar to the standard BucketRateLimiter but dedupes items in order to avoid reserving token slots for the
// same item multiple times. Intened to be used with a DelayingQueue, which dedupes items on insertion.
type dedupingBucketRateLimiter struct {
	Limiter      interfaces.Limiter
	mu           sync.Mutex
	reservations map[interface{}]interfaces.Reservation
}

func NewDedupingBucketRateLimiter(limiter interfaces.Limiter) workqueue.RateLimiter {
	return &dedupingBucketRateLimiter{
		Limiter:      limiter,
		reservations: make(map[interface{}]interfaces.Reservation),
	}
}

var _ workqueue.RateLimiter = &dedupingBucketRateLimiter{}

func (r *dedupingBucketRateLimiter) When(item interface{}) time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Check if this item has an outstanding reservation. If so, use it to avoid a duplicate reservation.
	if res, ok := r.reservations[item]; ok && res.Delay() > 0 {
		return res.Delay()
	}
	r.reservations[item] = r.Limiter.Reserve()
	return r.reservations[item].Delay()
}

func (r *dedupingBucketRateLimiter) NumRequeues(interface{}) int {
	return 0
}

func (r *dedupingBucketRateLimiter) Forget(item interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if res, ok := r.reservations[item]; ok && res.Delay() <= 0 {
		delete(r.reservations, item)
	}
}
