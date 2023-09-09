package plugins

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	assert.NotNil(t, &r.m)
	assert.NotNil(t, &r.mDefault)
}

func TestNewAtomicRegistry(t *testing.T) {
	ar := NewAtomicRegistry(nil)
	r := NewRegistry()
	r.RegisterDefault(PluginIDDataProxy, 5)
	ar.Store(r)
	r = ar.Load()
	assert.Equal(t, 5, r.Get(PluginIDDataProxy))
}

type PreRedirectHookFunc func(ctx context.Context) error

func TestRedirectHook(t *testing.T) {
	ar := NewAtomicRegistry(nil)
	r := NewRegistry()

	var redirectHookfn PreRedirectHookFunc = func(ctx context.Context) error {
		return fmt.Errorf("redirect hook error")
	}
	err := r.Register(PluginIDPreRedirectHook, redirectHookfn)
	assert.NoError(t, err)
	ar.Store(r)
	r = ar.Load()
	fn := Get[PreRedirectHookFunc](r, PluginIDPreRedirectHook)
	err = fn(context.Background())
	assert.Equal(t, fmt.Errorf("redirect hook error"), err)
}

func TestRegistry_RegisterDefault(t *testing.T) {
	r := NewRegistry()
	r.RegisterDefault("hello", 5)
	assert.Equal(t, 5, r.Get("hello"))
	assert.NotEqual(t, 5, r.Get("world"))
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	r.RegisterDefault("hello", 5)
	assert.NoError(t, r.Register("hello", 2))
	assert.Equal(t, 2, r.Get("hello"))
	assert.NotEqual(t, 5, r.Get("world"))

	assert.Error(t, r.Register("hello", 5))
}

func TestGet(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		r := NewRegistry()
		r.RegisterDefault("hello", 5)
		assert.Equal(t, 5, Get[int](r, "hello"))
	})

	t.Run("invalid type", func(t *testing.T) {
		r := NewRegistry()
		r.RegisterDefault("hello", 5)
		assert.Equal(t, int64(0), Get[int64](r, "hello"))
	})
}
