package implementations

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
)

func TestAWSSplitURI(t *testing.T) {
	remoteURL := AWSRemoteURL{}
	s3Object, err := remoteURL.splitURI(context.Background(), "s3://i/am/valid")
	assert.Nil(t, err)
	assert.Equal(t, "i", s3Object.bucket)
	assert.Equal(t, "am/valid", s3Object.key)
}

func TestAWSSplitURI_InvalidScheme(t *testing.T) {
	remoteURL := AWSRemoteURL{}
	_, err := remoteURL.splitURI(context.Background(), "azure://i/am/invalid")
	assert.NotNil(t, err)
}

func TestSplitURI_InvalidDataReference(t *testing.T) {
	remoteURL := AWSRemoteURL{}
	_, err := remoteURL.splitURI(context.Background(), "s3://invalid\\")
	assert.NotNil(t, err)
}

// Mock s3.S3 interface for testing.
type mockS3Impl struct {
	headObjectFunc func(input *s3.HeadObjectInput) (*s3.HeadObjectOutput, error)
	getObjectFunc  func(input *s3.GetObjectInput) (req *request.Request, output *s3.GetObjectOutput)
}

func (m *mockS3Impl) HeadObject(input *s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
	return m.headObjectFunc(input)
}

func (m *mockS3Impl) GetObjectRequest(input *s3.GetObjectInput) (req *request.Request, output *s3.GetObjectOutput) {
	return m.getObjectFunc(input)
}

func TestAWSGet(t *testing.T) {
	contentLength := int64(100)
	presignDuration := 3 * time.Minute

	mockS3 := mockS3Impl{}
	mockS3.headObjectFunc = func(input *s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
		assert.Equal(t, "bucket", *input.Bucket)
		assert.Equal(t, "key", *input.Key)
		return &s3.HeadObjectOutput{
			ContentLength: &contentLength,
		}, nil
	}
	mockS3.getObjectFunc = func(input *s3.GetObjectInput) (req *request.Request, output *s3.GetObjectOutput) {
		assert.Equal(t, "bucket", *input.Bucket)
		assert.Equal(t, "key", *input.Key)
		return &request.Request{
				Operation: &request.Operation{},
				HTTPRequest: &http.Request{
					URL: &url.URL{
						Scheme: "www",
						Host:   "host",
						Path:   "path",
					},
				},
			}, &s3.GetObjectOutput{
				ContentLength: &contentLength,
			}
	}
	remoteURL := AWSRemoteURL{
		s3Client:        &mockS3,
		presignDuration: presignDuration,
	}
	urlBlob, err := remoteURL.Get(context.Background(), "s3://bucket/key")
	assert.Nil(t, err)
	assert.Equal(t, "www://host/path", urlBlob.Url)
	assert.Equal(t, contentLength, urlBlob.Bytes)
}
