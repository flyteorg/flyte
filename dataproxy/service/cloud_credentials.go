package service

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
)

// GetCloudCredentials vends short-lived, prefix-scoped object-store credentials
// for one resource inside the caller's project/domain.
//
// It is the capability form of CreateUploadLocation, and the reason to have
// both is that they constrain different things. A signed URL is a capability
// for one object, and it can bind that object's checksum and size. These
// credentials are a capability over a prefix for a bounded time, which is the
// only workable shape for a client that touches many keys it discovers as it
// goes -- a mounted Volume reads and writes thousands of chunk keys, and a
// round trip per key is not a design.
//
// Three rules hold the security of this together, and all three are enforced
// below rather than documented and hoped for:
//
//  1. The caller names a RESOURCE inside a SCOPE, never a path. Authorization
//     is the scope question the platform already answers.
//  2. Every prefix granted is inside that scope. A locator is a path, and a
//     path a client chose is not authorization, so a volume whose data lives
//     outside the caller's scope is refused rather than vended.
//  3. The policy is computed here from the resource. Nothing the client sends
//     can widen it: `subpaths` only narrows, and there is no way to pass verbs
//     or a policy.
func (s *Service) GetCloudCredentials(
	ctx context.Context,
	req *connect.Request[dataproxy.GetCloudCredentialsRequest],
) (*connect.Response[dataproxy.GetCloudCredentialsResponse], error) {
	msg := req.Msg
	if err := msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if s.credentialMinter == nil {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf(
			"this data proxy cannot vend object-store credentials: no credential minter is "+
				"configured for storage type %q", s.storageKind()))
	}

	scope := msg.GetScope()
	access := msg.GetAccess()
	if access == dataproxy.CredentialAccess_CREDENTIAL_ACCESS_UNSPECIFIED {
		// The safe direction for a caller that did not say.
		access = dataproxy.CredentialAccess_CREDENTIAL_ACCESS_READ
	}

	// Authorization is the scope question, answered the same way
	// CreateUploadLocation answers it.
	if err := s.validateProjectExists(ctx, scope.GetName()); err != nil {
		return nil, err
	}

	bucket, err := s.storeBucket(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	resolved, err := s.resolveCredentialResource(ctx, msg, bucket)
	if err != nil {
		return nil, err
	}
	if len(resolved.prefixes) == 0 {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf(
			"resolved no prefixes for the request, and credentials for a whole bucket are never vended"))
	}

	// Rule 2. Checked after resolution, because resolution is what turns a
	// resource into paths, and it is the paths that have to be in bounds.
	for _, p := range resolved.prefixes {
		if !withinScope(p, resolved.scopePrefix) {
			logger.Warnf(ctx, "refusing credentials: prefix %q resolved outside scope %q", p, resolved.scopePrefix)
			return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf(
				"the requested resource's data lives outside %s/%s/%s, so credentials for it cannot be "+
					"authorized by that scope", scope.GetOrganization(), scope.GetName(), scope.GetDomain()))
		}
	}

	duration := s.credentialDuration(msg.GetDuration().AsDuration())
	creds, err := s.credentialMinter.Mint(ctx, MintRequest{
		Bucket:       bucket,
		Prefixes:     resolved.prefixes,
		Access:       access,
		Duration:     duration,
		SessionLabel: sessionLabel(scope.GetName(), scope.GetDomain(), resolved.label),
	})
	if err != nil {
		// Deliberately not wrapping the minter's error into the response: an
		// STS failure can carry a role ARN and the policy that was rejected.
		logger.Errorf(ctx, "minting credentials failed: %v", err)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf(
			"could not mint object-store credentials for this resource; see the data proxy logs"))
	}

	logger.Infof(ctx, "vended %s credentials for %s/%s (access=%s, prefixes=%v, expires=%s)",
		creds.Kind, scope.GetName(), scope.GetDomain(), access,
		resolved.prefixes, creds.Expiration.Format(time.RFC3339))

	return connect.NewResponse(&dataproxy.GetCloudCredentialsResponse{
		Kind:            creds.Kind,
		AccessKeyId:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
		Expiration:      timestamppb.New(creds.Expiration),
		BucketUri:       s.clientBucketURI(bucket),
		GrantedPrefixes: resolved.prefixes,
	}), nil
}

