package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func GenTokenWithCustomExpiry(t *testing.T, expiry time.Time) *oauth2.Token {
	var signingKey = []byte("your_secret_key")
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = expiry.Unix()
	tokenString, err := token.SignedString(signingKey)
	assert.NoError(t, err)
	return &oauth2.Token{
		AccessToken: tokenString,
		Expiry:      expiry,
		TokenType:   "bearer",
	}
}
