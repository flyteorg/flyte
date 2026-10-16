package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
)

// MintRequest is a resolved request for credentials: a bucket, the prefixes
// inside it, and what the caller intends to do. Note what is NOT here -- a
// policy, or a set of actions. The verbs for an access are decided by the
// minter, because a caller that could name its own would be a caller that
// could ask for more than its resource.
type MintRequest struct {
	Bucket       string
	Prefixes     []string
	Access       dataproxy.CredentialAccess
	Duration     time.Duration
	SessionLabel string
}

// MintedCredentials is one credential, with the expiry that makes it
// refreshable. A client that cannot see an expiry cannot refresh, so it is
// never left unset.
type MintedCredentials struct {
	Kind            dataproxy.CredentialKind
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Expiration      time.Time
}

// CredentialMinter turns a resolved request into credentials. One
// implementation per cloud: the request is cloud-agnostic, the primitive that
// narrows a credential is not (AWS has session policies, GCS has Credential
// Access Boundaries, Azure has user-delegation SAS).
type CredentialMinter interface {
	Mint(ctx context.Context, req MintRequest) (MintedCredentials, error)
}

// stsMinter narrows the data proxy's OWN identity and hands the result out.
//
// The mechanism is worth stating because it is why this needs no IAM change in
// a customer's account: a pod running with IRSA holds a projected service
// account token, and its role's trust policy already permits that service
// account to assume it -- that is what IRSA is. Exchanging the same token
// again with a session Policy therefore asks for nothing new; a session policy
// can only SUBTRACT from the role it is minted against.
//
// Deliberately AssumeRoleWithWebIdentity rather than AssumeRole: assuming a
// role from an already-assumed role is role chaining, which AWS caps at one
// hour with no way to raise it. This is not chaining, so the ceiling is the
// role's own MaxSessionDuration.
type stsMinter struct {
	client    stsAPI
	roleARN   string
	tokenFile string

	mu     sync.Mutex
	cached map[string]MintedCredentials
}

type stsAPI interface {
	AssumeRoleWithWebIdentity(
		ctx context.Context,
		params *sts.AssumeRoleWithWebIdentityInput,
		optFns ...func(*sts.Options),
	) (*sts.AssumeRoleWithWebIdentityOutput, error)
}

// refreshMargin is how long before expiry a cached session stops being handed
// out. A credential that expires while a caller is still setting up its client
// is indistinguishable, from that caller's side, from a broken endpoint.
const refreshMargin = 5 * time.Minute

// NewSTSMinter returns a minter that narrows this process's own AWS identity,
// or nil when there is no web identity to narrow (not an error: a data proxy
// without one simply cannot vend, and GetCloudCredentials says so).
func NewSTSMinter(ctx context.Context) CredentialMinter {
	roleARN := os.Getenv("AWS_ROLE_ARN")
	tokenFile := os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")
	if roleARN == "" || tokenFile == "" {
		logger.Infof(ctx, "no web identity in this environment (AWS_ROLE_ARN / "+
			"AWS_WEB_IDENTITY_TOKEN_FILE), so object-store credentials cannot be vended")
		return nil
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Warnf(ctx, "could not load an AWS config for credential vending: %v", err)
		return nil
	}
	return &stsMinter{
		client:    sts.NewFromConfig(cfg),
		roleARN:   roleARN,
		tokenFile: tokenFile,
		cached:    map[string]MintedCredentials{},
	}
}

func (m *stsMinter) Mint(ctx context.Context, req MintRequest) (MintedCredentials, error) {
	policy, err := S3SessionPolicy(req.Bucket, req.Prefixes, req.Access)
	if err != nil {
		return MintedCredentials{}, err
	}

	// Cached on the policy plus the session label: together they are exactly
	// what the credential can do and who it is attributed to, so two requests
	// that agree on both can share a session. One STS call per refresh period
	// rather than one per request.
	key := req.SessionLabel + "\x00" + policy
	m.mu.Lock()
	if hit, ok := m.cached[key]; ok && time.Until(hit.Expiration) > refreshMargin {
		m.mu.Unlock()
		return hit, nil
	}
	m.mu.Unlock()

	// Read the token on every exchange, never once at construction: the
	// kubelet rotates the projected token in place, and a minter holding a
	// stale copy starts failing with InvalidIdentityToken an hour in -- which
	// reads like a bug in this code and is not.
	token, err := os.ReadFile(m.tokenFile)
	if err != nil {
		return MintedCredentials{}, fmt.Errorf("reading the projected identity token: %w", err)
	}

	out, err := m.client.AssumeRoleWithWebIdentity(ctx, &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          aws.String(m.roleARN),
		RoleSessionName:  aws.String(req.SessionLabel),
		WebIdentityToken: aws.String(strings.TrimSpace(string(token))),
		Policy:           aws.String(policy),
		DurationSeconds:  aws.Int32(int32(req.Duration.Seconds())),
	})
	if err != nil {
		return MintedCredentials{}, fmt.Errorf("%s: %w", explainSTSFailure(err), err)
	}
	if out.Credentials == nil || out.Credentials.Expiration == nil {
		return MintedCredentials{}, fmt.Errorf("STS answered without credentials or an expiry")
	}

	minted := MintedCredentials{
		Kind:            dataproxy.CredentialKind_CREDENTIAL_KIND_AWS_SIGV4,
		AccessKeyID:     aws.ToString(out.Credentials.AccessKeyId),
		SecretAccessKey: aws.ToString(out.Credentials.SecretAccessKey),
		SessionToken:    aws.ToString(out.Credentials.SessionToken),
		Expiration:      *out.Credentials.Expiration,
	}
	m.mu.Lock()
	m.cached[key] = minted
	m.mu.Unlock()
	return minted, nil
}