// resolvedResource is what a request's resource turns into: the prefixes it
// occupies, the scope those must sit inside, and a label for the audit trail.
type resolvedResource struct {
	prefixes    []string
	scopePrefix string
	label       string
}

func (s *Service) resolveCredentialResource(
	ctx context.Context,
	msg *dataproxy.GetCloudCredentialsRequest,
	bucket string,
) (resolvedResource, error) {
	scope := msg.GetScope()
	scopePrefix := s.projectDataPrefix(scope.GetName(), scope.GetDomain())

	switch msg.GetResource().(type) {
	case *dataproxy.GetCloudCredentialsRequest_ProjectData:
		prefix := scopePrefix
		for _, sub := range msg.GetSubpaths() {
			// Narrowing: each subpath becomes its own prefix under the area.
			// A subpath that climbs out is a bug or an attack; either way it
			// is refused rather than cleaned up quietly.
			if !safeSubpath(sub) {
				return resolvedResource{}, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf(
					"subpath %q is not a relative path inside the resource", sub))
			}
		}
		prefixes := []string{prefix}
		if subs := msg.GetSubpaths(); len(subs) > 0 {
			prefixes = make([]string, 0, len(subs))
			for _, sub := range subs {
				prefixes = append(prefixes, path.Join(prefix, sub))
			}
		}
		return resolvedResource{prefixes: prefixes, scopePrefix: scopePrefix, label: "project-data"}, nil

	case *dataproxy.GetCloudCredentialsRequest_Volume:
		vol, err := s.readVolumeLocator(ctx, msg.GetVolume().GetLocator(), bucket)
		if err != nil {
			return resolvedResource{}, err
		}
		return resolvedResource{
			prefixes:    vol.prefixes,
			scopePrefix: scopePrefix,
			label:       "volume-" + vol.name,
		}, nil
	}

	return resolvedResource{}, connect.NewError(connect.CodeInvalidArgument,
		fmt.Errorf("the request named no resource to get credentials for"))
}

// safeSubpath rejects anything that is not a plain relative path. An absolute
// path or a `..` component would land the policy outside the resource, which
// the scope check would then catch -- but failing here says why.
func safeSubpath(sub string) bool {
	if sub == "" || strings.HasPrefix(sub, "/") {
		return false
	}
	cleaned := path.Clean(sub)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return false
	}
	return true
}

// withinScope reports whether prefix sits inside scope. An empty scope means
// the store has no project layout configured, in which case nothing is inside
// it and every request is refused -- the safe direction.
func withinScope(prefix, scope string) bool {
	if scope == "" {
		return false
	}
	prefix = strings.Trim(prefix, "/")
	scope = strings.Trim(scope, "/")
	return prefix == scope || strings.HasPrefix(prefix, scope+"/")
}

// projectDataPrefix is where a project/domain's data lives inside the bucket.
// It mirrors constructStoragePath, which is what makes a credential for this
// area equivalent to the signed URLs CreateUploadLocation hands out.
func (s *Service) projectDataPrefix(project, domain string) string {
	if project == "" || domain == "" {
		return ""
	}
	return strings.Trim(path.Join(s.cfg.Upload.StoragePrefix, project, domain), "/")
}

// sessionLabel is what lands in the cloud's own audit trail (for AWS, the
// RoleSessionName in every CloudTrail event). Every vended session assumes the
// same role, so without this the audit says only that the data proxy's role did
// something. 64 characters is the STS limit.
func sessionLabel(project, domain, resource string) string {
	label := fmt.Sprintf("%s-%s-%s", project, domain, resource)
	label = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		case r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, label)
	if len(label) > 64 {
		label = label[:64]
	}
	return strings.Trim(label, "-")
}

func (s *Service) credentialDuration(requested time.Duration) time.Duration {
	maxDuration := s.cfg.Credentials.MaxDuration.Duration
	defaultDuration := s.cfg.Credentials.DefaultDuration.Duration
	if requested <= 0 {
		requested = defaultDuration
	}
	if maxDuration > 0 && requested > maxDuration {
		return maxDuration
	}
	return requested
}
