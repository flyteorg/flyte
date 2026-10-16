package service

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	ststypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
	awsv1 "github.com/aws/aws-sdk-go/aws"
	v1creds "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	s3v1 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
)

const (
	accessRead  = dataproxy.CredentialAccess_CREDENTIAL_ACCESS_READ
	accessWrite = dataproxy.CredentialAccess_CREDENTIAL_ACCESS_READ_WRITE
)

type policyDoc struct {
	Version   string `json:"Version"`
	Statement []struct {
		Sid       string   `json:"Sid"`
		Effect    string   `json:"Effect"`
		Action    []string `json:"Action"`
		Resource  []string `json:"Resource"`
		Condition struct {
			StringLike struct {
				Prefix []string `json:"s3:prefix"`
			} `json:"StringLike"`
		} `json:"Condition"`
	} `json:"Statement"`
}

func parsePolicy(t *testing.T, raw string) policyDoc {
	t.Helper()
	var doc policyDoc
	require.NoError(t, json.Unmarshal([]byte(raw), &doc))
	return doc
}

func TestS3SessionPolicyScopesObjectsToPrefixes(t *testing.T) {
	raw, err := S3SessionPolicy("my-bucket", []string{"vol-a", "uploads/p/d"}, accessWrite)
	require.NoError(t, err)

	doc := parsePolicy(t, raw)
	assert.Equal(t, []string{
		"arn:aws:s3:::my-bucket/vol-a/*",
		"arn:aws:s3:::my-bucket/uploads/p/d/*",
	}, doc.Statement[0].Resource)
}

// The single most important assertion in this file. s3:ListBucket is a
// bucket-level action, so a prefix cannot be expressed in the resource ARN --
// without the condition, the statement grants listing of the WHOLE bucket.
// That is the easiest way to write a policy that looks scoped and is not.
func TestS3SessionPolicyConditionsListingOnThePrefix(t *testing.T) {
	raw, err := S3SessionPolicy("my-bucket", []string{"vol-a"}, accessRead)
	require.NoError(t, err)

	doc := parsePolicy(t, raw)
	listing := doc.Statement[1]
	assert.Equal(t, []string{"arn:aws:s3:::my-bucket"}, listing.Resource)
	assert.Equal(t, []string{"vol-a/*", "vol-a"}, listing.Condition.StringLike.Prefix)
}

func TestS3SessionPolicyReadGrantsOnlyGet(t *testing.T) {
	raw, err := S3SessionPolicy("b", []string{"p"}, accessRead)
	require.NoError(t, err)

	assert.Equal(t, []string{"s3:GetObject"}, parsePolicy(t, raw).Statement[0].Action)
}

// An interrupted multipart upload that cannot be aborted leaves a billed,
// invisible partial upload behind.
func TestS3SessionPolicyWriteIncludesMultipartVerbs(t *testing.T) {
	raw, err := S3SessionPolicy("b", []string{"p"}, accessWrite)
	require.NoError(t, err)

	actions := parsePolicy(t, raw).Statement[0].Action
	assert.Contains(t, actions, "s3:AbortMultipartUpload")
	assert.Contains(t, actions, "s3:ListMultipartUploadParts")
}

// Measured, not assumed: a data plane role need not have GetBucketLocation,
// and a session policy can only subtract -- so granting it would be theatre.
// The client is handed a region-qualified bucket URI instead, which is the
// only reason it never makes that call.
func TestS3SessionPolicyDoesNotPretendToGrantBucketLocation(t *testing.T) {
	raw, err := S3SessionPolicy("b", []string{"p"}, accessWrite)
	require.NoError(t, err)

	assert.NotContains(t, raw, "GetBucketLocation")
}

func TestS3SessionPolicyRefusesAWholeBucket(t *testing.T) {
	for _, prefixes := range [][]string{nil, {}, {""}, {"/"}} {
		_, err := S3SessionPolicy("b", prefixes, accessWrite)
		require.Error(t, err, "prefixes %v", prefixes)
		assert.Contains(t, err.Error(), "at least one prefix")
	}
}

func TestS3SessionPolicyRefusesAnOversizedPolicy(t *testing.T) {
	prefixes := make([]string, 0, 80)
	for i := 0; i < 80; i++ {
		prefixes = append(prefixes, strings.Repeat("p", 20))
	}

	_, err := S3SessionPolicy("b", prefixes, accessWrite)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "over STS's 2048 limit")
}