// explainSTSFailure names the two failures that look like a bug here and are
// not: an aged-out projected token, and a policy that asks for more than the
// role itself has.
func explainSTSFailure(err error) string {
	text := err.Error()
	switch {
	case strings.Contains(text, "InvalidIdentityToken"), strings.Contains(text, "ExpiredTokenException"):
		return "the projected service account token was not accepted (it is rotated in place, so it " +
			"must be read on every exchange)"
	case strings.Contains(text, "PackedPolicyTooLarge"):
		return "the session policy is over STS's 2048 character limit; scope to fewer prefixes"
	case strings.Contains(text, "AccessDenied"):
		return "STS refused the exchange; note that a session policy can only subtract, so a verb " +
			"the role itself lacks cannot be granted here"
	default:
		return "the scoped credential exchange failed"
	}
}

// S3SessionPolicy builds the policy that narrows a session to `prefixes`
// inside `bucket`.
//
// What is deliberately absent is s3:GetBucketLocation. A session policy can
// only subtract, and a data plane role does not necessarily have it -- so
// granting it here would be theatre. The response hands back a
// region-qualified bucket URI instead, which is what stops a client having to
// ask the store where its bucket is.
func S3SessionPolicy(bucket string, prefixes []string, access dataproxy.CredentialAccess) (string, error) {
	bucket = strings.Trim(bucket, "/")
	if bucket == "" {
		return "", fmt.Errorf("a session policy needs a bucket")
	}
	norm := make([]string, 0, len(prefixes))
	for _, p := range prefixes {
		if p = strings.Trim(p, "/"); p != "" {
			norm = append(norm, p)
		}
	}
	if len(norm) == 0 {
		return "", fmt.Errorf("a session policy needs at least one prefix; scoping to a whole bucket " +
			"defeats the point of minting a session")
	}

	objectActions := []string{"s3:GetObject"}
	if access == dataproxy.CredentialAccess_CREDENTIAL_ACCESS_READ_WRITE {
		// PutObject covers a single PUT and a multipart part upload;
		// AbortMultipartUpload and ListMultipartUploadParts are their own
		// verbs, and an interrupted large upload needs them to clean up after
		// itself instead of leaving a billed, invisible partial.
		objectActions = append(objectActions,
			"s3:PutObject",
			"s3:DeleteObject",
			"s3:AbortMultipartUpload",
			"s3:ListMultipartUploadParts",
		)
	}

	objectResources := make([]string, 0, len(norm))
	listPrefixes := make([]string, 0, len(norm)*2)
	for _, p := range norm {
		objectResources = append(objectResources, fmt.Sprintf("arn:aws:s3:::%s/%s/*", bucket, p))
		listPrefixes = append(listPrefixes, p+"/*", p)
	}

	policy := map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
			{
				"Sid":      "ResourceObjects",
				"Effect":   "Allow",
				"Action":   objectActions,
				"Resource": objectResources,
			},
			{
				"Sid":      "ResourceListing",
				"Effect":   "Allow",
				"Action":   []string{"s3:ListBucket"},
				"Resource": []string{fmt.Sprintf("arn:aws:s3:::%s", bucket)},
				// Without this condition the statement grants listing of the
				// WHOLE bucket: s3:ListBucket is a bucket-level action, so the
				// prefix cannot be expressed in the resource ARN. It is the
				// easiest way to write a policy that looks scoped and is not.
				"Condition": map[string]any{
					"StringLike": map[string]any{"s3:prefix": listPrefixes},
				},
			},
		},
	}

	encoded, err := json.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("encoding the session policy: %w", err)
	}
	if len(encoded) > 2048 {
		return "", fmt.Errorf("the session policy is %d characters, over STS's 2048 limit; scope to "+
			"fewer prefixes (%d given)", len(encoded), len(norm))
	}
	return string(encoded), nil
}
