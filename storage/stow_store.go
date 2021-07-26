package storage

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/flyteorg/flytestdlib/errors"

	"github.com/aws/aws-sdk-go/aws/awserr"
	s32 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/graymeta/stow/azure"
	"github.com/graymeta/stow/google"
	"github.com/graymeta/stow/local"
	"github.com/graymeta/stow/oracle"
	"github.com/graymeta/stow/s3"
	"github.com/graymeta/stow/swift"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils/labeled"

	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/graymeta/stow"
	errs "github.com/pkg/errors"
)

const (
	FailureTypeLabel contextutils.Key = "failure_type"
)

var fQNFn = map[string]func(string) DataReference{
	s3.Kind: func(bucket string) DataReference {
		return DataReference(fmt.Sprintf("s3://%s", bucket))
	},
	google.Kind: func(bucket string) DataReference {
		return DataReference(fmt.Sprintf("gs://%s", bucket))
	},
	oracle.Kind: func(bucket string) DataReference {
		return DataReference(fmt.Sprintf("os://%s", bucket))
	},
	swift.Kind: func(bucket string) DataReference {
		return DataReference(fmt.Sprintf("sw://%s", bucket))
	},
	azure.Kind: func(bucket string) DataReference {
		return DataReference(fmt.Sprintf("afs://%s", bucket))
	},
	local.Kind: func(bucket string) DataReference {
		return DataReference(fmt.Sprintf("file://%s", bucket))
	},
}

// Checks if the error is AWS S3 bucket not found error
func awsBucketIsNotFound(err error) bool {
	if awsErr, errOk := errs.Cause(err).(awserr.Error); errOk {
		return awsErr.Code() == s32.ErrCodeNoSuchBucket
	}

	return false
}

// Checks if the error is AWS S3 bucket already exists error.
func awsBucketAlreadyExists(err error) bool {
	if IsExists(err) {
		return true
	}

	if awsErr, errOk := errs.Cause(err).(awserr.Error); errOk {
		return awsErr.Code() == s32.ErrCodeBucketAlreadyOwnedByYou
	}

	return false
}

// Metrics for Stow store
type stowMetrics struct {
	BadReference labeled.Counter
	BadContainer labeled.Counter

	HeadFailure labeled.Counter
	HeadLatency labeled.StopWatch

	ReadFailure     labeled.Counter
	ReadOpenLatency labeled.StopWatch

	WriteFailure labeled.Counter
	WriteLatency labeled.StopWatch
}

// Metadata that will be returned
type StowMetadata struct {
	exists bool
	size   int64
}

func (s StowMetadata) Size() int64 {
	return s.size
}

func (s StowMetadata) Exists() bool {
	return s.exists
}

// Implements DataStore to talk to stow location store.
type StowStore struct {
	copyImpl
	loc stow.Location
	// This is a default configured container.
	baseContainer stow.Container
	// If dynamic container loading is enabled, then for any new container that is not the base container
	// stowstore will dynamically load the given container
	enableDynamicContainerLoading bool
	// all dynamically loaded containers will be recorded in this map. It is possible that we may load the same container concurrently multiple times
	dynamicContainerMap sync.Map
	metrics             *stowMetrics
	baseContainerFQN    DataReference
}

func (s *StowStore) CreateContainer(ctx context.Context, container string) (stow.Container, error) {
	logger.Infof(ctx, "Attempting to create container [%s]", container)
	c, err := s.loc.CreateContainer(container)
	if err != nil && !awsBucketAlreadyExists(err) && !IsExists(err) {
		return nil, fmt.Errorf("unable to initialize container [%v]. Error: %v", container, err)
	}
	return c, nil
}

func (s *StowStore) LoadContainer(ctx context.Context, container string, createIfNotFound bool) (stow.Container, error) {
	c, err := s.loc.Container(container)
	if err != nil {
		// IsNotFound is not always guaranteed to be returned if the underlying container doesn't exist!
		// As of stow v0.2.6, the call to get container elides the lookup when a bucket region is set for S3 containers.
		if IsNotFound(err) && createIfNotFound {
			c, err = s.CreateContainer(ctx, container)
			if err != nil {
				logger.Errorf(ctx, "Call to create container [%s] failed. Error %s", container, err)
				return nil, err
			}
		} else {
			logger.Errorf(ctx, "Container [%s] lookup failed. Error %s", container, err)
			return nil, err
		}
	}
	return c, nil
}

