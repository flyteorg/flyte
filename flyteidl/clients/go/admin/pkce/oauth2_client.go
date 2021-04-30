// Provides the setup required for the client to perform the "Authorization Code" flow with PKCE in order to obtain an
// access token for public/untrusted clients.
package pkce

import (
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/oauth2"
	rand2 "k8s.io/apimachinery/pkg/util/rand"
)

// The following sets up the requirements for generating a standards compliant PKCE code verifier.
const codeVerifierLenMin = 43
const codeVerifierLenMax = 128

// generateCodeVerifier provides an easy way to generate an n-length randomised
// code verifier.
func generateCodeVerifier(n int) string {
	// Enforce standards compliance...
	if n < codeVerifierLenMin {
		n = codeVerifierLenMin
	}

	if n > codeVerifierLenMax {
		n = codeVerifierLenMax
	}

	return randomString(n)
}

func randomString(length int) string {
	return rand2.String(length)
}

// generateCodeChallenge returns a standards compliant PKCE S(HA)256 code
// challenge.
func generateCodeChallenge(codeVerifier string) string {
	// Create a sha-265 hash from the code verifier...
	s256 := sha256.New()
	_, _ = s256.Write([]byte(codeVerifier))
	// Then base64 encode the hash sum to create a code challenge...
	return base64.RawURLEncoding.EncodeToString(s256.Sum(nil))
}

func state(n int) string {
	data := randomString(n)
	return base64.RawURLEncoding.EncodeToString([]byte(data))
}

// SimpleTokenSource defines a simple token source that caches a token in memory.
type SimpleTokenSource struct {
	CachedToken *oauth2.Token
}

func (ts *SimpleTokenSource) Token() (*oauth2.Token, error) {
	t := ts.CachedToken
	return t, nil
}
