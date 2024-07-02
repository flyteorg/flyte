package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"

	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// Ref : Taken from oAuth library implementation of expiry
// defaultExpiryDelta determines how earlier a token should be considered
// expired than its actual expiration time. It is used to avoid late
// expirations due to client-server time mismatches.
const defaultExpiryDelta = 10 * time.Second

// Valid reports whether t is non-nil, has an AccessToken, and is not expired.
func Valid(t *oauth2.Token) bool {
	if t == nil || t.AccessToken == "" {
		return false
	}
	expiryDelta := defaultExpiryDelta
	tokenExpiry, err := parseDateClaim(t.AccessToken)
	if err != nil {
		logger.Errorf(context.Background(), "parseDateClaim failed due to %v", err)
		return false
	}
	logger.Debugf(context.Background(), "Token expiry : %v, Access token expiry : %v, Are the equal : %v", t.Expiry, tokenExpiry, tokenExpiry.Equal(t.Expiry))
	return !tokenExpiry.Add(-expiryDelta).Before(time.Now())
}

// parseDateClaim parses the JWT token string and extracts the expiration time
func parseDateClaim(tokenString string) (time.Time, error) {
	// Parse the token
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return time.Time{}, err
	}

	// Extract the claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Get the expiration time
		if exp, ok := claims["exp"].(float64); ok {
			return time.Unix(int64(exp), 0), nil
		}
	}

	return time.Time{}, fmt.Errorf("no expiration claim found in token")
}
