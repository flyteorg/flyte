package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestGetClaims(t *testing.T) {
	noClaims := map[string]interface{}(nil)
	noClaimsCtx, err := NewIdentityContext("", "", "", time.Now(), nil, nil, nil)
	assert.NoError(t, err)
	assert.EqualValues(t, noClaims, noClaimsCtx.Claims())

	claims := map[string]interface{}{
		"groups":    []string{"g1", "g2"},
		"something": "else",
	}
	withClaimsCtx, err := NewIdentityContext("", "", "", time.Now(), nil, nil, claims)
	assert.NoError(t, err)
	assert.EqualValues(t, claims, withClaimsCtx.Claims())

	assert.NotEmpty(t, withClaimsCtx.UserInfo().AdditionalClaims)
}

func TestWithExecutionUserIdentifier(t *testing.T) {
	idctx, err := NewIdentityContext("", "", "", time.Now(), sets.String{}, nil, nil)
	assert.NoError(t, err)
	newIDCtx := idctx.WithExecutionUserIdentifier("byhsu")
	// make sure the original one is intact
	assert.Equal(t, "", idctx.ExecutionIdentity())
	assert.Equal(t, "byhsu", newIDCtx.ExecutionIdentity())
}
