package implementations

import (
	"context"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	gax "github.com/googleapis/gax-go/v2"
	"golang.org/x/oauth2"

	credentials "cloud.google.com/go/iam/credentials/apiv1"
	gcs "cloud.google.com/go/storage"
	"github.com/flyteorg/flyteadmin/pkg/data/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/api/option"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"
	"google.golang.org/grpc/codes"
)

const gcsScheme = "gs"

type iamCredentialsInterface interface {
	SignBlob(ctx context.Context, req *credentialspb.SignBlobRequest, opts ...gax.CallOption) (*credentialspb.SignBlobResponse, error)
	GenerateAccessToken(ctx context.Context, req *credentialspb.GenerateAccessTokenRequest, opts ...gax.CallOption) (*credentialspb.GenerateAccessTokenResponse, error)
}

type gcsClientWrapper struct {
	delegate *gcs.Client
}

type bucketHandleWrapper struct {
	delegate *gcs.BucketHandle
}

type objectHandleWrapper struct {
	delegate *gcs.ObjectHandle
}

type gcsInterface interface {
	Bucket(name string) bucketHandleInterface
}

type bucketHandleInterface interface {
	Object(name string) objectHandleInterface
}

type objectHandleInterface interface {
	Attrs(ctx context.Context) (attrs *gcs.ObjectAttrs, err error)
}

// GCP-specific implementation of RemoteURLInterface
type GCPRemoteURL struct {
	iamCredentialsClient iamCredentialsInterface
	gcsClient            gcsInterface
	signDuration         time.Duration
	signingPrincipal     string
}

type GCPGCSObject struct {
	bucket string
	object string
}

type impersonationTokenSource struct {
	iamCredentialsClient iamCredentialsInterface
	signingPrincipal     string
}

func (c *gcsClientWrapper) Bucket(name string) bucketHandleInterface {
	return &bucketHandleWrapper{delegate: c.delegate.Bucket(name)}
}

func (b *bucketHandleWrapper) Object(name string) objectHandleInterface {
	return &objectHandleWrapper{delegate: b.delegate.Object(name)}
}

func (o *objectHandleWrapper) Attrs(ctx context.Context) (attrs *gcs.ObjectAttrs, err error) {
	return o.delegate.Attrs(ctx)
}

func (g *GCPRemoteURL) splitURI(ctx context.Context, uri string) (GCPGCSObject, error) {
	scheme, container, key, err := storage.DataReference(uri).Split()
	if err != nil {
		return GCPGCSObject{}, err
	}
	if scheme != gcsScheme {
		logger.Debugf(ctx, "encountered unexpected scheme: %s for GCS URI: %s", scheme, uri)
		return GCPGCSObject{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"unexpected scheme %s for GCS URI", scheme)
	}
	return GCPGCSObject{
		bucket: container,
		object: key,
	}, nil
}

func (g *GCPRemoteURL) signURL(ctx context.Context, gcsURI GCPGCSObject) (string, error) {
	opts := &gcs.SignedURLOptions{
		Method:         "GET",
		GoogleAccessID: g.signingPrincipal,
		SignBytes: func(b []byte) ([]byte, error) {
			req := &credentialspb.SignBlobRequest{
				Payload: b,
				Name:    "projects/-/serviceAccounts/" + g.signingPrincipal,
			}
			resp, err := g.iamCredentialsClient.SignBlob(ctx, req)
			if err != nil {
				return nil, err
			}
			return resp.SignedBlob, nil
		},
		Expires: time.Now().Add(g.signDuration),
	}

	return gcs.SignedURL(gcsURI.bucket, gcsURI.object, opts)
}

func (g *GCPRemoteURL) Get(ctx context.Context, uri string) (admin.UrlBlob, error) {
	logger.Debugf(ctx, "Getting signed url for - %s", uri)
	gcsURI, err := g.splitURI(ctx, uri)
	if err != nil {
		logger.Debugf(ctx, "failed to extract gcs bucket and object from uri: %s", uri)
		return admin.UrlBlob{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "invalid uri: %s", uri)
	}

	// First, get the size of the url blob.
	attrs, err := g.gcsClient.Bucket(gcsURI.bucket).Object(gcsURI.object).Attrs(ctx)
	if err != nil {
		logger.Debugf(ctx, "failed to get object size for %s with %v", uri, err)
		return admin.UrlBlob{}, errors.NewFlyteAdminErrorf(
			codes.Internal, "failed to get object size for %s with %v", uri, err)
	}

	urlStr, err := g.signURL(ctx, gcsURI)
	if err != nil {
		logger.Warning(ctx,
			"failed to presign url for uri [%s] for %v with err %v", uri, g.signDuration, err)
		return admin.UrlBlob{}, errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to presign url for uri [%s] for %v with err %v", uri, g.signDuration, err)
	}
	return admin.UrlBlob{
		Url:   urlStr,
		Bytes: attrs.Size,
	}, nil
}

func (ts impersonationTokenSource) Token() (*oauth2.Token, error) {
	req := credentialspb.GenerateAccessTokenRequest{
		Name:  "projects/-/serviceAccounts/" + ts.signingPrincipal,
		Scope: []string{"https://www.googleapis.com/auth/devstorage.read_only"},
	}

	resp, err := ts.iamCredentialsClient.GenerateAccessToken(context.Background(), &req)
	if err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken: resp.AccessToken,
		Expiry:      asTime(resp.ExpireTime),
	}, nil
}

func asTime(t *timestamp.Timestamp) time.Time {
	return time.Unix(t.GetSeconds(), int64(t.GetNanos())).UTC()
}

func NewGCPRemoteURL(signingPrincipal string, signDuration time.Duration) interfaces.RemoteURLInterface {
	iamCredentialsClient, err := credentials.NewIamCredentialsClient(context.Background())
	if err != nil {
		panic(err)
	}

	gcsClient, err := gcs.NewClient(context.Background(),
		option.WithScopes(gcs.ScopeReadOnly),
		option.WithTokenSource(impersonationTokenSource{
			iamCredentialsClient: iamCredentialsClient,
			signingPrincipal:     signingPrincipal,
		}))
	if err != nil {
		panic(err)
	}

	return &GCPRemoteURL{
		iamCredentialsClient: iamCredentialsClient,
		gcsClient:            &gcsClientWrapper{delegate: gcsClient},
		signDuration:         signDuration,
		signingPrincipal:     signingPrincipal,
	}
}
