package pkce

import (
	"encoding/json"
	"fmt"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

// TokenCacheKeyringProvider wraps the logic to save and retrieve tokens from the OS's keyring implementation.
type TokenCacheKeyringProvider struct {
	ServiceName string
	ServiceUser string
}

const (
	KeyRingServiceUser = "flytectl-user"
	KeyRingServiceName = "flytectl"
)

func (t TokenCacheKeyringProvider) SaveToken(token *oauth2.Token) error {
	var tokenBytes []byte
	if token.AccessToken == "" {
		return fmt.Errorf("cannot save empty token with expiration %v", token.Expiry)
	}

	var err error
	if tokenBytes, err = json.Marshal(token); err != nil {
		return fmt.Errorf("unable to marshal token to save in cache due to %w", err)
	}

	// set token in keyring
	if err = keyring.Set(t.ServiceName, t.ServiceUser, string(tokenBytes)); err != nil {
		return fmt.Errorf("unable to save token. Error: %w", err)
	}

	return nil
}

func (t TokenCacheKeyringProvider) GetToken() (*oauth2.Token, error) {
	// get saved token
	tokenJSON, err := keyring.Get(t.ServiceName, t.ServiceUser)
	if len(tokenJSON) == 0 {
		return nil, fmt.Errorf("no token found in the cache")
	}

	if err != nil {
		return nil, err
	}

	token := oauth2.Token{}
	if err = json.Unmarshal([]byte(tokenJSON), &token); err != nil {
		return nil, fmt.Errorf("unmarshalling error for saved token. Error: %w", err)
	}

	return &token, nil
}
