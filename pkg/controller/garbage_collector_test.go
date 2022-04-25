package controller

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	config2 "github.com/flyteorg/flytepropeller/pkg/controller/config"

	"github.com/flyteorg/flytepropeller/pkg/client/clientset/versioned/typed/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	corev1Types "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/clock"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

func TestNewGarbageCollector(t *testing.T) {
	t.Run("enabled", func(t *testing.T) {
		cfg := &config2.Config{
			GCInterval:     config.Duration{Duration: time.Minute * 30},
			MaxTTLInHours:  2,
			LimitNamespace: "flyte",
		}
		gc, err := NewGarbageCollector(cfg, promutils.NewTestScope(), clock.NewFakeClock(time.Now()), nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, 2, gc.ttlHours)
	})

	t.Run("enabledBeyond23Hours", func(t *testing.T) {
		cfg := &config2.Config{
			GCInterval:     config.Duration{Duration: time.Minute * 30},
			MaxTTLInHours:  24,
			LimitNamespace: "flyte",
		}
		gc, err := NewGarbageCollector(cfg, promutils.NewTestScope(), clock.NewFakeClock(time.Now()), nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, 23, gc.ttlHours)
	})

	t.Run("ttl0", func(t *testing.T) {
		cfg := &config2.Config{
			GCInterval:     config.Duration{Duration: time.Minute * 30},
			MaxTTLInHours:  0,
			LimitNamespace: "flyte",
		}
		gc, err := NewGarbageCollector(cfg, promutils.NewTestScope(), nil, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, 0, gc.ttlHours)
		assert.NoError(t, gc.StartGC(context.TODO()))

	})

	t.Run("ttl-1", func(t *testing.T) {
		cfg := &config2.Config{
			GCInterval:     config.Duration{Duration: time.Minute * 30},
			MaxTTLInHours:  -1,
			LimitNamespace: "flyte",
		}
		gc, err := NewGarbageCollector(cfg, promutils.NewTestScope(), nil, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, -1, gc.ttlHours)
		assert.NoError(t, gc.StartGC(context.TODO()))
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
	wg := sync.WaitGroup{}
	b := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	mockWfClient := &mockWfClient{
		DeleteCollectionCb: func(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
			assert.NotNil(t, options)
			assert.NotNil(t, listOptions)
			if strings.HasPrefix(listOptions.LabelSelector, "completed-time") {
				assert.Equal(t, "completed-time notin (2009-11-10.21,2009-11-10.22,2009-11-10.23),!hour-of-day,termination-status=terminated", listOptions.LabelSelector)
			} else {
				assert.Equal(t, "hour-of-day in (0,1,10,11,12,13,14,15,16,17,18,19,2,20,21,3,4,5,6,7,8,9),termination-status=terminated", listOptions.LabelSelector)
			}
			wg.Done()
			return nil
		},
	}

	mockClient := &mockClient{
		FlyteWorkflowsCb: func(namespace string) v1alpha1.FlyteWorkflowInterface {
			return mockWfClient
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
		cfg := &config2.Config{
			GCInterval:     config.Duration{Duration: time.Minute * 30},
			MaxTTLInHours:  2,
			LimitNamespace: "flyte",
		}

		fakeClock := clock.NewFakeClock(b)
		mockNamespaceInvoked = false
		gc, err := NewGarbageCollector(cfg, promutils.NewTestScope(), fakeClock, mockNamespaceClient, mockClient)
		assert.NoError(t, err)
		wg.Add(2)
		ctx := context.TODO()
		ctx, cancel := context.WithCancel(ctx)
		assert.NoError(t, gc.StartGC(ctx))
		fakeClock.Step(time.Minute * 30)
		wg.Wait()
		cancel()
		assert.False(t, mockNamespaceInvoked)
	})

	t.Run("all-namespace", func(t *testing.T) {
		cfg := &config2.Config{
			GCInterval:     config.Duration{Duration: time.Minute * 30},
			MaxTTLInHours:  2,
			LimitNamespace: "all",
		}

		fakeClock := clock.NewFakeClock(b)
		mockNamespaceInvoked = false
		gc, err := NewGarbageCollector(cfg, promutils.NewTestScope(), fakeClock, mockNamespaceClient, mockClient)
		assert.NoError(t, err)
		wg.Add(4)
		ctx := context.TODO()
		ctx, cancel := context.WithCancel(ctx)
		assert.NoError(t, gc.StartGC(ctx))
		fakeClock.Step(time.Minute * 30)
		wg.Wait()
		cancel()
		assert.True(t, mockNamespaceInvoked)
	})
}
