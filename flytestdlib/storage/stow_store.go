package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	s32 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/flyteorg/stow"
	"github.com/flyteorg/stow/azure"
	"github.com/flyteorg/stow/google"
	"github.com/flyteorg/stow/local"
	"github.com/flyteorg/stow/oracle"
	"github.com/flyteorg/stow/s3"
	"github.com/flyteorg/stow/swift"
	errs "github.com/pkg/errors"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
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
		return DataReference(fmt.Sprintf("abfs://%s", bucket))
	},
	local.Kind: func(bucket string) DataReference {
		return DataReference(fmt.Sprintf("file://%s", bucket))
	},
}

// RegisterStowKind registers a new kind of stow store.
func RegisterStowKind(kind string, f func(string) DataReference) error {
	if _, ok := fQNFn[kind]; ok {
		return fmt.Errorf("kind [%v] already registered", kind)
	}

	fQNFn[kind] = f
	return nil
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

	DeleteFailure labeled.Counter
	DeleteLatency labeled.StopWatch
}

// StowMetadata that will be returned
type StowMetadata struct {
	exists bool
	size   int64
	etag   string
}

func (s StowMetadata) Size() int64 {
	return s.size
}

func (s StowMetadata) Exists() bool {
	return s.exists
}

func (s StowMetadata) Etag() string {
	return s.etag
}

// Implements DataStore to talk to stow location store.
type StowStore struct {
	copyImpl
	loc          stow.Location
	signedURLLoc stow.Location
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
	return s.createContainer(ctx, locationIDMain, container)
}

func (s *StowStore) createContainer(ctx context.Context, locID locationID, container string) (stow.Container, error) {
	logger.Infof(ctx, "Attempting to create container [%s]", container)
	c, err := s.getLocation(locID).CreateContainer(container)
	if err != nil && !awsBucketAlreadyExists(err) && !IsExists(err) {
		return nil, fmt.Errorf("unable to initialize container [%v]. Error: %v", container, err)
	}
	return c, nil
}

func (s *StowStore) LoadContainer(ctx context.Context, container string, createIfNotFound bool) (stow.Container, error) {
	return s.loadContainer(ctx, locationIDMain, container, createIfNotFound)
}

