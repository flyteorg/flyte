package authzserver

import (
	"encoding/json"
	"fmt"

	"github.com/ory/x/jwtx"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyteadmin/auth"
	"github.com/flyteorg/flyteadmin/auth/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
)

func verifyClaims(expectedAudience sets.String, claimsRaw map[string]interface{}) (interfaces.IdentityContext, error) {
	claims := jwtx.ParseMapStringInterfaceClaims(claimsRaw)

	foundAudIndex := -1
	for audIndex, aud := range claims.Audience {
		if expectedAudience.Has(aud) {
			foundAudIndex = audIndex
			break
		}
	}

	if foundAudIndex < 0 {
		return nil, fmt.Errorf("invalid audience [%v]", claims)
	}

	userInfo := &service.UserInfoResponse{}
	if userInfoClaim, found := claimsRaw[UserIDClaim]; found && userInfoClaim != nil {
		userInfoRaw := userInfoClaim.(map[string]interface{})
		raw, err := json.Marshal(userInfoRaw)
		if err != nil {
			return nil, err
		}

		if err = json.Unmarshal(raw, userInfo); err != nil {
			return nil, fmt.Errorf("failed to unmarshal user info claim into UserInfo type. Error: %w", err)
		}
	}

	clientID := ""
	if clientIDClaim, found := claimsRaw[ClientIDClaim]; found {
		clientID = clientIDClaim.(string)
	}

	scopes := sets.NewString()
	if scopesClaim, found := claimsRaw[ScopeClaim]; found {

		switch sct := scopesClaim.(type) {
		case []interface{}:
			scopes = sets.NewString(interfaceSliceToStringSlice(sct)...)
		case string:
			sets.NewString(fmt.Sprintf("%v", scopesClaim))
		default:
			return nil, fmt.Errorf("failed getting scope claims due to  unknown type %T with value %v", sct, sct)
		}
	}

	EmailKey := "email"
	// In some cases, "user_info" field doesn't exist in the raw claim,
	// but we can get email from "email" field
	if emailClaim, found := claimsRaw[EmailKey]; found {
		email := emailClaim.(string)
		userInfo.Email = email
	}
	// If this is a user-only access token with no scopes defined then add `all` scope by default because it's equivalent
	// to having a user's login cookie or an ID Token as means of accessing the service.
	if len(clientID) == 0 && scopes.Len() == 0 {
		scopes.Insert(auth.ScopeAll)
	}

	return auth.NewIdentityContext(claims.Audience[foundAudIndex], claims.Subject, clientID, claims.IssuedAt, scopes, userInfo, claimsRaw)
}