func TestS3SessionPolicyOneVolumeFitsComfortably(t *testing.T) {
	raw, err := S3SessionPolicy("union-oc-production-demo",
		[]string{"my-volume", "uploads/myproject/development/volumes/versions"}, accessWrite)
	require.NoError(t, err)

	assert.Less(t, len(raw), 1024)
}

// --- the minter -----------------------------------------------------------

type fakeSTS struct {
	calls   int
	lastIn  *sts.AssumeRoleWithWebIdentityInput
	expires time.Time
	err     error
}

func (f *fakeSTS) AssumeRoleWithWebIdentity(
	_ context.Context,
	in *sts.AssumeRoleWithWebIdentityInput,
	_ ...func(*sts.Options),
) (*sts.AssumeRoleWithWebIdentityOutput, error) {
	f.calls++
	f.lastIn = in
	if f.err != nil {
		return nil, f.err
	}
	expires := f.expires
	if expires.IsZero() {
		expires = time.Now().Add(time.Hour)
	}
	return &sts.AssumeRoleWithWebIdentityOutput{
		Credentials: &ststypes.Credentials{
			AccessKeyId:     aws.String("ASIA"),
			SecretAccessKey: aws.String("secret"),
			SessionToken:    aws.String("token"),
			Expiration:      aws.Time(expires),
		},
	}, nil
}

func newTestMinter(t *testing.T, fake *fakeSTS) *stsMinter {
	t.Helper()
	tokenFile := t.TempDir() + "/token"
	require.NoError(t, os.WriteFile(tokenFile, []byte("jwt-one\n"), 0o600))
	return &stsMinter{
		client:    fake,
		roleARN:   "arn:aws:iam::123456789012:role/dataplane",
		tokenFile: tokenFile,
		cached:    map[string]MintedCredentials{},
	}
}

func TestMintSendsThePolicyAndTheLabel(t *testing.T) {
	fake := &fakeSTS{}
	m := newTestMinter(t, fake)

	got, err := m.Mint(context.Background(), MintRequest{
		Bucket:       "b",
		Prefixes:     []string{"vol-a"},
		Access:       accessWrite,
		Duration:     time.Hour,
		SessionLabel: "proj-dev-volume-vol-a",
	})

	require.NoError(t, err)
	assert.Equal(t, dataproxy.CredentialKind_CREDENTIAL_KIND_AWS_SIGV4, got.Kind)
	assert.Equal(t, "ASIA", got.AccessKeyID)
	assert.Contains(t, aws.ToString(fake.lastIn.Policy), "arn:aws:s3:::b/vol-a/*")
	// The label is how a CloudTrail reader tells two vended sessions apart:
	// they all assume the same role.
	assert.Equal(t, "proj-dev-volume-vol-a", aws.ToString(fake.lastIn.RoleSessionName))
	assert.Equal(t, "jwt-one", aws.ToString(fake.lastIn.WebIdentityToken))
}

func TestMintCachesUntilNearlyExpired(t *testing.T) {
	fake := &fakeSTS{}
	m := newTestMinter(t, fake)
	req := MintRequest{Bucket: "b", Prefixes: []string{"p"}, Access: accessRead, Duration: time.Hour, SessionLabel: "l"}

	_, err := m.Mint(context.Background(), req)
	require.NoError(t, err)
	_, err = m.Mint(context.Background(), req)
	require.NoError(t, err)

	assert.Equal(t, 1, fake.calls, "a second identical request should reuse the session")
}

func TestMintDoesNotServeASessionInsideTheRefreshMargin(t *testing.T) {
	fake := &fakeSTS{expires: time.Now().Add(2 * time.Minute)}
	m := newTestMinter(t, fake)
	req := MintRequest{Bucket: "b", Prefixes: []string{"p"}, Access: accessRead, Duration: time.Hour, SessionLabel: "l"}

	_, err := m.Mint(context.Background(), req)
	require.NoError(t, err)
	_, err = m.Mint(context.Background(), req)
	require.NoError(t, err)

	assert.Equal(t, 2, fake.calls, "a credential about to expire must be replaced, not handed out")
}

