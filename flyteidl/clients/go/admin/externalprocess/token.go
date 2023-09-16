package externalprocess

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type jwtClaims struct {
	Exp int64 `json:"exp"`
}

// When we receive a token from an external process, we don't have a key to validate it
// For caching purposes, we still want to know when the token *allegedly* expires.
func GetUnvalidatedTokenExpiration(tokenString string) (time.Time, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return time.Time{}, errors.New("invalid token")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, err
	}

	var claims jwtClaims
	err = json.Unmarshal(payload, &claims)
	if err != nil {
		return time.Time{}, err
	}

	expiry := time.Unix(claims.Exp, 0)

	return expiry, nil
}
