package externalprocess

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetUnvalidatedTokenExpiration(t *testing.T) {

	exp := time.Now().Unix()
	claims := jwtClaims{Exp: exp}
	payload, _ := json.Marshal(claims)
	token := "header." + base64.RawURLEncoding.EncodeToString(payload) + ".signature"

	expiry, err := GetUnvalidatedTokenExpiration(token)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	assert.Equal(t, exp, expiry.Unix())
}
