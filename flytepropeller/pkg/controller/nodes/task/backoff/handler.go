package backoff

import (
	"context"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	stdAtomic "github.com/flyteorg/flytestdlib/atomic"
	"github.com/flyteorg/flytestdlib/logger"
	v1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/clock"
)

var (
	limitedLimitsRegexp     = regexp.MustCompile(`limited: (limits.[a-zA-Z]+=[a-zA-Z0-9]+[,]*)+`)
	limitedRequestsRegexp   = regexp.MustCompile(`limited: (requests.[a-zA-Z]+=[a-zA-Z0-9]+[,]*)+`)
	requestedLimitsRegexp   = regexp.MustCompile(`requested: (limits.[a-zA-Z]+=[a-zA-Z0-9]+[,]*)+`)
	requestedRequestsRegexp = regexp.MustCompile(`requested: (requests.[a-zA-Z]+=[a-zA-Z0-9]+[,]*)+`)
)

// SimpleBackOffBlocker is a simple exponential back-off timer that keeps track of the back-off period
type SimpleBackOffBlocker struct {
	Clock              clock.Clock
	BackOffBaseSecond  int
	MaxBackOffDuration time.Duration

	// Mutable fields
	BackOffExponent  stdAtomic.Uint32
	NextEligibleTime AtomicTime
}

func (b *SimpleBackOffBlocker) isBlocking(t time.Time) bool {
	return !b.NextEligibleTime.Load().Before(t)
}

func (b *SimpleBackOffBlocker) getBlockExpirationTime() time.Time {
	return b.NextEligibleTime.Load()
}

func (b *SimpleBackOffBlocker) reset() {
	b.BackOffExponent.Store(0)
	b.NextEligibleTime.Store(b.Clock.Now())
}

func (b *SimpleBackOffBlocker) backOff(ctx context.Context) time.Duration {
	logger.Debug(ctx, "BackOff params [BackOffBaseSecond: %v] [BackOffExponent: %v] [MaxBackOffDuration: %v]",
		b.BackOffBaseSecond, b.BackOffExponent, b.MaxBackOffDuration)

	backOffDuration := time.Duration(time.Second.Nanoseconds() * int64(math.Pow(float64(b.BackOffBaseSecond),
		float64(b.BackOffExponent.Load()))))

	if backOffDuration > b.MaxBackOffDuration {
		backOffDuration = b.MaxBackOffDuration
	} else {
		b.BackOffExponent.Inc()
	}

	b.NextEligibleTime.Store(b.Clock.Now().Add(backOffDuration))
	return backOffDuration
}

type ComputeResourceCeilings struct {
	computeResourceCeilings *SyncResourceList
}

func (r *ComputeResourceCeilings) isEligible(requestedResourceList v1.ResourceList) (eligibility bool) {
	for reqResource, reqQuantity := range requestedResourceList {
		val, found := r.computeResourceCeilings.Load(reqResource)
		if found && reqQuantity.Cmp(val) >= 0 {
			return false
		}
	}

	return true
}

func (r *ComputeResourceCeilings) update(reqResource v1.ResourceName, reqQuantity resource.Quantity) {
	if currentCeiling, ok := r.computeResourceCeilings.Load(reqResource); !ok || reqQuantity.Value() < currentCeiling.Value() {
		r.computeResourceCeilings.Store(reqResource, reqQuantity.DeepCopy())
	}
}

func (r *ComputeResourceCeilings) updateAll(resources *v1.ResourceList) {
	for reqResource, reqQuantity := range *resources {
		r.update(reqResource, reqQuantity)
	}
}

func (r *ComputeResourceCeilings) reset(resource v1.ResourceName) {
	r.computeResourceCeilings.Store(resource, r.inf())
}

func (r *ComputeResourceCeilings) resetAll() {
	r.computeResourceCeilings.Range(func(key v1.ResourceName, value resource.Quantity) bool {
		r.reset(key)
		return true
	})
}

func (r *ComputeResourceCeilings) inf() resource.Quantity {
	// A hack to represent RESOURCE_MAX
	return resource.MustParse("1Ei")
}

// ComputeResourceAwareBackOffHandler is an exponential back-off handler that also keeps track of the resource ceilings
// of the operations that are blocked or failed due to resource insufficiency
type ComputeResourceAwareBackOffHandler struct {
	*SimpleBackOffBlocker
	*ComputeResourceCeilings
}

func (h *ComputeResourceAwareBackOffHandler) IsActive() bool {
	return h.BackOffBaseSecond != 0
}

func (h *ComputeResourceAwareBackOffHandler) reset() {
	h.SimpleBackOffBlocker.reset()
	h.ComputeResourceCeilings.resetAll()
}