// Two callers that differ only in access must not share a session, or a
// read-only request gets a writable credential.
func TestMintDoesNotShareAcrossAccess(t *testing.T) {
	fake := &fakeSTS{}
	m := newTestMinter(t, fake)
	base := MintRequest{Bucket: "b", Prefixes: []string{"p"}, Duration: time.Hour, SessionLabel: "l"}

	readReq, writeReq := base, base
	readReq.Access, writeReq.Access = accessRead, accessWrite
	_, err := m.Mint(context.Background(), readReq)
	require.NoError(t, err)
	_, err = m.Mint(context.Background(), writeReq)
	require.NoError(t, err)

	assert.Equal(t, 2, fake.calls)
}

// The kubelet rotates the projected token in place. A minter that read it once
// starts failing with InvalidIdentityToken an hour in, which reads like a bug
// in this code and is not.
func TestMintRereadsTheTokenEveryExchange(t *testing.T) {
	fake := &fakeSTS{expires: time.Now().Add(2 * time.Minute)}
	m := newTestMinter(t, fake)
	req := MintRequest{Bucket: "b", Prefixes: []string{"p"}, Access: accessRead, Duration: time.Hour, SessionLabel: "l"}
	_, err := m.Mint(context.Background(), req)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(m.tokenFile, []byte("jwt-two\n"), 0o600))
	_, err = m.Mint(context.Background(), req)
	require.NoError(t, err)

	assert.Equal(t, "jwt-two", aws.ToString(fake.lastIn.WebIdentityToken))
}

func TestExplainSTSFailureNamesTheLikelyCause(t *testing.T) {
	cases := map[string]string{
		"InvalidIdentityToken":  "rotated in place",
		"PackedPolicyTooLarge":  "2048",
		"AccessDenied":          "can only subtract",
		"something else broken": "exchange failed",
	}
	for input, want := range cases {
		assert.Contains(t, explainSTSFailure(errors.New(input)), want, "input %q", input)
	}
}

// --- against real STS, where there is an identity to narrow ---------------

// This is the assertion no fake can make: that the policy DENIES. Every way of
// getting a policy wrong still mints a session and reads its own data fine, so
// a test that only checks the happy path checks nothing that matters.
//
// Runs wherever a Kubernetes web identity exists (any pod with IRSA) and skips
// everywhere else, including CI.
func TestScopedSessionIsDeniedOutsideItsPrefix(t *testing.T) {
	roleARN, tokenFile := os.Getenv("AWS_ROLE_ARN"), os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")
	bucket := os.Getenv("FLYTE_TEST_CREDENTIAL_BUCKET")
	if roleARN == "" || tokenFile == "" || bucket == "" {
		t.Skip("needs a web identity (AWS_ROLE_ARN + AWS_WEB_IDENTITY_TOKEN_FILE) and " +
			"$FLYTE_TEST_CREDENTIAL_BUCKET")
	}
	if _, err := os.Stat(tokenFile); err != nil {
		t.Skipf("no readable projected token: %v", err)
	}

	ctx := context.Background()
	minter := NewSTSMinter(ctx)
	require.NotNil(t, minter, "a pod with a web identity must be able to mint")

	creds, err := minter.Mint(ctx, MintRequest{
		Bucket:       bucket,
		Prefixes:     []string{"no-such-prefix-for-this-test"},
		Access:       accessRead,
		Duration:     15 * time.Minute,
		SessionLabel: "flyte-dataproxy-test",
	})
	require.NoError(t, err)
	require.NotEmpty(t, creds.SessionToken, "a scoped session must carry a session token")
	assert.True(t, creds.Expiration.After(time.Now()))

	// The session must not be able to list anything but its own prefix. Uses
	// aws-sdk-go v1, which is already a direct dependency, rather than adding
	// the v2 S3 module for one assertion.
	sess, err := session.NewSession(&awsv1.Config{
		Region: awsv1.String(regionOrDefault()),
		Credentials: v1creds.NewStaticCredentials(
			creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken),
	})
	require.NoError(t, err)
	_, err = s3v1.New(sess).ListObjectsV2WithContext(ctx, &s3v1.ListObjectsV2Input{
		Bucket: awsv1.String(bucket),
	})
	require.Error(t, err, "listing the bucket root with a prefix-scoped session must be denied")
	assert.Contains(t, err.Error(), "AccessDenied")
}

func regionOrDefault() string {
	if r := os.Getenv("AWS_REGION"); r != "" {
		return r
	}
	return "us-east-1"
}