func (s *StowStore) loadContainer(ctx context.Context, locID locationID, container string, createIfNotFound bool) (stow.Container, error) {
	c, err := s.getLocation(locID).Container(container)
	if err != nil {
		// IsNotFound is not always guaranteed to be returned if the underlying container doesn't exist!
		// As of stow v0.2.6, the call to get container elides the lookup when a bucket region is set for S3 containers.
		if IsNotFound(err) && createIfNotFound {
			c, err = s.createContainer(ctx, locID, container)
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

func (s *StowStore) getContainer(ctx context.Context, locID locationID, container string) (c stow.Container, err error) {
	if s.baseContainer != nil && s.baseContainer.Name() == container && locID == locationIDMain {
		return s.baseContainer, nil
	}

	if !s.enableDynamicContainerLoading && locID == locationIDMain {
		s.metrics.BadContainer.Inc(ctx)
		return nil, errs.Wrapf(stow.ErrNotFound, "Conf container:%v != Passed Container:%v. Dynamic loading is disabled", s.baseContainer.Name(), container)
	}

	containerID := locID.String() + container
	iface, ok := s.dynamicContainerMap.Load(containerID)
	if !ok {
		c, err := s.loadContainer(ctx, locID, container, false)
		if err != nil {
			logger.Errorf(ctx, "failed to load container [%s] dynamically, error %s", container, err)
			return nil, err
		}

		s.dynamicContainerMap.Store(containerID, c)
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

	container, err := s.getContainer(ctx, locationIDMain, c)
	if err != nil {
		return nil, err
	}

	t := s.metrics.HeadLatency.Start(ctx)
	item, err := container.Item(k)
	if err == nil {
		if _, err = item.Metadata(); err != nil {
			// Err will be caught below
		} else if size, err := item.Size(); err != nil {
			// Err will be caught below
		} else if etag, err := item.ETag(); err != nil {
			// Err will be caught below
		} else {
			t.Stop()
			return StowMetadata{
				exists: true,
				size:   size,
				etag:   etag,
			}, nil
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

	container, err := s.getContainer(ctx, locationIDMain, c)
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

	container, err := s.getContainer(ctx, locationIDMain, c)
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

// Delete removes the referenced data from the blob store.
func (s *StowStore) Delete(ctx context.Context, reference DataReference) error {
	_, c, k, err := reference.Split()
	if err != nil {
		s.metrics.BadReference.Inc(ctx)
		return err
	}

	container, err := s.getContainer(ctx, locationIDMain, c)
	if err != nil {
		return err
	}

	t := s.metrics.DeleteLatency.Start(ctx)
	defer t.Stop()

	if err := container.RemoveItem(k); err != nil {
		incFailureCounterForError(ctx, s.metrics.DeleteFailure, err)
		return errs.Wrapf(err, "failed to remove item at path %q from container", k)
	}

	return nil
}

func (s *StowStore) GetBaseContainerFQN(ctx context.Context) DataReference {
	return s.baseContainerFQN
}

func (s *StowStore) CreateSignedURL(ctx context.Context, reference DataReference, properties SignedURLProperties) (SignedURLResponse, error) {
	_, container, key, err := reference.Split()
	if err != nil {
		return SignedURLResponse{}, err
	}

	c, err := s.getContainer(ctx, locationIDSignedURL, container)
	if err != nil {
		return SignedURLResponse{}, err
	}

	urlStr, err := c.PreSignRequest(ctx, properties.Scope, key, stow.PresignRequestParams{
		ExpiresIn:  properties.ExpiresIn,
		ContentMD5: properties.ContentMD5,
	})

	if err != nil {
		return SignedURLResponse{}, err
	}

	urlVal, err := url.Parse(urlStr)
	if err != nil {
		return SignedURLResponse{}, err
	}

	return SignedURLResponse{
		URL: *urlVal,
	}, nil
}

type locationID uint

const (
	locationIDMain locationID = iota
	locationIDSignedURL
)

func (l locationID) String() string {
	return strconv.Itoa(int(l))
}

func (s *StowStore) getLocation(id locationID) stow.Location {
	switch id {
	case locationIDSignedURL:
		if s.signedURLLoc != nil {
			return s.signedURLLoc
		}

		fallthrough
	default:
		return s.loc
	}
}

func NewStowRawStore(baseContainerFQN DataReference, loc, signedURLLoc stow.Location, enableDynamicContainerLoading bool, metrics *dataStoreMetrics) (*StowStore, error) {
	self := &StowStore{
		loc:                           loc,
		signedURLLoc:                  signedURLLoc,
		baseContainerFQN:              baseContainerFQN,
		enableDynamicContainerLoading: enableDynamicContainerLoading,
		dynamicContainerMap:           sync.Map{},
		metrics:                       metrics.stowMetrics,
	}

	self.copyImpl = newCopyImpl(self, metrics.copyMetrics)
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

func newStowMetrics(scope promutils.Scope) *stowMetrics {
	failureTypeOption := labeled.AdditionalLabelsOption{Labels: []string{FailureTypeLabel.String()}}
	return &stowMetrics{
		BadReference: labeled.NewCounter("bad_key", "Indicates the provided storage reference/key is incorrectly formatted", scope, labeled.EmitUnlabeledMetric),
		BadContainer: labeled.NewCounter("bad_container", "Indicates request for a container that has not been initialized", scope, labeled.EmitUnlabeledMetric),

		HeadFailure: labeled.NewCounter("head_failure", "Indicates failure in HEAD for a given reference", scope, labeled.EmitUnlabeledMetric),
		HeadLatency: labeled.NewStopWatch("head", "Indicates time to fetch metadata using the Head API", time.Millisecond, scope, labeled.EmitUnlabeledMetric),

		ReadFailure:     labeled.NewCounter("read_failure", "Indicates failure in GET for a given reference", scope, labeled.EmitUnlabeledMetric, failureTypeOption),
		ReadOpenLatency: labeled.NewStopWatch("read_open", "Indicates time to first byte when reading", time.Millisecond, scope, labeled.EmitUnlabeledMetric),

		WriteFailure: labeled.NewCounter("write_failure", "Indicates failure in storing/PUT for a given reference", scope, labeled.EmitUnlabeledMetric, failureTypeOption),
		WriteLatency: labeled.NewStopWatch("write", "Time to write an object irrespective of size", time.Millisecond, scope, labeled.EmitUnlabeledMetric),

		DeleteFailure: labeled.NewCounter("delete_failure", "Indicates failure in removing/DELETE for a given reference", scope, labeled.EmitUnlabeledMetric, failureTypeOption),
		DeleteLatency: labeled.NewStopWatch("delete", "Time to delete an object irrespective of size", time.Millisecond, scope, labeled.EmitUnlabeledMetric),
	}
}

// Constructor for the StowRawStore
func newStowRawStore(_ context.Context, cfg *Config, metrics *dataStoreMetrics) (RawStore, error) {
	if cfg.InitContainer == "" {
		return nil, fmt.Errorf("initContainer is required even with `enable-multicontainer`")
	}

	var cfgMap stow.ConfigMap
	var kind string
	if len(cfg.Stow.Kind) > 0 && len(cfg.Stow.Config) > 0 {
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

	var signedURLLoc stow.Location
	if len(cfg.SignedURL.StowConfigOverride) > 0 {
		var newCfg stow.ConfigMap = make(map[string]string, len(cfgMap))
		MergeMaps(newCfg, cfgMap, cfg.SignedURL.StowConfigOverride)
		signedURLLoc, err = stow.Dial(kind, newCfg)
		if err != nil {
			return emptyStore, fmt.Errorf("unable to configure the storage for %s. Error: %v", kind, err)
		}
	}

	return NewStowRawStore(fn(cfg.InitContainer), loc, signedURLLoc, cfg.MultiContainerEnabled, metrics)
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
