package controller

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/mocks"
)

type rateLimiterTests struct {
	suite.Suite
	limiter  *mocks.Limiter
	deduping *dedupingBucketRateLimiter
}

func TestDedupingBucketRateLimiter(t *testing.T) {
	suite.Run(t, &rateLimiterTests{})
}

func (s *rateLimiterTests) SetupTest() {
	s.limiter = mocks.NewLimiter(s.T())
	s.deduping = NewDedupingBucketRateLimiter(s.limiter).(*dedupingBucketRateLimiter)
}

func (s *rateLimiterTests) TearDownTest() {
	s.limiter.AssertExpectations(s.T())
}

func (s *rateLimiterTests) Test_When_NotFound() {
	newReservation := mocks.NewReservation(s.T())
	defer newReservation.AssertExpectations(s.T())
	newReservation.EXPECT().Delay().Return(time.Minute).Once()
	s.limiter.EXPECT().Reserve().Return(newReservation).Once()

	d := s.deduping.When("item1")

	assert.Equal(s.T(), newReservation, s.deduping.reservations["item1"])
	assert.Equal(s.T(), time.Minute, d)
}

func (s *rateLimiterTests) Test_When_FoundPast() {
	pastReservation := mocks.NewReservation(s.T())
	defer pastReservation.AssertExpectations(s.T())
	pastReservation.EXPECT().Delay().Return(-time.Minute).Once()
	s.deduping.reservations["item1"] = pastReservation
	newReservation := mocks.NewReservation(s.T())
	defer newReservation.AssertExpectations(s.T())
	newReservation.EXPECT().Delay().Return(time.Minute).Once()
	s.limiter.EXPECT().Reserve().Return(newReservation).Once()

	d := s.deduping.When("item1")

	assert.Equal(s.T(), newReservation, s.deduping.reservations["item1"])
	assert.Equal(s.T(), time.Minute, d)
}

func (s *rateLimiterTests) Test_When_FoundFuture() {
	futureReservation := mocks.NewReservation(s.T())
	defer futureReservation.AssertExpectations(s.T())
	futureReservation.EXPECT().Delay().Return(time.Minute).Twice()
	s.deduping.reservations["item1"] = futureReservation

	d := s.deduping.When("item1")

	assert.Equal(s.T(), futureReservation, s.deduping.reservations["item1"])
	assert.Equal(s.T(), time.Minute, d)
}

func (s *rateLimiterTests) Test_Forget_NotFound() {
	s.deduping.Forget("item1")

	assert.NotContains(s.T(), s.deduping.reservations, "item1")
}

func (s *rateLimiterTests) Test_Forget_PastReservation() {
	pastReservation := mocks.NewReservation(s.T())
	defer pastReservation.AssertExpectations(s.T())
	pastReservation.EXPECT().Delay().Return(-time.Minute).Once()
	s.deduping.reservations["item1"] = pastReservation

	s.deduping.Forget("item1")

	assert.NotContains(s.T(), s.deduping.reservations, "item1")
}

func (s *rateLimiterTests) Test_Forget_FutureReservation() {
	futureReservation := mocks.NewReservation(s.T())
	defer futureReservation.AssertExpectations(s.T())
	futureReservation.EXPECT().Delay().Return(time.Minute).Once()
	s.deduping.reservations["item1"] = futureReservation

	s.deduping.Forget("item1")

	assert.Equal(s.T(), futureReservation, s.deduping.reservations["item1"])
}
