package service

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth/authconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/runs/config"
)

// errNoIdentity is returned when a request carries no resolvable caller identity.
var errNoIdentity = errors.New("no authenticated identity on request")

// IdentityService implements the IdentityServiceHandler interface.
type IdentityService struct {
	// enricher fills name/email from the OIDC userinfo endpoint on the Bearer path,
	// where the access token carries only a subject. nil when no auth server is set.
	enricher *identityEnricher

	// trustHeaders gates resolving the caller from proxy-forwarded auth headers.
	// See Config.TrustForwardedIdentityHeaders.
	trustHeaders bool

	// identityHeaders names the proxy-forwarded headers identity is read from.
	identityHeaders config.IdentityHeadersConfig
}

// NewIdentityService creates a new IdentityService instance. It takes the same
// identity settings as RunService so whoami reports exactly the identity that would
// be recorded as a run's executed_by.
func NewIdentityService(
	authServerBaseURL string,
	trustForwardedIdentityHeaders bool,
	identityHeaders config.IdentityHeadersConfig,
) *IdentityService {
	return &IdentityService{
		enricher:        newIdentityEnricher(authServerBaseURL),
		trustHeaders:    trustForwardedIdentityHeaders,
		identityHeaders: identityHeaders,
	}
}

var _ authconnect.IdentityServiceHandler = (*IdentityService)(nil)

// UserInfo returns information about the currently logged in user, resolved from the
// auth headers the proxy forwards — the same path run attribution uses, so whoami and
// a run's executed_by can never disagree.
//
// Returns Unauthenticated when the request carries no identity, or when forwarded
// headers aren't trusted (in which case the caller cannot be established at all), so
// `flyte whoami` reports not-logged-in rather than printing an empty record.
func (s *IdentityService) UserInfo(
	ctx context.Context,
	req *connect.Request[auth.UserInfoRequest],
) (*connect.Response[auth.UserInfoResponse], error) {
	if !s.trustHeaders {
		return nil, connect.NewError(connect.CodeUnauthenticated, errNoIdentity)
	}

	id := identityFromHeaders(req.Header(), s.identityHeaders)
	// On the Bearer path the token holds only a subject; userinfo supplies name/email.
	id = s.enricher.enrich(ctx, bearerToken(req.Header()), id)

	user := id.GetUser()
	if user == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errNoIdentity)
	}
	return connect.NewResponse(userInfoFromUser(user)), nil
}

// userInfoFromUser maps the internal identity onto the OIDC-shaped UserInfoResponse.
// name is derived from the first/last pair since UserSpec carries no full-name field.
func userInfoFromUser(user *common.User) *auth.UserInfoResponse {
	spec := user.GetSpec()
	return &auth.UserInfoResponse{
		Subject:           user.GetId().GetSubject(),
		Name:              strings.TrimSpace(spec.GetFirstName() + " " + spec.GetLastName()),
		PreferredUsername: spec.GetUserHandle(),
		GivenName:         spec.GetFirstName(),
		FamilyName:        spec.GetLastName(),
		Email:             spec.GetEmail(),
		Picture:           spec.GetPhotoUrl(),
	}
}
