package backoff_manager

import (
	"context"
	"fmt"
	"github.com/lyft/flyteplugins/go/tasks/errors"
	errors3 "github.com/lyft/flytestdlib/errors"
	"github.com/lyft/flytestdlib/logger"
	"k8s.io/api/core/v1"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/clock"
	"math"
	"regexp"
	"strings"
	"time"
)

type SimpleBackOffBlocker struct {
	Clock              clock.Clock
	BackOffBaseSecond  int
	BackOffExponent    int
	NextEligibleTime   time.Time
	MaxBackOffDuration time.Duration
}

func (b *SimpleBackOffBlocker) isBlocking(t time.Time) bool {
	return !b.NextEligibleTime.Before(t)
}

func (b *SimpleBackOffBlocker) getBlockExpirationTime() time.Time {
	return b.NextEligibleTime
}

func (b *SimpleBackOffBlocker) reset() {
	b.BackOffExponent = 0
	b.NextEligibleTime = b.Clock.Now()
}

func (b *SimpleBackOffBlocker) backOff() time.Time {
	// b.backOffExponent += 1
	durationString := fmt.Sprintf("%vs", math.Pow(float64(b.BackOffBaseSecond), float64(b.BackOffExponent)))
	backOffDuration, _ := time.ParseDuration(durationString)

	if backOffDuration > b.MaxBackOffDuration {
		backOffDuration = b.MaxBackOffDuration
	}

	b.NextEligibleTime = b.Clock.Now().Add(backOffDuration)
	b.BackOffExponent += 1
	return b.NextEligibleTime
}

type ComputeResourceCeilings struct {
	computeResourceCeilings v1.ResourceList
}

func (r *ComputeResourceCeilings) isEligible(requestedResourceList v1.ResourceList) bool {
	eligibility := true
	for reqResource, reqQuantity := range requestedResourceList {
		eligibility = eligibility && (reqQuantity.Cmp(r.computeResourceCeilings[reqResource]) == -1)
	}
	return eligibility
}

func (r *ComputeResourceCeilings) update(reqResource v1.ResourceName, reqQuantity resource.Quantity) {

	if currentCeiling, ok := r.computeResourceCeilings[reqResource]; !ok || reqQuantity.Value() < currentCeiling.Value() {
		r.computeResourceCeilings[reqResource] = reqQuantity.DeepCopy()
	}
}

func (r *ComputeResourceCeilings) updateAll(resources *v1.ResourceList) {
	for reqResource, reqQuantity := range *resources {
		r.update(reqResource, reqQuantity)
	}
}

func (r *ComputeResourceCeilings) reset(resource v1.ResourceName) {
	r.computeResourceCeilings[resource] = r.inf()
}

func (r *ComputeResourceCeilings) resetAll() {
	for compResource := range r.computeResourceCeilings {
		r.reset(compResource)
	}
}

func (r *ComputeResourceCeilings) inf() resource.Quantity {
	// A hack to represent RESOURCE_MAX
	return resource.MustParse("1Ei")
}

type ComputeResourceAwareBackOffHandler struct {
	*SimpleBackOffBlocker
	*ComputeResourceCeilings
}

func (h *ComputeResourceAwareBackOffHandler) IsActive() bool {
	return h.BackOffBaseSecond == 0
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
	if !h.IsActive() {
		return operation()
	} else if !h.SimpleBackOffBlocker.isBlocking(now) || h.ComputeResourceCeilings.isEligible(requestedResourceList) {
		err := operation()
		if err != nil {
			if IsResourceQuotaExceeded(err) {
				// It is necessary to parse the error message to get the actual constraints
				// in this case, if the error message indicates constraints on memory only, then we shouldn't be used to lower the CPU ceiling
				// even if CPU appears in requestedResourceList
				newCeiling := GetComputeResourceAndQuantityRequested(err)
				h.ComputeResourceCeilings.updateAll(&newCeiling)

				if !h.SimpleBackOffBlocker.isBlocking(now) {
					// if the backOffBlocker is not blocking and we are still encountering insufficient resource issue,
					// we should increase the exponent in the backoff and update the NextEligibleTime

					newEligibleTime := h.SimpleBackOffBlocker.backOff()
					logger.Errorf(ctx, "BackOffBlocker is not blocking, but the request failed due to "+
						"insufficient resource (backing off further to [%v]): [%v]\n", newEligibleTime, err)
				} else {
					// When lowering the ceiling, we only want to lower the ceiling that actually needs to be lowered.
					// For example, if the creation of a pod requiring X cpus and Y memory got rejected because of
					// 	insufficient memory, we should only lower the ceiling of memory to Y, without touching the cpu ceiling

					logger.Errorf(ctx, "BackOffBlocker is active but the requested amount of resource is lower "+
						"than ceiling so is not blocked. "+
						"Failed to run the operation due to insufficient resource "+
						"(next eligible time remains unchanged [%v]): [%v]\n", h.SimpleBackOffBlocker.NextEligibleTime, err)
				}
			} else {
				logger.Errorf(ctx, "Failed to run the operation due to reasons other than insufficient resource: [%v]\n", err)
				return err
			}
			return errors.Wrapf(errors.BackOffError, err, "Failed to execute the operation")
		} else {
			h.SimpleBackOffBlocker.reset()
			h.ComputeResourceCeilings.resetAll()
			return nil
		}
	} else { // The backoff is active and the resource request exceeds the ceiling
		logger.Errorf(ctx, "Failed to execute the operation due to backoff")
		return errors.Errorf(errors.BackOffError, "Failed to execute the operation due to backoff is "+
			"active [attempted at: %v][block expires at: %v] and the requested "+
			"resource(s) exceeds resource ceiling(s)", now, h.SimpleBackOffBlocker.getBlockExpirationTime())
	}

}

func IsResourceQuotaExceeded(err error) bool {
	return errors2.IsForbidden(err) && strings.Contains(err.Error(), "exceeded quota")
}

func GetComputeResourceAndQuantityRequested(err error) v1.ResourceList {
	// re := regexp.MustCompile(`(?P<part>(?P<key>requested|used|limited): limits.(?P<resource_type>[a-zA-Z]+)=(?P<quantity_expr>[a-zA-Z0-9]+))`)
	// Playground: https://play.golang.org/p/oOr6CMmW7IE
	// limits.cpu=7,limits.memory=64Gi, used: limits.cpu=249,limits.memory=2012730Mi, limited: limits.cpu=250,limits.memory=2000Gi
	// re := regexp.MustCompile(`(?P<key>requested): (limits.(?P<resource_type>[a-zA-Z]+)=(?P<quantity_expr>[a-zA-Z0-9]+)[,]*)+`)
	// re := regexp.MustCompile(`(?P<key>requested): (limits.(?P<resource_type>[a-zA-Z]+)=(?P<quantity_expr>[a-zA-Z0-9]+)[,]*)+`)

	// Extracting "requested: limits.cpu=7,limits.memory=64Gi"
	re := regexp.MustCompile(`requested: (limits.[a-zA-Z]+=[a-zA-Z0-9]+[,]*)+`)
	matches := re.FindAllStringSubmatch(err.Error(), -1)
	requestedComputeResources := v1.ResourceList{}

	if len(matches) == 0 || len(matches[0]) == 0 {
		return requestedComputeResources
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
		requestedComputeResources[v1.ResourceName(tuple[0])] = resource.MustParse(tuple[1])
	}
	return requestedComputeResources
}

// TODO ssingh: clean it and move to its right place
func IsBackoffError(err error) bool {
	code, found := errors3.GetErrorCode(err)
	if found && code == errors.BackOffError {
		return true
	}

	return false
}
