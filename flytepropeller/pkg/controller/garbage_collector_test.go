package controller

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1Types "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	k8stesting "k8s.io/utils/clock/testing"

	"github.com/flyteorg/flyte/flytepropeller/pkg/client/clientset/versioned/typed/flyteworkflow/v1alpha1"
	ctrlConfig "github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytestdlib/atomic"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

func TestNewGarbageCollector(t *testing.T) {
	t.Run("enabled", func(t *testing.T) {
		cfg := &ctrlConfig.Config{
			GCInterval:     config.Duration{Duration: time.Minute * 30},
			MaxTTL:         config.Duration{Duration: 2 * time.Hour},
			LimitNamespace: "flyte",
		}
		gc, err := NewGarbageCollector(cfg, promutils.NewTestScope(), k8stesting.NewFakeClock(time.Now()), nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, 2*time.Hour, gc.ttl)
	})

	t.Run("enabledBeyond23Hours", func(t *testing.T) {
		cfg := &ctrlConfig.Config{
			GCInterval:     config.Duration{Duration: time.Minute * 30},
			MaxTTL:         config.Duration{Duration: 24 * time.Hour},
			LimitNamespace: "flyte",
		}
		gc, err := NewGarbageCollector(cfg, promutils.NewTestScope(), k8stesting.NewFakeClock(time.Now()), nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, 23*time.Hour, gc.ttl)
	})

	t.Run("ttl0", func(t *testing.T) {
		cfg := &ctrlConfig.Config{
			GCInterval:     config.Duration{Duration: time.Minute * 30},
			MaxTTL:         config.Duration{Duration: 0 * time.Hour},
			LimitNamespace: "flyte",
		}
		gc, err := NewGarbageCollector(cfg, promutils.NewTestScope(), nil, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, 0*time.Hour, gc.ttl)
		assert.NoError(t, gc.StartGC(context.TODO()))

	})

	t.Run("ttl-1", func(t *testing.T) {
		cfg := &ctrlConfig.Config{
			GCInterval:     config.Duration{Duration: time.Minute * 30},
			MaxTTL:         config.Duration{Duration: -1 * time.Hour},
			LimitNamespace: "flyte",
		}
		gc, err := NewGarbageCollector(cfg, promutils.NewTestScope(), nil, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, 0*time.Hour, gc.ttl)
		assert.NoError(t, gc.StartGC(context.TODO()))
	})

	t.Run("ttl17m", func(t *testing.T) {
		cfg := &ctrlConfig.Config{
			GCInterval:     config.Duration{Duration: time.Minute * 30},
			MaxTTL:         config.Duration{Duration: 17 * time.Minute},
			LimitNamespace: "flyte",
		}
		_, err := NewGarbageCollector(cfg, promutils.NewTestScope(), nil, nil, nil)
		assert.EqualError(t, err, `invalid maxTTL. Allowed values: ["0s" "5m0s" "10m0s" "15m0s" "20m0s" "25m0s" "30m0s" "35m0s" "40m0s" "45m0s" "50m0s" "55m0s" "1h0m0s" "2h0m0s" "3h0m0s" "4h0m0s" "5h0m0s" "6h0m0s" "7h0m0s" "8h0m0s" "9h0m0s" "10h0m0s" "11h0m0s" "12h0m0s" "13h0m0s" "14h0m0s" "15h0m0s" "16h0m0s" "17h0m0s" "18h0m0s" "19h0m0s" "20h0m0s" "21h0m0s" "22h0m0s" "23h0m0s"]`)
	})
}

type mockWfClient struct {
	v1alpha1.FlyteWorkflowInterface
	DeleteCollectionCb func(options *v1.DeleteOptions, listOptions v1.ListOptions) error
}

func (m *mockWfClient) DeleteCollection(ctx context.Context, options v1.DeleteOptions, listOptions v1.ListOptions) error {
	return m.DeleteCollectionCb(&options, listOptions)
}

type mockClient struct {
	v1alpha1.FlyteworkflowV1alpha1Client
	FlyteWorkflowsCb func(namespace string) v1alpha1.FlyteWorkflowInterface
}

func (m *mockClient) FlyteWorkflows(namespace string) v1alpha1.FlyteWorkflowInterface {
	return m.FlyteWorkflowsCb(namespace)
}

type mockNamespaceClient struct {
	corev1.NamespaceInterface
	ListCb func(opts v1.ListOptions) (*corev1Types.NamespaceList, error)
}

func (m *mockNamespaceClient) List(ctx context.Context, opts v1.ListOptions) (*corev1Types.NamespaceList, error) {
	return m.ListCb(opts)
}

