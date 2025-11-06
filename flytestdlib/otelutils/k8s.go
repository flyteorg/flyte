package otelutils

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const k8sSpanPathPrefix = "controller-runtime.pkg.client"

type K8sCacheWrapper struct {
	cache.Cache
}

func WrapK8sCache(c cache.Cache) cache.Cache {
	return &K8sCacheWrapper{c}
}

func (c *K8sCacheWrapper) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) (err error) {
	ctx, span := NewSpan(ctx, K8sClientTracer, fmt.Sprintf("%s.Cache/Get", k8sSpanPathPrefix))
	defer span.EndErr(err)
	return c.Cache.Get(ctx, key, obj, opts...)
}

func (c *K8sCacheWrapper) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (err error) {
	ctx, span := NewSpan(ctx, K8sClientTracer, fmt.Sprintf("%s.Cache/List", k8sSpanPathPrefix))
	defer span.EndErr(err)
	return c.Cache.List(ctx, list, opts...)
}

type K8sClientWrapper struct {
	client.Client

	statusWriter *K8sStatusWriterWrapper
}

func WrapK8sClient(c client.Client) client.Client {
	return &K8sClientWrapper{c, &K8sStatusWriterWrapper{c.Status()}}
}

func (c *K8sClientWrapper) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) (err error) {
	ctx, span := NewSpan(ctx, K8sClientTracer, fmt.Sprintf("%s.Client/Get", k8sSpanPathPrefix))
	defer span.EndErr(err)
	return c.Client.Get(ctx, key, obj, opts...)
}

func (c *K8sClientWrapper) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (err error) {
	ctx, span := NewSpan(ctx, K8sClientTracer, fmt.Sprintf("%s.Client/List", k8sSpanPathPrefix))
	defer span.EndErr(err)
	return c.Client.List(ctx, list, opts...)
}

func (c *K8sClientWrapper) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) (err error) {
	ctx, span := NewSpan(ctx, K8sClientTracer, fmt.Sprintf("%s.Client/Create", k8sSpanPathPrefix))
	defer span.EndErr(err)
	return c.Client.Create(ctx, obj, opts...)
}

func (c *K8sClientWrapper) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) (err error) {
	ctx, span := NewSpan(ctx, K8sClientTracer, fmt.Sprintf("%s.Client/Delete", k8sSpanPathPrefix))
	defer span.EndErr(err)
	return c.Client.Delete(ctx, obj, opts...)
}

func (c *K8sClientWrapper) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) (err error) {
	ctx, span := NewSpan(ctx, K8sClientTracer, fmt.Sprintf("%s.Client/Update", k8sSpanPathPrefix))
	defer span.EndErr(err)
	return c.Client.Update(ctx, obj, opts...)
}

func (c *K8sClientWrapper) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) (err error) {
	ctx, span := NewSpan(ctx, K8sClientTracer, fmt.Sprintf("%s.Client/Patch", k8sSpanPathPrefix))
	defer span.EndErr(err)
	return c.Client.Patch(ctx, obj, patch, opts...)
}

func (c *K8sClientWrapper) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) (err error) {
	ctx, span := NewSpan(ctx, K8sClientTracer, fmt.Sprintf("%s.Client/DeleteAllOf", k8sSpanPathPrefix))
	defer span.EndErr(err)
	return c.Client.DeleteAllOf(ctx, obj, opts...)
}

func (c *K8sClientWrapper) Status() client.StatusWriter {
	return c.statusWriter
}

type K8sStatusWriterWrapper struct {
	client.StatusWriter
}

func (s *K8sStatusWriterWrapper) Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) (err error) {
	ctx, span := NewSpan(ctx, K8sClientTracer, fmt.Sprintf("%s.StatusWriter/Update", k8sSpanPathPrefix))
	defer span.EndErr(err)
	return s.StatusWriter.Update(ctx, obj, opts...)
}

func (s *K8sStatusWriterWrapper) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) (err error) {
	ctx, span := NewSpan(ctx, K8sClientTracer, fmt.Sprintf("%s.StatusWriter/Patch", k8sSpanPathPrefix))
	defer span.EndErr(err)
	return s.StatusWriter.Patch(ctx, obj, patch, opts...)
}
