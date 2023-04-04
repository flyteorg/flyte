package implementations

import (
	"context"
	"encoding/base64"
	"net/url"
	"testing"
	"time"

	gcs "cloud.google.com/go/storage"
	"github.com/golang/protobuf/ptypes/timestamp"
	gax "github.com/googleapis/gax-go/v2"
	"github.com/stretchr/testify/assert"
	credentialspb "google.golang.org/genproto/googleapis/iam/credentials/v1"
)

func TestGCPSplitURI(t *testing.T) {
	remoteURL := GCPRemoteURL{}
	gcsObject, err := remoteURL.splitURI(context.Background(), "gs://i/am/valid")
	assert.Nil(t, err)
	assert.Equal(t, "i", gcsObject.bucket)
	assert.Equal(t, "am/valid", gcsObject.object)
}

func TestGCPSplitURI_InvalidScheme(t *testing.T) {
	remoteURL := GCPRemoteURL{}
	_, err := remoteURL.splitURI(context.Background(), "azure://i/am/invalid")
	assert.NotNil(t, err)
}

func TestGCPSplitURI_InvalidDataReference(t *testing.T) {
	remoteURL := GCPRemoteURL{}
	_, err := remoteURL.splitURI(context.Background(), "gs://invalid\\")
	assert.NotNil(t, err)
}

type mockGCSImpl struct {
}

type mockBucketHandle struct {
	name string
}

type mockObjectHandle struct {
	bucket string
	name   string
}

func (m *mockGCSImpl) Bucket(name string) bucketHandleInterface {
	return &mockBucketHandle{
		name: name,
	}
}

func (m *mockBucketHandle) Object(name string) objectHandleInterface {
	return &mockObjectHandle{
		bucket: m.name,
		name:   name,
	}
}

func (m *mockObjectHandle) Attrs(ctx context.Context) (attrs *gcs.ObjectAttrs, err error) {
	return &gcs.ObjectAttrs{
		Bucket: m.bucket,
		Name:   m.name,
		Size:   int64(100),
	}, nil
}

type mockIAMCredentialsImpl struct {
	signBlobFunc            func(ctx context.Context, req *credentialspb.SignBlobRequest, opts ...gax.CallOption) (*credentialspb.SignBlobResponse, error)
	generateAccessTokenFunc func(ctx context.Context, req *credentialspb.GenerateAccessTokenRequest, opts ...gax.CallOption) (*credentialspb.GenerateAccessTokenResponse, error)
}

func (m *mockIAMCredentialsImpl) SignBlob(ctx context.Context, req *credentialspb.SignBlobRequest, opts ...gax.CallOption) (*credentialspb.SignBlobResponse, error) {
	return m.signBlobFunc(ctx, req, opts...)
}

func (m *mockIAMCredentialsImpl) GenerateAccessToken(ctx context.Context, req *credentialspb.GenerateAccessTokenRequest, opts ...gax.CallOption) (*credentialspb.GenerateAccessTokenResponse, error) {
	return m.generateAccessTokenFunc(ctx, req, opts...)
}

func TestGCPGet(t *testing.T) {
	signDuration := 3 * time.Minute
	signingPrincipal := "principal@example.com"
	signedBlob := "signed"
	encodedSignedBlob := base64.StdEncoding.EncodeToString([]byte(signedBlob))

	mockIAMCredentials := mockIAMCredentialsImpl{}
	mockIAMCredentials.signBlobFunc = func(ctx context.Context, req *credentialspb.SignBlobRequest, opts ...gax.CallOption) (*credentialspb.SignBlobResponse, error) {
		assert.Equal(t, "projects/-/serviceAccounts/"+signingPrincipal, req.Name)
		return &credentialspb.SignBlobResponse{SignedBlob: []byte(signedBlob)}, nil
	}

	mockGCS := mockGCSImpl{}
	remoteURL := GCPRemoteURL{
		iamCredentialsClient: &mockIAMCredentials,
		gcsClient:            &mockGCS,
		signDuration:         signDuration,
		signingPrincipal:     signingPrincipal,
	}
	urlBlob, err := remoteURL.Get(context.Background(), "gs://bucket/key")
	assert.Nil(t, err)

	u, _ := url.Parse(urlBlob.Url)
	assert.Equal(t, "https", u.Scheme)
	assert.Equal(t, "storage.googleapis.com", u.Hostname())
	assert.Equal(t, "/bucket/key", u.Path)
	assert.Equal(t, encodedSignedBlob, u.Query().Get("Signature"))
	assert.Equal(t, int64(100), urlBlob.Bytes)
}

func TestToken(t *testing.T) {
	token := "token"
	signingPrincipal := "principal@example.com"
	timestamp := timestamp.Timestamp{Seconds: int64(42)}

	mockIAMCredentials := mockIAMCredentialsImpl{}
	mockIAMCredentials.generateAccessTokenFunc = func(ctx context.Context, req *credentialspb.GenerateAccessTokenRequest, opts ...gax.CallOption) (*credentialspb.GenerateAccessTokenResponse, error) {
		assert.Equal(t, "projects/-/serviceAccounts/"+signingPrincipal, req.Name)
		assert.Equal(t, []string{"https://www.googleapis.com/auth/devstorage.read_only"}, req.Scope)
		return &credentialspb.GenerateAccessTokenResponse{
			AccessToken: token,
			ExpireTime:  &timestamp,
		}, nil
	}

	ts := impersonationTokenSource{
		iamCredentialsClient: &mockIAMCredentials,
		signingPrincipal:     signingPrincipal,
	}

	oauthToken, err := ts.Token()
	assert.Nil(t, err)
	assert.Equal(t, token, oauthToken.AccessToken)
	assert.Equal(t, asTime(&timestamp), oauthToken.Expiry)
}