// Act based on current backoff interval and set the next one accordingly
func (h *ComputeResourceAwareBackOffHandler) Handle(ctx context.Context, operation func() error, requestedResourceList v1.ResourceList) error {

	// Pseudo code:
	// If the backoff is inactive => we should just go ahead and execute the operation(), and handle the error properly
	//		If operation() fails because of resource => lower the ceiling
	//		Else we return whatever the result is
	//
	// Else if the backoff is active => we should reduce the number of calls to the API server in this case
	//		If resource is lower than the ceiling => We should try the operation().
	//			If operation() fails because of the lack of resource, we will lower the ceiling
	//          Else we return whatever the operation() returns
	//      Else => we block the operation(), which is where the main improvement comes from

	now := h.Clock.Now()
	isBlocking := h.SimpleBackOffBlocker.isBlocking(now)
	isTryable := h.ComputeResourceCeilings.isEligible(requestedResourceList)
	if !h.IsActive() {
		return operation()
	} else if !isBlocking || isTryable {
		err := operation()
		if err == nil {
			logger.Infof(ctx, "The operation was attempted and finished without an error\n")
			h.reset()
			return nil
		}

		if IsBackOffError(err) {
			if !isBlocking {
				// if the backOffBlocker is not blocking and we are still encountering insufficient resource issue,
				// we should increase the exponent in the backoff and update the NextEligibleTime

				backOffDuration := h.SimpleBackOffBlocker.backOff(ctx)
				logger.Infof(ctx, "The operation was attempted because the back-off handler is not blocking, but failed due to "+
					"%s (backing off for a duration of [%v] to timestamp [%v])\n",
					err, backOffDuration, h.SimpleBackOffBlocker.NextEligibleTime)
			} else {
				// When lowering the ceiling, we only want to lower the ceiling that actually needs to be lowered.
				// For example, if the creation of a pod requiring X cpus and Y memory got rejected because of
				// 	insufficient memory, we should only lower the ceiling of memory to Y, without touching the cpu ceiling

				logger.Infof(ctx, "The operation was attempted because the resource requested is lower than the ceilings, "+
					"but failed due to %s (the next eligible time "+
					"remains unchanged [%v]). The requests are [%v]. The ceilings are [%v]\n",
					err, h.SimpleBackOffBlocker.NextEligibleTime, requestedResourceList, h.computeResourceCeilings.String())
			}
			if IsResourceQuotaExceeded(err) {
				// It is necessary to parse the error message to get the actual constraints
				// in this case, if the error message indicates constraints on memory only, then we shouldn't be used to lower the CPU ceiling
				// even if CPU appears in requestedResourceList
				newCeiling := GetComputeResourceAndQuantity(err, requestedLimitsRegexp)
				h.ComputeResourceCeilings.updateAll(&newCeiling)
			}

			return errors.Wrapf(errors.BackOffError, err, "The operation was attempted but failed")
		}
		logger.Infof(ctx, "The operation was attempted but failed due to reason(s) other than insufficient resource: [%v]\n", err)
		return err
	} else { // The backoff is active and the resource request exceeds the ceiling
		logger.Infof(ctx, "The operation was blocked due to back-off")
		return errors.Errorf(errors.BackOffError, "The operation attempt was blocked by back-off "+
			"[attempted at: %v][the block expires at: %v] and the requested "+
			"resource(s) exceeds resource ceiling(s)", now, h.SimpleBackOffBlocker.getBlockExpirationTime())
	}

}

func IsResourceQuotaExceeded(err error) bool {
	return apiErrors.IsForbidden(err) && strings.Contains(err.Error(), "exceeded quota")
}

func IsBackOffError(err error) bool {
	return IsResourceQuotaExceeded(err) || apiErrors.IsTooManyRequests(err) || apiErrors.IsServerTimeout(err)
}

func GetComputeResourceAndQuantity(err error, resourceRegex *regexp.Regexp) v1.ResourceList {
	// Playground: https://play.golang.org/p/oOr6CMmW7IE

	// Sample message:
	// "requested: limits.cpu=7,limits.memory=64Gi, used: limits.cpu=249,limits.memory=2012730Mi, limited: limits.cpu=250,limits.memory=2000Gi"

	// Extracting "requested: limits.cpu=7,limits.memory=64Gi"
	matches := resourceRegex.FindAllStringSubmatch(err.Error(), -1)
	computeResources := v1.ResourceList{}

	if len(matches) == 0 || len(matches[0]) == 0 {
		return computeResources
	}

	// Extracting "limits.cpu=7,limits.memory=64Gi"
	descr := strings.SplitN(matches[0][0], ":", 2)

	// Extracting "limits.cpu=7","limits.memory=64Gi"
	chunks := strings.SplitN(descr[1], ",", -1)
	for _, c := range chunks {
		// Extracting "cpu=7","memory=64Gi"
		resrcString := strings.SplitN(c, ".", 2)
		if len(resrcString) < 2 {
			continue
		}
		// Extracting ["cpu","7"], ["memory","64Gi"]
		tuple := strings.SplitN(resrcString[1], "=", 2)
		if len(tuple) < 2 {
			continue
		}
		computeResources[v1.ResourceName(tuple[0])] = resource.MustParse(tuple[1])
	}
	return computeResources
}

func IsResourceRequestsEligible(err error) bool {
	limitedLimitsResourceList := GetComputeResourceAndQuantity(err, limitedLimitsRegexp)
	limitedRequestsResourceList := GetComputeResourceAndQuantity(err, limitedRequestsRegexp)
	requestedLimitsResourceList := GetComputeResourceAndQuantity(err, requestedLimitsRegexp)
	requestedRequestsResourceList := GetComputeResourceAndQuantity(err, requestedRequestsRegexp)

	return isEligible(requestedLimitsResourceList, limitedLimitsResourceList) &&
		isEligible(requestedRequestsResourceList, limitedRequestsResourceList)
}

func isEligible(requestedResourceList, quotaResourceList v1.ResourceList) (eligibility bool) {
	for resource, requestedQuantity := range requestedResourceList {
		quotaQuantity, exists := quotaResourceList[resource]
		if exists && requestedQuantity.Cmp(quotaQuantity) >= 0 {
			return false
		}
	}

	return true
}