func (s *StowStore) getContainer(ctx context.Context, container string) (c stow.Container, err error) {
	if s.baseContainer != nil && s.baseContainer.Name() == container {
		return s.baseContainer, nil
	}

	if !s.enableDynamicContainerLoading {
		s.metrics.BadContainer.Inc(ctx)
		return nil, errs.Wrapf(stow.ErrNotFound, "Conf container:%v != Passed Container:%v. Dynamic loading is disabled", s.baseContainer.Name(), container)
	}

	iface, ok := s.dynamicContainerMap.Load(container)
	if !ok {
		c, err := s.LoadContainer(ctx, container, false)
		if err != nil {
			logger.Errorf(ctx, "failed to load container [%s] dynamically, error %s", container, err)
			return nil, err
		}
		s.dynamicContainerMap.Store(container, c)
		return c, nil
	}
	return iface.(stow.Container), nil
}

func (s *StowStore) Head(ctx context.Context, reference DataReference) (Metadata, error) {
	_, c, k, err := reference.Split()
	if err != nil {
		s.metrics.BadReference.Inc(ctx)
		return nil, err
	}

	container, err := s.getContainer(ctx, c)
	if err != nil {
		return nil, err
	}

	t := s.metrics.HeadLatency.Start(ctx)
	item, err := container.Item(k)
	if err == nil {
		if _, err = item.Metadata(); err == nil {
			size, err := item.Size()
			if err == nil {
				t.Stop()
				return StowMetadata{
					exists: true,
					size:   size,
				}, nil
			}
		}
	}

	if IsNotFound(err) || awsBucketIsNotFound(err) {
		return StowMetadata{exists: false}, nil
	}

	incFailureCounterForError(ctx, s.metrics.HeadFailure, err)
	return StowMetadata{exists: false}, errs.Wrapf(err, "path:%v", k)
}

func (s *StowStore) ReadRaw(ctx context.Context, reference DataReference) (io.ReadCloser, error) {
	_, c, k, err := reference.Split()
	if err != nil {
		s.metrics.BadReference.Inc(ctx)
		return nil, err
	}

	container, err := s.getContainer(ctx, c)
	if err != nil {
		return nil, err
	}

	t := s.metrics.ReadOpenLatency.Start(ctx)
	item, err := container.Item(k)
	if err != nil {
		incFailureCounterForError(ctx, s.metrics.ReadFailure, err)
		return nil, err
	}
	t.Stop()

	sizeBytes, err := item.Size()
	if err != nil {
		return nil, err
	}

	if GetConfig().Limits.GetLimitMegabytes != 0 {
		if sizeMbs := sizeBytes / MiB; sizeMbs > GetConfig().Limits.GetLimitMegabytes {
			return nil, errors.Errorf(ErrExceedsLimit, "limit exceeded. %vmb > %vmb.", sizeMbs, GetConfig().Limits.GetLimitMegabytes)
		}
	}

	return item.Open()
}

func (s *StowStore) WriteRaw(ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader) error {
	_, c, k, err := reference.Split()
	if err != nil {
		s.metrics.BadReference.Inc(ctx)
		return err
	}

	container, err := s.getContainer(ctx, c)
	if err != nil {
		return err
	}

	t := s.metrics.WriteLatency.Start(ctx)
	_, err = container.Put(k, raw, size, opts.Metadata)
	if err != nil {
		// If this error is due to the bucket not existing, first attempt to create it and retry the getContainer call.
		if IsNotFound(err) || awsBucketIsNotFound(err) {
			container, err = s.CreateContainer(ctx, c)
			if err == nil {
				s.dynamicContainerMap.Store(container, c)
			}
		}
		if err != nil {
			incFailureCounterForError(ctx, s.metrics.WriteFailure, err)
			return errs.Wrapf(err, "Failed to write data [%vb] to path [%v].", size, k)
		}
	}

	t.Stop()

	return nil
}

func (s *StowStore) GetBaseContainerFQN(ctx context.Context) DataReference {
	return s.baseContainerFQN
}

