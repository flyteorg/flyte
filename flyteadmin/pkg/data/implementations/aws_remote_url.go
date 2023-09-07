package implementations

import (
	"context"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/flyteorg/flyteadmin/pkg/data/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"
	"google.golang.org/grpc/codes"
)

const s3Scheme = "s3"

// Defines a subset of the s3.S3 interface for easy mock-ability in testing.
type s3Interface interface {
	HeadObject(input *s3.HeadObjectInput) (*s3.HeadObjectOutput, error)
	GetObjectRequest(input *s3.GetObjectInput) (req *request.Request, output *s3.GetObjectOutput)
}

// AWS-specific implementation of RemoteURLInterface
type AWSRemoteURL struct {
	s3Client        s3Interface
	presignDuration time.Duration
}

type AWSS3Object struct {
	bucket string
	key    string
}

func (a *AWSRemoteURL) splitURI(ctx context.Context, uri string) (AWSS3Object, error) {
	scheme, container, key, err := storage.DataReference(uri).Split()
	if err != nil {
		return AWSS3Object{}, err
	}
	if scheme != s3Scheme {
		logger.Debugf(ctx, "encountered unexpected scheme: %s for AWS URI: %s", scheme, uri)
		return AWSS3Object{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"unexpected scheme %s for AWS URI", scheme)
	}
	return AWSS3Object{
		bucket: container,
		key:    key,
	}, nil
}

func (a *AWSRemoteURL) Get(ctx context.Context, uri string) (admin.UrlBlob, error) {
	logger.Debugf(ctx, "Getting signed url for - %s", uri)
	s3URI, err := a.splitURI(ctx, uri)
	if err != nil {
		logger.Debugf(ctx, "failed to extract s3 bucket and key from uri: %s", uri)
		return admin.UrlBlob{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid uri: %s", uri)
	}
	// First, get the size of the url blob.
	headResult, err := a.s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket: &s3URI.bucket,
		Key:    &s3URI.key,
	})
	if err != nil {
		logger.Debugf(ctx, "failed to get object size for %s with %v", uri, err)
		return admin.UrlBlob{}, errors.NewFlyteAdminErrorf(
			codes.Internal, "failed to get object size for %s with %v", uri, err)
	}

	// The second return argument here is the GetObjectOutput, which we don't use below.
	req, _ := a.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: &s3URI.bucket,
		Key:    &s3URI.key,
	})
	urlStr, err := req.Presign(a.presignDuration)
	if err != nil {
		logger.Warning(ctx,
			"failed to presign url for uri [%s] for %v with err %v", uri, a.presignDuration, err)
		return admin.UrlBlob{}, errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to presign url for uri [%s] for %v with err %v", uri, a.presignDuration, err)
	}
	var contentLength int64
	if headResult.ContentLength != nil {
		contentLength = *headResult.ContentLength
	}
	return admin.UrlBlob{
		Url:   urlStr,
		Bytes: contentLength,
	}, nil
}

func NewAWSRemoteURL(config *aws.Config, presignDuration time.Duration) interfaces.RemoteURLInterface {
	sesh, err := session.NewSession(config)
	if err != nil {
		panic(err)
	}
	s3Client := s3.New(sesh)
	return &AWSRemoteURL{
		s3Client:        s3Client,
		presignDuration: presignDuration,
	}
}
