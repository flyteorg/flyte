// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	models "github.com/flyteorg/flyte/cacheservice/pkg/repositories/models"

	time "time"
)

// ReservationRepo is an autogenerated mock type for the ReservationRepo type
type ReservationRepo struct {
	mock.Mock
}

type ReservationRepo_Create struct {
	*mock.Call
}

func (_m ReservationRepo_Create) Return(_a0 error) *ReservationRepo_Create {
	return &ReservationRepo_Create{Call: _m.Call.Return(_a0)}
}

func (_m *ReservationRepo) OnCreate(ctx context.Context, reservation *models.CacheReservation, now time.Time) *ReservationRepo_Create {
	c_call := _m.On("Create", ctx, reservation, now)
	return &ReservationRepo_Create{Call: c_call}
}

func (_m *ReservationRepo) OnCreateMatch(matchers ...interface{}) *ReservationRepo_Create {
	c_call := _m.On("Create", matchers...)
	return &ReservationRepo_Create{Call: c_call}
}

// Create provides a mock function with given fields: ctx, reservation, now
func (_m *ReservationRepo) Create(ctx context.Context, reservation *models.CacheReservation, now time.Time) error {
	ret := _m.Called(ctx, reservation, now)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.CacheReservation, time.Time) error); ok {
		r0 = rf(ctx, reservation, now)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type ReservationRepo_Delete struct {
	*mock.Call
}

func (_m ReservationRepo_Delete) Return(_a0 error) *ReservationRepo_Delete {
	return &ReservationRepo_Delete{Call: _m.Call.Return(_a0)}
}

func (_m *ReservationRepo) OnDelete(ctx context.Context, key string, ownerID string) *ReservationRepo_Delete {
	c_call := _m.On("Delete", ctx, key, ownerID)
	return &ReservationRepo_Delete{Call: c_call}
}

func (_m *ReservationRepo) OnDeleteMatch(matchers ...interface{}) *ReservationRepo_Delete {
	c_call := _m.On("Delete", matchers...)
	return &ReservationRepo_Delete{Call: c_call}
}

// Delete provides a mock function with given fields: ctx, key, ownerID
func (_m *ReservationRepo) Delete(ctx context.Context, key string, ownerID string) error {
	ret := _m.Called(ctx, key, ownerID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, key, ownerID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type ReservationRepo_Get struct {
	*mock.Call
}

func (_m ReservationRepo_Get) Return(_a0 *models.CacheReservation, _a1 error) *ReservationRepo_Get {
	return &ReservationRepo_Get{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *ReservationRepo) OnGet(ctx context.Context, key string) *ReservationRepo_Get {
	c_call := _m.On("Get", ctx, key)
	return &ReservationRepo_Get{Call: c_call}
}

func (_m *ReservationRepo) OnGetMatch(matchers ...interface{}) *ReservationRepo_Get {
	c_call := _m.On("Get", matchers...)
	return &ReservationRepo_Get{Call: c_call}
}

// Get provides a mock function with given fields: ctx, key
func (_m *ReservationRepo) Get(ctx context.Context, key string) (*models.CacheReservation, error) {
	ret := _m.Called(ctx, key)

	var r0 *models.CacheReservation
	if rf, ok := ret.Get(0).(func(context.Context, string) *models.CacheReservation); ok {
		r0 = rf(ctx, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.CacheReservation)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type ReservationRepo_Update struct {
	*mock.Call
}

func (_m ReservationRepo_Update) Return(_a0 error) *ReservationRepo_Update {
	return &ReservationRepo_Update{Call: _m.Call.Return(_a0)}
}

func (_m *ReservationRepo) OnUpdate(ctx context.Context, reservation *models.CacheReservation, now time.Time) *ReservationRepo_Update {
	c_call := _m.On("Update", ctx, reservation, now)
	return &ReservationRepo_Update{Call: c_call}
}

func (_m *ReservationRepo) OnUpdateMatch(matchers ...interface{}) *ReservationRepo_Update {
	c_call := _m.On("Update", matchers...)
	return &ReservationRepo_Update{Call: c_call}
}

// Update provides a mock function with given fields: ctx, reservation, now
func (_m *ReservationRepo) Update(ctx context.Context, reservation *models.CacheReservation, now time.Time) error {
	ret := _m.Called(ctx, reservation, now)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *models.CacheReservation, time.Time) error); ok {
		r0 = rf(ctx, reservation, now)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}