package interfaces

import (
	"context"
	"time"

	"golang.org/x/time/rate"
)

//go:generate mockery --name Limiter --output ../mocks --case=snake --with-expecter
//go:generate mockery --name Reservation --output ../mocks --case=snake --with-expecter

type Limiter interface {
	Allow() bool
	AllowN(t time.Time, n int) bool
	Burst() int
	Limit() rate.Limit
	Reserve() Reservation
	ReserveN(t time.Time, n int) Reservation
	SetBurst(newBurst int)
	SetBurstAt(t time.Time, newBurst int)
	SetLimit(newLimit rate.Limit)
	SetLimitAt(t time.Time, newLimit rate.Limit)
	Tokens() float64
	TokensAt(t time.Time) float64
	Wait(ctx context.Context) (err error)
	WaitN(ctx context.Context, n int) (err error)
}

type Reservation interface {
	Cancel()
	CancelAt(t time.Time)
	Delay() time.Duration
	DelayFrom(t time.Time) time.Duration
	OK() bool
}