func TestGarbageCollector_StartGC(t *testing.T) {
	b := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	wfClient := &mockWfClient{}

	mockClient := &mockClient{
		FlyteWorkflowsCb: func(namespace string) v1alpha1.FlyteWorkflowInterface {
			return wfClient
		},
	}

	mockNamespaceInvoked := false
	mockNamespaceClient := &mockNamespaceClient{
		ListCb: func(opts v1.ListOptions) (*corev1Types.NamespaceList, error) {
			mockNamespaceInvoked = true
			return &corev1Types.NamespaceList{
				Items: []corev1Types.Namespace{
					{
						ObjectMeta: v1.ObjectMeta{
							Name: "ns1",
						},
					},
					{
						ObjectMeta: v1.ObjectMeta{
							Name: "ns2",
						},
					},
				},
			}, nil
		},
	}

	t.Run("one-namespace", func(t *testing.T) {
		cfg := &ctrlConfig.Config{
			GCInterval:     config.Duration{Duration: 30 * time.Minute},
			MaxTTL:         config.Duration{Duration: 2 * time.Hour},
			LimitNamespace: "flyte",
		}
		var deleteCollectionCalled atomic.Int32
		wfClient.DeleteCollectionCb = func(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
			assert.NotNil(t, options)
			assert.NotNil(t, listOptions)
			assert.Equal(t, "completed-time notin (2009-11-10.21,2009-11-10.22,2009-11-10.23),termination-status=terminated", listOptions.LabelSelector)
			deleteCollectionCalled.Add(1)
			return nil
		}
		fakeClock := k8stesting.NewFakeClock(b)
		mockNamespaceInvoked = false

		gc, err := NewGarbageCollector(cfg, promutils.NewTestScope(), fakeClock, mockNamespaceClient, mockClient)
		assert.NoError(t, err)
		ctx := context.TODO()
		ctx, cancel := context.WithCancel(ctx)
		assert.NoError(t, gc.StartGC(ctx))
		fakeClock.Step(cfg.GCInterval.Duration)
		assert.Eventually(t, func() bool { return deleteCollectionCalled.Load() == 1 }, time.Second, time.Millisecond)
		cancel()
		assert.False(t, mockNamespaceInvoked)
	})

	t.Run("all-namespace", func(t *testing.T) {
		cfg := &ctrlConfig.Config{
			GCInterval:     config.Duration{Duration: 30 * time.Minute},
			MaxTTL:         config.Duration{Duration: 2 * time.Hour},
			LimitNamespace: "all",
		}
		var deleteCollectionCalled atomic.Int32
		wfClient.DeleteCollectionCb = func(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
			assert.NotNil(t, options)
			assert.NotNil(t, listOptions)
			assert.Equal(t, "completed-time notin (2009-11-10.21,2009-11-10.22,2009-11-10.23),termination-status=terminated", listOptions.LabelSelector)
			deleteCollectionCalled.Add(1)
			return nil
		}

		fakeClock := k8stesting.NewFakeClock(b)
		mockNamespaceInvoked = false
		gc, err := NewGarbageCollector(cfg, promutils.NewTestScope(), fakeClock, mockNamespaceClient, mockClient)
		assert.NoError(t, err)
		ctx := context.TODO()
		ctx, cancel := context.WithCancel(ctx)
		assert.NoError(t, gc.StartGC(ctx))
		fakeClock.Step(cfg.GCInterval.Duration)
		assert.Eventually(t, func() bool { return deleteCollectionCalled.Load() == 2 }, time.Second, time.Millisecond)
		cancel()
		assert.True(t, mockNamespaceInvoked)
	})

	t.Run("all-namespaces-minute-granularity", func(t *testing.T) {
		cfg := &ctrlConfig.Config{
			GCInterval:     config.Duration{Duration: 10 * time.Minute},
			MaxTTL:         config.Duration{Duration: 15 * time.Minute},
			LimitNamespace: "all",
		}
		var deleteCollectionCalled atomic.Int32
		wfClient.DeleteCollectionCb = func(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
			assert.NotNil(t, options)
			assert.NotNil(t, listOptions)
			assert.Equal(t, "completed-time-minutes notin (2009-11-10T22_55,2009-11-10T23_00,2009-11-10T23_05,2009-11-10T23_10),termination-status=terminated", listOptions.LabelSelector)
			deleteCollectionCalled.Add(1)
			return nil
		}
		fakeClock := k8stesting.NewFakeClock(b)
		mockNamespaceInvoked = false

		gc, err := NewGarbageCollector(cfg, promutils.NewTestScope(), fakeClock, mockNamespaceClient, mockClient)
		assert.NoError(t, err)
		ctx := context.TODO()
		ctx, cancel := context.WithCancel(ctx)
		assert.NoError(t, gc.StartGC(ctx))
		fakeClock.Step(cfg.GCInterval.Duration)
		assert.Eventually(t, func() bool { return deleteCollectionCalled.Load() == 2 }, time.Second, time.Millisecond, "got %v", deleteCollectionCalled.Load())
		cancel()
		assert.True(t, mockNamespaceInvoked)
	})
}
