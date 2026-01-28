package cache

import (
	"context"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/protobuf/proto"
)

// Marshaler is the struct that marshal and unmarshal cache values
type Marshaler struct {
	cache.CacheInterface[any]
	useProto bool
}

// Get obtains a value from cache and unmarshal value with given object
func (c *Marshaler) Get(ctx context.Context, key any, returnObj any) (any, error) {
	result, err := c.CacheInterface.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var raw []byte
	switch v := result.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	}

	if c.useProto {
		err = proto.Unmarshal(raw, returnObj.(proto.Message))
	} else {
		err = msgpack.Unmarshal(raw, returnObj)
	}

	if err != nil {
		return nil, err
	}

	return returnObj, nil
}

// Set sets a value in cache by marshaling value
func (c *Marshaler) Set(ctx context.Context, key, object any, options ...store.Option) (err error) {
	var raw []byte
	if c.useProto {
		raw, err = proto.Marshal(object.(proto.Message))
		if err != nil {
			return err
		}

	} else {
		raw, err = msgpack.Marshal(object)
		if err != nil {
			return err
		}
	}

	return c.CacheInterface.Set(ctx, key, raw, options...)
}

// NewProtoMarshaler creates a new marshaler that marshals/unmarshals cache values using proto.Marshal.
func NewProtoMarshaler(cache cache.CacheInterface[any]) *Marshaler {
	return &Marshaler{
		CacheInterface: cache,
		useProto:       true,
	}
}

// NewMsgPackMarshaler creates a new marshaler that marshals/unmarshals cache values using MsgPack.
func NewMsgPackMarshaler(cache cache.CacheInterface[any]) *Marshaler {
	return &Marshaler{
		CacheInterface: cache,
		useProto:       false,
	}
}
