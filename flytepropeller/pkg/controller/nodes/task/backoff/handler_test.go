package backoff

import (
	"context"
	"errors"
	"reflect"
	"regexp"
	"testing"
	"time"

	stdAtomic "github.com/flyteorg/flytestdlib/atomic"

	"github.com/stretchr/testify/assert"

	taskErrors "github.com/flyteorg/flyteplugins/go/tasks/errors"
	stdlibErrors "github.com/flyteorg/flytestdlib/errors"
	v1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/clock"
)

func TestComputeResourceAwareBackOffHandler_Handle(t *testing.T) {
	var callCount = 0

	operWithErr := func() error {
		callCount++
		return apiErrors.NewForbidden(
			schema.GroupResource{}, "", errors.New("is forbidden: "+
				"exceeded quota: project-quota, requested: limits.memory=1Gi, "+
				"used: limits.memory=7976Gi, limited: limits.memory=8000Gi"))
	}

	operWithNoErr := func() error {
		callCount++
		return nil
	}

	ctx := context.TODO()
	tc := clock.NewFakeClock(time.Now())
	type fields struct {
		SimpleBackOffBlocker    *SimpleBackOffBlocker
		ComputeResourceCeilings *ComputeResourceCeilings
	}
	type args struct {
		operation             func() error
		requestedResourceList v1.ResourceList
	}
	tests := []struct {
		name                 string
		fields               fields
		args                 args
		wantErr              bool
		wantErrCode          stdlibErrors.ErrorCode
		wantExp              uint32
		wantNextEligibleTime time.Time
		wantCeilings         v1.ResourceList
		wantCallCount        int
	}{

		{name: "Request gte the ceilings coming in before the block expires does not call operation()",
			fields: fields{
				SimpleBackOffBlocker: &SimpleBackOffBlocker{
					Clock:              tc,
					BackOffBaseSecond:  2,
					BackOffExponent:    stdAtomic.NewUint32(1),
					NextEligibleTime:   NewAtomicTime(tc.Now().Add(time.Second * 7)),
					MaxBackOffDuration: 10 * time.Second,
				},
				ComputeResourceCeilings: &ComputeResourceCeilings{
					computeResourceCeilings: SyncResourceListFromResourceList(v1.ResourceList{v1.ResourceCPU: resource.MustParse("10"), v1.ResourceMemory: resource.MustParse("64Gi")}),
				},
			},
			args: args{
				operation:             operWithNoErr,
				requestedResourceList: v1.ResourceList{v1.ResourceCPU: resource.MustParse("10"), v1.ResourceMemory: resource.MustParse("64Gi")},
			},
			wantErr:              true,
			wantErrCode:          taskErrors.BackOffError,
			wantExp:              1,
			wantNextEligibleTime: tc.Now().Add(time.Second * 7),
			wantCeilings:         v1.ResourceList{v1.ResourceCPU: resource.MustParse("10"), v1.ResourceMemory: resource.MustParse("64Gi")},
			wantCallCount:        0,
		},

		{name: "Smaller request before the block expires that gets rejected does not lead to further any back off",
			fields: fields{
				SimpleBackOffBlocker: &SimpleBackOffBlocker{
					Clock:              tc,
					BackOffBaseSecond:  2,
					BackOffExponent:    stdAtomic.NewUint32(1),
					NextEligibleTime:   NewAtomicTime(tc.Now().Add(time.Second * 7)),
					MaxBackOffDuration: 10 * time.Second,
				},
				ComputeResourceCeilings: &ComputeResourceCeilings{
					computeResourceCeilings: SyncResourceListFromResourceList(v1.ResourceList{v1.ResourceCPU: resource.MustParse("10"), v1.ResourceMemory: resource.MustParse("64Gi")}),
				},
			},
			args: args{
				operation:             operWithErr,
				requestedResourceList: v1.ResourceList{v1.ResourceCPU: resource.MustParse("9"), v1.ResourceMemory: resource.MustParse("1Gi")}},
			wantErr:              true,
			wantErrCode:          taskErrors.BackOffError,
			wantExp:              1,
			wantNextEligibleTime: tc.Now().Add(time.Second * 7),
			// The following resource quantity of CPU is correct, as the error message will show the actual resource constraint
			// in this case, the error message indicates constraints on memory only, so while requestedResourceList has a CPU component,
			// it shouldn't be used to lower the ceiling
			wantCeilings:  v1.ResourceList{v1.ResourceCPU: resource.MustParse("10"), v1.ResourceMemory: resource.MustParse("1Gi")},
			wantCallCount: 1,
		},

		{name: "Request coming after the block expires, if rejected, should lead to further backoff and proper ceilings",
			fields: fields{
				SimpleBackOffBlocker: &SimpleBackOffBlocker{
					Clock:              tc,
					BackOffBaseSecond:  2,
					BackOffExponent:    stdAtomic.NewUint32(1),
					NextEligibleTime:   NewAtomicTime(tc.Now().Add(time.Second * -2)),
					MaxBackOffDuration: 10 * time.Second,
				},
				ComputeResourceCeilings: &ComputeResourceCeilings{
					computeResourceCeilings: SyncResourceListFromResourceList(v1.ResourceList{v1.ResourceCPU: resource.MustParse("1Ei"), v1.ResourceMemory: resource.MustParse("1Ei")}),
				},
			},
			args: args{
				operation:             operWithErr,
				requestedResourceList: v1.ResourceList{v1.ResourceMemory: resource.MustParse("1Gi")}},
			wantErr:              true,
			wantErrCode:          taskErrors.BackOffError,
			wantExp:              2,
			wantNextEligibleTime: tc.Now().Add(time.Second * 2),
			wantCeilings:         v1.ResourceList{v1.ResourceCPU: resource.MustParse("1Ei"), v1.ResourceMemory: resource.MustParse("1Gi")},
			wantCallCount:        1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			callCount = 0
			h := &ComputeResourceAwareBackOffHandler{
				SimpleBackOffBlocker:    tt.fields.SimpleBackOffBlocker,
				ComputeResourceCeilings: tt.fields.ComputeResourceCeilings,
			}
			err := h.Handle(ctx, tt.args.operation, tt.args.requestedResourceList)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				if ec, found := stdlibErrors.GetErrorCode(err); found && ec != tt.wantErrCode {
					t.Errorf("Handle() errorCode = %v, wantErrCode %v", ec, tt.wantErrCode)
				}
			}
			if tt.wantExp != h.BackOffExponent.Load() {
				t.Errorf("post-Handle() BackOffExponent = %v, wantBackOffExponent %v", h.BackOffExponent, tt.wantExp)
			}
			if tt.wantNextEligibleTime != h.NextEligibleTime.Load() {
				t.Errorf("post-Handle() NextEligibleTime = %v, wantNextEligibleTime %v", h.NextEligibleTime, tt.wantNextEligibleTime)
			}
			if !reflect.DeepEqual(h.computeResourceCeilings.AsResourceList(), tt.wantCeilings) {
				t.Errorf("ResourceCeilings = %v, want %v", h.computeResourceCeilings.AsResourceList(), tt.wantCeilings)
			}
			if tt.wantCallCount != callCount {
				t.Errorf("Operation call count = %v, want %v", callCount, tt.wantCallCount)
			}
		})
	}
}