func NewStowRawStore(baseContainerFQN DataReference, loc stow.Location, enableDynamicContainerLoading bool, metricsScope promutils.Scope) (*StowStore, error) {
	failureTypeOption := labeled.AdditionalLabelsOption{Labels: []string{FailureTypeLabel.String()}}
	self := &StowStore{
		loc:                           loc,
		baseContainerFQN:              baseContainerFQN,
		enableDynamicContainerLoading: enableDynamicContainerLoading,
		dynamicContainerMap:           sync.Map{},
		metrics: &stowMetrics{
			BadReference: labeled.NewCounter("bad_key", "Indicates the provided storage reference/key is incorrectly formatted", metricsScope, labeled.EmitUnlabeledMetric),
			BadContainer: labeled.NewCounter("bad_container", "Indicates request for a container that has not been initialized", metricsScope, labeled.EmitUnlabeledMetric),

			HeadFailure: labeled.NewCounter("head_failure", "Indicates failure in HEAD for a given reference", metricsScope, labeled.EmitUnlabeledMetric),
			HeadLatency: labeled.NewStopWatch("head", "Indicates time to fetch metadata using the Head API", time.Millisecond, metricsScope, labeled.EmitUnlabeledMetric),

			ReadFailure:     labeled.NewCounter("read_failure", "Indicates failure in GET for a given reference", metricsScope, labeled.EmitUnlabeledMetric, failureTypeOption),
			ReadOpenLatency: labeled.NewStopWatch("read_open", "Indicates time to first byte when reading", time.Millisecond, metricsScope, labeled.EmitUnlabeledMetric),

			WriteFailure: labeled.NewCounter("write_failure", "Indicates failure in storing/PUT for a given reference", metricsScope, labeled.EmitUnlabeledMetric, failureTypeOption),
			WriteLatency: labeled.NewStopWatch("write", "Time to write an object irrespective of size", time.Millisecond, metricsScope, labeled.EmitUnlabeledMetric),
		},
	}

	self.copyImpl = newCopyImpl(self, metricsScope)
	_, c, _, err := baseContainerFQN.Split()
	if err != nil {
		return nil, err
	}
	container, err := self.LoadContainer(context.TODO(), c, true)
	if err != nil {
		return nil, err
	}
	self.baseContainer = container
	return self, nil
}

// Constructor for the StowRawStore
func newStowRawStore(cfg *Config, metricsScope promutils.Scope) (RawStore, error) {
	if cfg.InitContainer == "" {
		return nil, fmt.Errorf("initContainer is required even with `enable-multicontainer`")
	}

	var cfgMap stow.ConfigMap
	var kind string
	if cfg.Stow != nil {
		kind = cfg.Stow.Kind
		cfgMap = cfg.Stow.Config
	} else {
		logger.Warnf(context.TODO(), "stow configuration section missing, defaulting to legacy s3/minio connection config")
		// This is for supporting legacy configurations which configure S3 via connection config
		kind = s3.Kind
		cfgMap = legacyS3ConfigMap(cfg.Connection)
	}

	fn, ok := fQNFn[kind]
	if !ok {
		return nil, errs.Errorf("unsupported stow.kind [%s], add support in flytestdlib?", kind)
	}

	loc, err := stow.Dial(kind, cfgMap)
	if err != nil {
		return emptyStore, fmt.Errorf("unable to configure the storage for %s. Error: %v", kind, err)
	}

	return NewStowRawStore(fn(cfg.InitContainer), loc, cfg.MultiContainerEnabled, metricsScope)
}

func legacyS3ConfigMap(cfg ConnectionConfig) stow.ConfigMap {
	// Non-nullable fields
	stowConfig := stow.ConfigMap{
		s3.ConfigAuthType: cfg.AuthType,
		s3.ConfigRegion:   cfg.Region,
	}

	// Fields that differ between minio and real S3
	if endpoint := cfg.Endpoint.String(); endpoint != "" {
		stowConfig[s3.ConfigEndpoint] = endpoint
	}

	if accessKey := cfg.AccessKey; accessKey != "" {
		stowConfig[s3.ConfigAccessKeyID] = accessKey
	}

	if secretKey := cfg.SecretKey; secretKey != "" {
		stowConfig[s3.ConfigSecretKey] = secretKey
	}

	if disableSsl := cfg.DisableSSL; disableSsl {
		stowConfig[s3.ConfigDisableSSL] = "True"
	}

	return stowConfig
}
