package cache

import (
	"context"
	"reflect"

	"github.com/eko/gocache/lib/v4/store"
)

// TypedMarshaler is the struct that handles typed marshaling and unmarshaling
type TypedMarshaler[T any] struct {
	*Marshaler
	isPointerType bool
	elemType      reflect.Type
}

func (t *TypedMarshaler[T]) Set(ctx context.Context, key any, object T, options ...store.Option) error {
	return t.Marshaler.Set(ctx, key, object, options...)
}

func (t *TypedMarshaler[T]) Get(ctx context.Context, key any) (T, error) {
	if t.isPointerType {
		justT := reflect.New(t.elemType).Interface().(T)
		obj, err := t.Marshaler.Get(ctx, key, justT)
		if err != nil {
			return *new(T), err
		}

		return obj.(T), nil
	}

	obj, err := t.Marshaler.Get(ctx, key, new(T))
	if err != nil {
		return *new(T), err
	}

	return *(obj.(*T)), nil
}

// NewTypedMarshaler creates a new typed marshaler. It takes a marshaler and a type T.
// It returns a typed marshaler that can be used to marshal and unmarshal values of type T.
// If T is a pointer type, it will unmarshal into a new instance of T. Otherwise, it will unmarshal into a new instance of *T.
func NewTypedMarshaler[T any](marshaler *Marshaler) *TypedMarshaler[T] {
	isPointerType := reflect.TypeOf(new(T)).Elem().Kind() == reflect.Ptr
	var elemType reflect.Type
	if isPointerType {
		elemType = reflect.TypeOf(new(T)).Elem().Elem()
	}

	return &TypedMarshaler[T]{
		Marshaler:     marshaler,
		isPointerType: isPointerType,
		elemType:      elemType,
	}
}
