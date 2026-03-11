package authzserver

import (
	"encoding/json"
	"fmt"
	"time"

	jwtgo "github.com/golang-jwt/jwt/v5"

	authpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
	auth "github.com/flyteorg/flyte/v2/runs/service/auth"
)

const (
	// ClientIDClaim is the JWT claim key for the client ID.
	ClientIDClaim = "client_id"
	// UserIDClaim is the JWT claim key for user info.
	UserIDClaim = "user_info"
	// ScopeClaim is the JWT claim key for scopes.
	ScopeClaim = "scp"
)

// verifyClaims extracts identity information from raw JWT claims and validates the audience.
func verifyClaims(expectedAudience map[string]bool, claims jwtgo.MapClaims) (*auth.IdentityContext, error) {
	aud, err := claims.GetAudience()
	if err != nil {
		return nil, fmt.Errorf("failed to get audience: %w", err)
	}

	matchedAudience := ""
	for _, a := range aud {
		if expectedAudience[a] {
			matchedAudience = a
			break
		}
	}
	if matchedAudience == "" {
		return nil, fmt.Errorf("invalid audience %v, wanted one of %v", aud, expectedAudience)
	}

	sub, _ := claims.GetSubject()

	issuedAt := time.Time{}
	if iat, err := claims.GetIssuedAt(); err == nil && iat != nil {
		issuedAt = iat.Time
	}

	userInfo := &authpb.UserInfoResponse{}
	if userInfoClaim, found := claims[UserIDClaim]; found && userInfoClaim != nil {
		if userInfoRaw, ok := userInfoClaim.(map[string]interface{}); ok {
			raw, err := json.Marshal(userInfoRaw)
			if err != nil {
				return nil, err
			}
			if err = json.Unmarshal(raw, userInfo); err != nil {
				return nil, fmt.Errorf("failed to unmarshal user info claim: %w", err)
			}
		}
	}

	clientID := ""
	if clientIDClaim, found := claims[ClientIDClaim]; found {
		if s, ok := clientIDClaim.(string); ok {
			clientID = s
		}
	}

	var scopes []string
	if scopesClaim, found := claims[ScopeClaim]; found {
		switch sct := scopesClaim.(type) {
		case []interface{}:
			scopes = interfaceSliceToStringSlice(sct)
		case string:
			scopes = []string{sct}
		default:
			return nil, fmt.Errorf("failed getting scope claims due to unknown type %T with value %v", sct, sct)
		}
	}

	// In some cases, "user_info" field doesn't exist in the raw claim,
	// but we can get email from "email" field.
	if emailClaim, found := claims["email"]; found {
		if email, ok := emailClaim.(string); ok {
			userInfo.Email = email
		}
	}

	// If this is a user-only access token with no scopes defined then add `all` scope by default because it's equivalent
	// to having a user's login cookie or an ID Token as means of accessing the service.
	if len(clientID) == 0 && len(scopes) == 0 {
		scopes = []string{auth.ScopeAll}
	}

	return auth.NewIdentityContext(matchedAudience, sub, clientID, issuedAt, scopes, userInfo, claims), nil
}

func interfaceSliceToStringSlice(raw []interface{}) []string {
	res := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok {
			res = append(res, s)
		}
	}
	return res
}
