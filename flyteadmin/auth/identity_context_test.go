package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