func TestComputeResourceCeilings_inf(t *testing.T) {
	type fields struct {
		v1.ResourceList
	}
	tests := []struct {
		name   string
		fields fields
		want   resource.Quantity
	}{
		{name: "inf is set properly", fields: fields{ResourceList: v1.ResourceList{}}, want: resource.MustParse("1Ei")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ComputeResourceCeilings{
				computeResourceCeilings: SyncResourceListFromResourceList(tt.fields.ResourceList),
			}
			if got := r.inf(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("inf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComputeResourceCeilings_isEligible(t *testing.T) {
	type fields struct {
		computeResourceCeilings v1.ResourceList
	}
	type args struct {
		requestedResourceList v1.ResourceList
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{name: "Smaller request should be accepted",
			fields: fields{computeResourceCeilings: v1.ResourceList{v1.ResourceCPU: resource.MustParse("0.7m")}},
			args:   args{requestedResourceList: v1.ResourceList{v1.ResourceCPU: resource.MustParse("0.5m")}},
			want:   true},
		{name: "Larger request should be rejected",
			fields: fields{computeResourceCeilings: v1.ResourceList{v1.ResourceMemory: resource.MustParse("32Gi")}},
			args:   args{requestedResourceList: v1.ResourceList{v1.ResourceMemory: resource.MustParse("33Gi")}},
			want:   false},
		{name: "If not all of the resource requests are smaller, it should be rejected",
			fields: fields{computeResourceCeilings: v1.ResourceList{v1.ResourceCPU: resource.MustParse("0.7m"), v1.ResourceMemory: resource.MustParse("32Gi")}},
			args:   args{requestedResourceList: v1.ResourceList{v1.ResourceCPU: resource.MustParse("0.69m"), v1.ResourceMemory: resource.MustParse("33Gi")}},
			want:   false},
		{name: "All the resources should be strictly smaller (<); otherwise it is rejected",
			fields: fields{computeResourceCeilings: v1.ResourceList{v1.ResourceCPU: resource.MustParse("0.7m"), v1.ResourceMemory: resource.MustParse("32Gi")}},
			args:   args{requestedResourceList: v1.ResourceList{v1.ResourceCPU: resource.MustParse("0.7m"), v1.ResourceMemory: resource.MustParse("32Gi")}},
			want:   false},
		{name: "If all the resource requests are smaller, it should be accepted",
			fields: fields{computeResourceCeilings: v1.ResourceList{v1.ResourceCPU: resource.MustParse("0.8m"), v1.ResourceMemory: resource.MustParse("34Gi")}},
			args:   args{requestedResourceList: v1.ResourceList{v1.ResourceCPU: resource.MustParse("0.7m"), v1.ResourceMemory: resource.MustParse("32Gi")}},
			want:   true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ComputeResourceCeilings{
				computeResourceCeilings: SyncResourceListFromResourceList(tt.fields.computeResourceCeilings),
			}
			if got := r.isEligible(tt.args.requestedResourceList); got != tt.want {
				t.Errorf("isEligible() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComputeResourceCeilings_update(t *testing.T) {
	type fields struct {
		computeResourceCeilings v1.ResourceList
	}
	type args struct {
		reqResource v1.ResourceName
		reqQuantity resource.Quantity
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wants          []args
		shouldNotExist []v1.ResourceName
	}{
		{name: "Update CPU ceiling on top of an existing ceiling",
			fields:         fields{computeResourceCeilings: v1.ResourceList{v1.ResourceCPU: resource.MustParse("1"), v1.ResourceMemory: resource.MustParse("1Gi")}},
			args:           args{reqResource: v1.ResourceCPU, reqQuantity: resource.MustParse("0")},
			wants:          []args{{reqResource: v1.ResourceCPU, reqQuantity: resource.MustParse("0")}, {reqResource: v1.ResourceMemory, reqQuantity: resource.MustParse("1Gi")}},
			shouldNotExist: []v1.ResourceName{}},
		{name: "Update Memory ceiling on top of an existing ceiling",
			fields:         fields{computeResourceCeilings: v1.ResourceList{v1.ResourceCPU: resource.MustParse("1"), v1.ResourceMemory: resource.MustParse("1Gi")}},
			args:           args{reqResource: v1.ResourceMemory, reqQuantity: resource.MustParse("500Mi")},
			wants:          []args{{reqResource: v1.ResourceCPU, reqQuantity: resource.MustParse("1")}, {reqResource: v1.ResourceMemory, reqQuantity: resource.MustParse("500Mi")}},
			shouldNotExist: []v1.ResourceName{}},
		{name: "Update CPU ceiling without an existing ceiling",
			fields:         fields{computeResourceCeilings: v1.ResourceList{}},
			args:           args{reqResource: v1.ResourceCPU, reqQuantity: resource.MustParse("1")},
			wants:          []args{{reqResource: v1.ResourceCPU, reqQuantity: resource.MustParse("1")}},
			shouldNotExist: []v1.ResourceName{v1.ResourceMemory}},
		{name: "Update Memory ceiling without an existing ceiling",
			fields:         fields{computeResourceCeilings: v1.ResourceList{}},
			args:           args{reqResource: v1.ResourceMemory, reqQuantity: resource.MustParse("500Mi")},
			wants:          []args{{reqResource: v1.ResourceMemory, reqQuantity: resource.MustParse("500Mi")}},
			shouldNotExist: []v1.ResourceName{v1.ResourceCPU}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ComputeResourceCeilings{
				computeResourceCeilings: SyncResourceListFromResourceList(tt.fields.computeResourceCeilings),
			}
			r.update(tt.args.reqResource, tt.args.reqQuantity)
			for _, rsrcTuple := range tt.wants {
				if quantity, found := r.computeResourceCeilings.Load(rsrcTuple.reqResource); !found || rsrcTuple.reqQuantity != quantity {
					t.Errorf("Resource ceiling: (%v, %v), want (%v, %v)",
						rsrcTuple.reqResource, quantity, rsrcTuple.reqResource, rsrcTuple.reqQuantity)
				}
			}
			for _, ne := range tt.shouldNotExist {
				if _, ok := r.computeResourceCeilings.Load(ne); ok {
					t.Errorf("Resource should not exist in ceiling: %v", ne)
				}
			}
		})
	}
}

func TestComputeResourceCeilings_updateAll(t *testing.T) {
	type fields struct {
		computeResourceCeilings v1.ResourceList
	}
	type args struct {
		resources *v1.ResourceList
	}
	type wants struct {
		resources map[v1.ResourceName]resource.Quantity
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wants          wants
		shouldNotExist []v1.ResourceName
	}{
		{name: "Update Memory ceiling without an existing ceiling",
			fields:         fields{computeResourceCeilings: v1.ResourceList{}},
			args:           args{&v1.ResourceList{v1.ResourceMemory: resource.MustParse("500Mi")}},
			wants:          wants{map[v1.ResourceName]resource.Quantity{v1.ResourceMemory: resource.MustParse("500Mi")}},
			shouldNotExist: []v1.ResourceName{v1.ResourceCPU}},
		{name: "Update Memory ceiling without an existing ceiling",
			fields:         fields{computeResourceCeilings: v1.ResourceList{v1.ResourceCPU: resource.MustParse("1"), v1.ResourceMemory: resource.MustParse("1Gi")}},
			args:           args{&v1.ResourceList{v1.ResourceMemory: resource.MustParse("500Mi")}},
			wants:          wants{map[v1.ResourceName]resource.Quantity{v1.ResourceCPU: resource.MustParse("1"), v1.ResourceMemory: resource.MustParse("500Mi")}},
			shouldNotExist: []v1.ResourceName{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ComputeResourceCeilings{
				computeResourceCeilings: SyncResourceListFromResourceList(tt.fields.computeResourceCeilings),
			}
			r.updateAll(tt.args.resources)
			for rs, qs := range tt.wants.resources {
				if quantity, found := r.computeResourceCeilings.Load(rs); !found || quantity != qs {
					t.Errorf("Resource ceiling: (%v, %v), want (%v, %v)",
						rs, quantity, rs, qs)
				}
			}
			for _, ne := range tt.shouldNotExist {
				if _, ok := r.computeResourceCeilings.Load(ne); ok {
					t.Errorf("Resource should not exist in ceiling: %v", ne)
				}
			}
		})
	}
}

func TestGetComputeResourceAndQuantityRequested(t *testing.T) {
	type args struct {
		err    error
		regexp *regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want v1.ResourceList
	}{
		{name: "Limited memory limits", args: args{err: apiErrors.NewForbidden(
			schema.GroupResource{}, "", errors.New("is forbidden: "+
				"exceeded quota: project-quota, requested: limits.memory=3Gi, used: limits.memory=7976Gi, limited: limits.memory=8000Gi")),
			regexp: limitedLimitsRegexp},
			want: v1.ResourceList{v1.ResourceMemory: resource.MustParse("8000Gi")}},
		{name: "Limited CPU limits", args: args{err: apiErrors.NewForbidden(
			schema.GroupResource{}, "", errors.New("is forbidden: "+
				"exceeded quota: project-quota, requested: limits.cpu=3640m, used: limits.cpu=6000m, limited: limits.cpu=8000m")),
			regexp: limitedLimitsRegexp},
			want: v1.ResourceList{v1.ResourceCPU: resource.MustParse("8000m")}},
		{name: "Limited multiple limits ", args: args{err: apiErrors.NewForbidden(
			schema.GroupResource{}, "", errors.New("is forbidden: "+
				"exceeded quota: project-quota, requested: limits.cpu=7,limits.memory=64Gi, used: limits.cpu=249,limits.memory=2012730Mi, limited: limits.cpu=250,limits.memory=2000Gi")),
			regexp: limitedLimitsRegexp},
			want: v1.ResourceList{v1.ResourceCPU: resource.MustParse("250"), v1.ResourceMemory: resource.MustParse("2000Gi")}},
		{name: "Limited memory requests", args: args{err: apiErrors.NewForbidden(
			schema.GroupResource{}, "", errors.New("is forbidden: "+
				"exceeded quota: project-quota, requested: requests.memory=3Gi, used: requests.memory=7976Gi, limited: requests.memory=8000Gi")),
			regexp: limitedRequestsRegexp},
			want: v1.ResourceList{v1.ResourceMemory: resource.MustParse("8000Gi")}},
		{name: "Limited CPU requests", args: args{err: apiErrors.NewForbidden(
			schema.GroupResource{}, "", errors.New("is forbidden: "+
				"exceeded quota: project-quota, requested: requests.cpu=3640m, used: requests.cpu=6000m, limited: requests.cpu=8000m")),
			regexp: limitedRequestsRegexp},
			want: v1.ResourceList{v1.ResourceCPU: resource.MustParse("8000m")}},
		{name: "Limited multiple requests ", args: args{err: apiErrors.NewForbidden(
			schema.GroupResource{}, "", errors.New("is forbidden: "+
				"exceeded quota: project-quota, requested: requests.cpu=7,requests.memory=64Gi, used: requests.cpu=249,requests.memory=2012730Mi, limited: requests.cpu=250,requests.memory=2000Gi")),
			regexp: limitedRequestsRegexp},
			want: v1.ResourceList{v1.ResourceCPU: resource.MustParse("250"), v1.ResourceMemory: resource.MustParse("2000Gi")}},
		{name: "Requested memory limits", args: args{err: apiErrors.NewForbidden(
			schema.GroupResource{}, "", errors.New("is forbidden: "+
				"exceeded quota: project-quota, requested: limits.memory=3Gi, used: limits.memory=7976Gi, limited: limits.memory=8000Gi")),
			regexp: requestedLimitsRegexp},
			want: v1.ResourceList{v1.ResourceMemory: resource.MustParse("3Gi")}},
		{name: "Requested CPU limits", args: args{err: apiErrors.NewForbidden(
			schema.GroupResource{}, "", errors.New("is forbidden: "+
				"exceeded quota: project-quota, requested: limits.cpu=3640m, used: limits.cpu=6000m, limited: limits.cpu=8000m")),
			regexp: requestedLimitsRegexp},
			want: v1.ResourceList{v1.ResourceCPU: resource.MustParse("3640m")}},
		{name: "Requested multiple limits ", args: args{err: apiErrors.NewForbidden(
			schema.GroupResource{}, "", errors.New("is forbidden: "+
				"exceeded quota: project-quota, requested: limits.cpu=7,limits.memory=64Gi, used: limits.cpu=249,limits.memory=2012730Mi, limited: limits.cpu=250,limits.memory=2000Gi")),
			regexp: requestedLimitsRegexp},
			want: v1.ResourceList{v1.ResourceCPU: resource.MustParse("7"), v1.ResourceMemory: resource.MustParse("64Gi")}},
		{name: "Requested memory requests", args: args{err: apiErrors.NewForbidden(
			schema.GroupResource{}, "", errors.New("is forbidden: "+
				"exceeded quota: project-quota, requested: requests.memory=3Gi, used: requests.memory=7976Gi, limited: requests.memory=8000Gi")),
			regexp: requestedRequestsRegexp},
			want: v1.ResourceList{v1.ResourceMemory: resource.MustParse("3Gi")}},
		{name: "Requested CPU requests", args: args{err: apiErrors.NewForbidden(
			schema.GroupResource{}, "", errors.New("is forbidden: "+
				"exceeded quota: project-quota, requested: requests.cpu=3640m, used: requests.cpu=6000m, limited: requests.cpu=8000m")),
			regexp: requestedRequestsRegexp},
			want: v1.ResourceList{v1.ResourceCPU: resource.MustParse("3640m")}},
		{name: "Requested multiple requests ", args: args{err: apiErrors.NewForbidden(
			schema.GroupResource{}, "", errors.New("is forbidden: "+
				"exceeded quota: project-quota, requested: requests.cpu=7,requests.memory=64Gi, used: requests.cpu=249,requests.memory=2012730Mi, limited: requests.cpu=250,requests.memory=2000Gi")),
			regexp: requestedRequestsRegexp},
			want: v1.ResourceList{v1.ResourceCPU: resource.MustParse("7"), v1.ResourceMemory: resource.MustParse("64Gi")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetComputeResourceAndQuantity(tt.args.err, tt.args.regexp); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetComputeResourceAndQuantity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsBackoffError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "True BackOffError", args: args{err: taskErrors.Errorf(taskErrors.BackOffError, "")}, want: true},
		{name: "BadTaskSpecification error", args: args{err: taskErrors.Errorf(taskErrors.BadTaskSpecification, "")}, want: false},
		{name: "TaskFailedUnknownError", args: args{err: taskErrors.Errorf(taskErrors.TaskFailedUnknownError, "")}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stdlibErrors.IsCausedBy(tt.args.err, taskErrors.BackOffError); got != tt.want {
				t.Errorf("IsBackoffError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsResourceQuotaExceeded(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "True ResourceQuotaExceeded error msg", args: args{err: apiErrors.NewForbidden(
			schema.GroupResource{}, "", errors.New("is forbidden: "+
				"exceeded quota: project-quota, requested: limits.memory=3Gi, "+
				"used: limits.memory=7976Gi, limited: limits.memory=8000Gi"))},
			want: true},
		{name: "False ResourceQuotaExceeded error msg", args: args{err: apiErrors.NewForbidden(
			schema.GroupResource{}, "", errors.New("is forbidden: "+
				"exxxceed quota: project-quota, requested: limits.memory=3Gi, "+
				"used: limits.memory=7976Gi, limited: limits.memory=8000Gi"))},
			want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsResourceQuotaExceeded(tt.args.err); got != tt.want {
				t.Errorf("IsResourceQuotaExceeded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimpleBackOffBlocker_backOff(t *testing.T) {
	tc := clock.NewFakeClock(time.Now())
	maxBackOffDuration := 10 * time.Minute
	type fields struct {
		Clock              clock.Clock
		BackOffBaseSecond  int
		BackOffExponent    uint32
		NextEligibleTime   time.Time
		MaxBackOffDuration time.Duration
	}
	tests := []struct {
		name         string
		fields       fields
		wantExponent uint32
		wantDuration time.Duration
	}{
		{name: "backoff should increase exponent",
			fields:       fields{Clock: tc, BackOffBaseSecond: 2, BackOffExponent: 0, NextEligibleTime: tc.Now(), MaxBackOffDuration: maxBackOffDuration},
			wantExponent: 1,
			wantDuration: time.Second * 1,
		},
		{name: "backoff should not saturate",
			fields:       fields{Clock: tc, BackOffBaseSecond: 2, BackOffExponent: 9, NextEligibleTime: tc.Now(), MaxBackOffDuration: maxBackOffDuration},
			wantExponent: 10,
			wantDuration: time.Second * 512,
		},
		{name: "backoff should saturate",
			fields:       fields{Clock: tc, BackOffBaseSecond: 2, BackOffExponent: 10, NextEligibleTime: tc.Now(), MaxBackOffDuration: maxBackOffDuration},
			wantExponent: 10,
			wantDuration: maxBackOffDuration,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &SimpleBackOffBlocker{
				Clock:              tt.fields.Clock,
				BackOffBaseSecond:  tt.fields.BackOffBaseSecond,
				BackOffExponent:    stdAtomic.NewUint32(tt.fields.BackOffExponent),
				NextEligibleTime:   NewAtomicTime(tt.fields.NextEligibleTime),
				MaxBackOffDuration: tt.fields.MaxBackOffDuration,
			}

			if got := b.backOff(context.Background()); !reflect.DeepEqual(got, tt.wantDuration) {
				t.Errorf("backOff() = %v, want %v", got, tt.wantDuration)
			}
			if gotExp := b.BackOffExponent; !reflect.DeepEqual(gotExp.Load(), tt.wantExponent) {
				t.Errorf("backOffExponent = %v, want %v", gotExp.Load(), tt.wantExponent)
			}
		})
	}

	t.Run("backoff many times after maxBackOffDuration is hit", func(t *testing.T) {
		b := &SimpleBackOffBlocker{
			Clock:              tc,
			BackOffBaseSecond:  2,
			BackOffExponent:    stdAtomic.NewUint32(10),
			NextEligibleTime:   NewAtomicTime(tc.Now()),
			MaxBackOffDuration: maxBackOffDuration,
		}

		for i := 0; i < 10; i++ {
			backOffDuration := b.backOff(context.Background())
			assert.Equal(t, maxBackOffDuration, backOffDuration)
			assert.Equal(t, uint32(10), b.BackOffExponent.Load())
		}
	})
}

func TestErrorTypes(t *testing.T) {
	t.Run("too many requests", func(t *testing.T) {
		err := apiErrors.NewTooManyRequestsError("test")
		res := IsBackOffError(err)
		assert.True(t, res)
	})

	t.Run("server timeout", func(t *testing.T) {
		err := apiErrors.NewServerTimeout(schema.GroupResource{}, "test op", 1)
		res := IsBackOffError(err)
		assert.True(t, res)
	})

	t.Run("resource exceeded", func(t *testing.T) {
		err := apiErrors.NewForbidden(
			schema.GroupResource{}, "", errors.New("is forbidden: "+
				"exceeded quota: project-quota, requested: limits.memory=1Gi, "+
				"used: limits.memory=7976Gi, limited: limits.memory=8000Gi"))
		res := IsBackOffError(err)
		assert.True(t, res)
	})
}

func TestIsEligible(t *testing.T) {
	type args struct {
		requested v1.ResourceList
		quota     v1.ResourceList
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "CPUElgible",
			args: args{
				requested: v1.ResourceList{v1.ResourceCPU: resource.MustParse("1")},
				quota:     v1.ResourceList{v1.ResourceCPU: resource.MustParse("2")},
			},
			want: true,
		},
		{
			name: "CPUInelgible",
			args: args{
				requested: v1.ResourceList{v1.ResourceCPU: resource.MustParse("1")},
				quota:     v1.ResourceList{v1.ResourceCPU: resource.MustParse("200m")},
			},
			want: false,
		},
		{
			name: "MemoryElgible",
			args: args{
				requested: v1.ResourceList{v1.ResourceMemory: resource.MustParse("32Gi")},
				quota:     v1.ResourceList{v1.ResourceMemory: resource.MustParse("64Gi")},
			},
			want: true,
		},
		{
			name: "MemoryInelgible",
			args: args{
				requested: v1.ResourceList{v1.ResourceMemory: resource.MustParse("64Gi")},
				quota:     v1.ResourceList{v1.ResourceMemory: resource.MustParse("64Gi")},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEligible(tt.args.requested, tt.args.quota); got != tt.want {
				t.Errorf("isEligible() = %v, want %v", got, tt.want)
			}
		})
	}
}
