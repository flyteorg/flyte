package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetClaims(t *testing.T) {
	noClaims := map[string]interface{}(nil)
	noClaimsCtx := NewIdentityContext("", "", "", time.Now(), nil, nil, nil)
	assert.EqualValues(t, noClaims, noClaimsCtx.Claims())

	claims := map[string]interface{}{
		"groups":    []string{"g1", "g2"},
		"something": "else",
	}
	withClaimsCtx := NewIdentityContext("", "", "", time.Now(), nil, nil, claims)
	assert.EqualValues(t, claims, withClaimsCtx.Claims())
}
