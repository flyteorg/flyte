package service

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	config2 "github.com/flyteorg/flyte/v2/dataproxy/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
)

func TestSafeSubpathRejectsAnythingThatClimbsOut(t *testing.T) {
	// A `..` would be caught by the scope check a moment later anyway, but
	// failing here says which argument was wrong.
	for _, bad := range []string{"", "/absolute", "..", "../sibling", "a/../..", "/"} {
		assert.False(t, safeSubpath(bad), "%q should be rejected", bad)
	}
	for _, good := range []string{"datasets", "datasets/train", "a/b/c.bin"} {
		assert.True(t, safeSubpath(good), "%q should be accepted", good)
	}
}

func TestWithinScope(t *testing.T) {
	assert.True(t, withinScope("uploads/p/d", "uploads/p/d"))
	assert.True(t, withinScope("uploads/p/d/sub", "uploads/p/d"))
	assert.True(t, withinScope("/uploads/p/d/sub/", "uploads/p/d"))

	assert.False(t, withinScope("uploads/p/other", "uploads/p/d"))
	// The prefix-string trap: "uploads/p/development" must not pass as being
	// inside "uploads/p/dev".
	assert.False(t, withinScope("uploads/p/development", "uploads/p/dev"))
	assert.False(t, withinScope("", "uploads/p/d"))
}

// An unconfigured store has no project layout, so nothing is inside its scope
// and every request must be refused. Defaulting the other way would vend
// bucket-wide credentials on a misconfiguration.
func TestWithinScopeRefusesWhenThereIsNoScope(t *testing.T) {
	assert.False(t, withinScope("anything", ""))
	assert.False(t, withinScope("", ""))
}

func TestProjectDataPrefixMirrorsTheUploadLayout(t *testing.T) {
	s := &Service{cfg: configWithPrefix("uploads")}

	assert.Equal(t, "uploads/myproject/development", s.projectDataPrefix("myproject", "development"))
}

func TestProjectDataPrefixIsEmptyWithoutAProjectOrDomain(t *testing.T) {
	s := &Service{cfg: configWithPrefix("uploads")}

	assert.Empty(t, s.projectDataPrefix("", "development"))
	assert.Empty(t, s.projectDataPrefix("myproject", ""))
}

// Every vended session assumes the same role, so this label is all a
// CloudTrail reader has to tell two of them apart.
func TestSessionLabelIsSafeAndBounded(t *testing.T) {
	label := sessionLabel("my-project", "development", "volume-my_vol")

	assert.Equal(t, "my-project-development-volume-my_vol", label)

	long := sessionLabel("p", "d", string(make([]byte, 200)))
	assert.LessOrEqual(t, len(long), 64, "STS caps RoleSessionName at 64 characters")
	assert.NotContains(t, long, "\x00", "non-alphanumerics must be replaced, not passed through")
}

func TestCredentialDurationClamps(t *testing.T) {
	s := &Service{cfg: configWithDurations(time.Hour, 12*time.Hour)}

	assert.Equal(t, time.Hour, s.credentialDuration(0), "no request means the default")
	assert.Equal(t, 30*time.Minute, s.credentialDuration(30*time.Minute))
	assert.Equal(t, 12*time.Hour, s.credentialDuration(48*time.Hour), "over the max is clamped, not refused")
}

func TestGetCloudCredentialsIsUnimplementedWithoutAMinter(t *testing.T) {
	s := &Service{cfg: configWithPrefix("uploads")}

	_, err := s.GetCloudCredentials(context.Background(), connectRequest(&dataproxy.GetCloudCredentialsRequest{
		Scope:    projectScope(),
		Resource: &dataproxy.GetCloudCredentialsRequest_ProjectData{ProjectData: &dataproxy.ProjectDataArea{}},
	}))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot vend object-store credentials")
}

func TestBucketOf(t *testing.T) {
	assert.Equal(t, "my-bucket", bucketOf("s3://my-bucket/volumes/v1/metadata.json"))
	assert.Equal(t, "my-bucket", bucketOf("s3://my-bucket"))
	assert.Equal(t, "my-bucket", bucketOf("gs://my-bucket/p"))
	// Azure: the container plays the bucket's part, and the account must not
	// come back as one.
	assert.Equal(t, "cont", bucketOf("abfss://cont@acct.dfs.core.windows.net/pre"))
}

func TestBucketSubpath(t *testing.T) {
	assert.Equal(t, "volumes/versions", bucketSubpath("s3://b/volumes/versions"))
	assert.Equal(t, "volumes/versions", bucketSubpath("s3://b/volumes/versions/"))
	assert.Empty(t, bucketSubpath("s3://b"))
}

func TestSameBucket(t *testing.T) {
	assert.True(t, sameBucket("s3://b/volumes/v/metadata.json", "b"))
	assert.False(t, sameBucket("s3://other/volumes/v/metadata.json", "b"))
	// An unset bucket must not match everything.
	assert.False(t, sameBucket("s3://b/x", ""))
}

// --- helpers --------------------------------------------------------------

func configWithPrefix(prefix string) config2.DataProxyConfig {
	return config2.DataProxyConfig{Upload: config2.DataProxyUploadConfig{StoragePrefix: prefix}}
}

func configWithDurations(def, max time.Duration) config2.DataProxyConfig {
	return config2.DataProxyConfig{
		Credentials: config2.DataProxyCredentialsConfig{
			DefaultDuration: config.Duration{Duration: def},
			MaxDuration:     config.Duration{Duration: max},
		},
	}
}

func projectScope() *common.ProjectIdentifier {
	return &common.ProjectIdentifier{Organization: "org", Name: "myproject", Domain: "development"}
}

func connectRequest[T any](msg *T) *connect.Request[T] {
	return connect.NewRequest(msg)
}
