package storage

import (
	"bytes"
	"context"
	"fmt"
	"google.golang.org/protobuf/reflect/protoreflect"
	"time"

	protoV1 "github.com/golang/protobuf/proto"
	errs "github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/flytestdlib/ioutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

type protoMetrics struct {
	FetchLatency                 promutils.StopWatch
	MarshalTime                  promutils.StopWatch
	UnmarshalTime                promutils.StopWatch
	MarshalFailure               prometheus.Counter
	UnmarshalFailure             prometheus.Counter
	WriteFailureUnrelatedToCache prometheus.Counter
	ReadFailureUnrelatedToCache  prometheus.Counter
}

// Implements ProtobufStore to marshal and unmarshal protobufs to/from a RawStore
type DefaultProtobufStore struct {
	RawStore
	metrics *protoMetrics
}

func hasUnrecognizedFields(msg protoreflect.Message) bool {
	if msg == nil {
		return false
	}

	if len(msg.GetUnknown()) > 0 {
		return true
	}

	unrecognized := false

	msg.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		defer func() {
			if err := recover(); err != nil {
				return
			}
		}()

		if !v.IsValid() {
			return true
		}

		if fd.Kind() != protoreflect.MessageKind {
			return true
		}

		iface := v.Interface()
		if iface == nil {
			return true
			//} else if _, casted := iface.(protoV1.Message); !casted {
			//	unrecognized = true
			//	return false
		}

		if vMessage := v.Message(); vMessage != nil && len(v.Message().GetUnknown()) > 0 {
			unrecognized = true
			return false
		}

		unrecognized = hasUnrecognizedFields(v.Message())
		return !unrecognized
	})

	return unrecognized
}

func (s DefaultProtobufStore) ReadProtobufAny(ctx context.Context, reference DataReference, msg ...protoV1.Message) (msgIndex int, err error) {
	ctx, span := otelutils.NewSpan(ctx, otelutils.BlobstoreClientTracer, "flytestdlib.storage.DefaultProtobufStore/ReadProtobuf")
	defer span.End()

	rc, err := s.ReadRaw(ctx, reference)
	if err != nil && !IsFailedWriteToCache(err) {
		logger.Errorf(ctx, "Failed to read from the raw store [%s] Error: %v", reference, err)
		s.metrics.ReadFailureUnrelatedToCache.Inc()
		return -1, errs.Wrap(err, fmt.Sprintf("path:%v", reference))
	}

	defer func() {
		err = rc.Close()
		if err != nil {
			logger.Warnf(ctx, "Failed to close reference [%v]. Error: %v", reference, err)
		}
	}()

	docContents, err := ioutils.ReadAll(rc, s.metrics.FetchLatency.Start())
	if err != nil {
		return -1, errs.Wrap(err, fmt.Sprintf("readAll: %v", reference))
	}

	var lastErr error
	for i, m := range msg {
		t := s.metrics.UnmarshalTime.Start()
		var mCopy proto.Message
		v2Message := protoV1.MessageV2(m)
		if len(msg) > 1 {
			mCopy = proto.Clone(v2Message)
		}

		err = proto.UnmarshalOptions{DiscardUnknown: false, AllowPartial: false}.Unmarshal(docContents, v2Message)
		t.Stop()
		if err != nil {
			s.metrics.UnmarshalFailure.Inc()
			lastErr = errs.Wrap(err, fmt.Sprintf("unmarshall: %v", reference))
			continue
		}

		if len(msg) == 1 || (!hasUnrecognizedFields(protoV1.MessageV2(m).ProtoReflect()) && !proto.Equal(mCopy, v2Message)) {
			return i, nil
		}
	}

	return -1, lastErr
}

func (s DefaultProtobufStore) ReadProtobuf(ctx context.Context, reference DataReference, msg protoV1.Message) error {
	_, err := s.ReadProtobufAny(ctx, reference, msg)
	return err
}

func (s DefaultProtobufStore) WriteProtobuf(ctx context.Context, reference DataReference, opts Options, msg protoV1.Message) error {
	ctx, span := otelutils.NewSpan(ctx, otelutils.BlobstoreClientTracer, "flytestdlib.storage.DefaultProtobufStore/WriteProtobuf")
	defer span.End()

	t := s.metrics.MarshalTime.Start()
	raw, err := protoV1.Marshal(msg)
	t.Stop()
	if err != nil {
		s.metrics.MarshalFailure.Inc()
		return err
	}

	err = s.WriteRaw(ctx, reference, int64(len(raw)), opts, bytes.NewReader(raw))
	if err != nil && !IsFailedWriteToCache(err) {
		logger.Errorf(ctx, "Failed to write to the raw store [%s] Error: %v", reference, err)
		s.metrics.WriteFailureUnrelatedToCache.Inc()
		return err
	}
	return nil
}

func newProtoMetrics(scope promutils.Scope) *protoMetrics {
	return &protoMetrics{
		FetchLatency:                 scope.MustNewStopWatch("proto_fetch", "Time to read data before unmarshalling", time.Millisecond),
		MarshalTime:                  scope.MustNewStopWatch("marshal", "Time incurred in marshalling data before writing", time.Millisecond),
		UnmarshalTime:                scope.MustNewStopWatch("unmarshal", "Time incurred in unmarshalling received data", time.Millisecond),
		MarshalFailure:               scope.MustNewCounter("marshal_failure", "Failures when marshalling"),
		UnmarshalFailure:             scope.MustNewCounter("unmarshal_failure", "Failures when unmarshalling"),
		WriteFailureUnrelatedToCache: scope.MustNewCounter("write_failure_unrelated_to_cache", "Raw store write failures that are not caused by ErrFailedToWriteCache"),
		ReadFailureUnrelatedToCache:  scope.MustNewCounter("read_failure_unrelated_to_cache", "Raw store read failures that are not caused by ErrFailedToWriteCache"),
	}
}

func NewDefaultProtobufStore(store RawStore, scope promutils.Scope) DefaultProtobufStore {
	return NewDefaultProtobufStoreWithMetrics(store, newProtoMetrics(scope))
}

func NewDefaultProtobufStoreWithMetrics(store RawStore, metrics *protoMetrics) DefaultProtobufStore {
	return DefaultProtobufStore{
		RawStore: store,
		metrics:  metrics,
	}
}
