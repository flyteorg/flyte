package futures

import (
	"context"
	"fmt"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
)

func Example() {
	ctx := context.Background()
	f := NewAsyncFuture(ctx, func(ctx2 context.Context) (interface{}, error) {
		// can do large async / non-blocking work
		time.Sleep(time.Second)
		return "hello", nil
	})

	f.Ready()         // can be checked for completion
	_, _ = f.Get(ctx) // will block till the given sub-routine returns
}

func TestNewSyncFuture(t *testing.T) {
	type args struct {
		val interface{}
		err error
	}
	tests := []struct {
		name string
		args args
	}{
		{"val", args{val: "val"}},
		{"nil-val", args{}},
		{"error", args{err: fmt.Errorf("err")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSyncFuture(tt.args.val, tt.args.err)
			assert.NotNil(t, got)
			assert.Equal(t, tt.args.val, got.val)
			assert.Equal(t, tt.args.err, got.err)
			assert.True(t, got.Ready())
			v, err := got.Get(context.TODO())
			assert.Equal(t, tt.args.val, v)
			assert.Equal(t, tt.args.err, err)
		})
	}
}

func TestAsyncFuture(t *testing.T) {

	const val = "val"
	t.Run("immediate-return-val", func(t *testing.T) {
		v := val
		err := fmt.Errorf("err")
		af := NewAsyncFuture(context.TODO(), func(ctx context.Context) (interface{}, error) {
			return v, err
		})
		assert.NotNil(t, af)
		rv, rerr := af.Get(context.TODO())
		assert.Equal(t, v, rv)
		assert.Equal(t, err, rerr)
		assert.True(t, af.Ready())
	})

	t.Run("wait-return-val", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			v := val
			err := fmt.Errorf("err")
			af := NewAsyncFuture(t.Context(), func(ctx context.Context) (interface{}, error) {
				time.Sleep(time.Second)
				return v, err
			})
			assert.NotNil(t, af)
			rv, rerr := af.Get(t.Context())
			assert.Equal(t, v, rv)
			assert.Equal(t, err, rerr)
			assert.True(t, af.Ready())
		})
	})

	t.Run("timeout", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			v := val
			af := NewAsyncFuture(t.Context(), func(ctx context.Context) (interface{}, error) {
				select {
				case <-time.After(5 * time.Second):
					return v, nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			})
			synctest.Wait()
			cctx, cancel := context.WithCancel(t.Context())
			cancel()
			_, rerr := af.Get(cctx)
			assert.Error(t, rerr)
			assert.Equal(t, ErrAsyncFutureCanceled, rerr)
			synctest.Wait()
		})
	})
}
