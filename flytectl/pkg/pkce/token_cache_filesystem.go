package pkce

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/oauth2"

	b64 "encoding/base64"
	"encoding/json"

	f "github.com/flyteorg/flyte/flytectl/pkg/filesystemutils"
)

// tokenCacheFilesystemProvider wraps the logic to save and retrieve tokens from the fs.
type tokenCacheFilesystemProvider struct {
	ServiceUser string

	// credentialsFile is the path to the file where the credentials are stored. This is
	// typically $HOME/.flyte/credentials.json but embedded as a private field for tests.
	credentialsFile string

	mu sync.RWMutex
}

func NewtokenCacheFilesystemProvider(serviceUser string) *tokenCacheFilesystemProvider {
	return &tokenCacheFilesystemProvider{
		ServiceUser:     serviceUser,
		credentialsFile: f.FilePathJoin(f.UserHomeDir(), ".flyte", "credentials.json"),
	}
}

type credentials map[string]*oauth2.Token

func (c credentials) MarshalJSON() ([]byte, error) {
	m := make(map[string]string)
	for k, v := range c {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		m[k] = b64.StdEncoding.EncodeToString(b)
	}
	return json.Marshal(m)
}

func (c credentials) UnmarshalJSON(b []byte) error {
	m := make(map[string]string)
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	for k, v := range m {
		s, err := b64.StdEncoding.DecodeString(v)
		if err != nil {
			return err
		}
		tk := &oauth2.Token{}
		if err = json.Unmarshal(s, tk); err != nil {
			return err
		}
		c[k] = tk
	}
	return nil
}

func (t *tokenCacheFilesystemProvider) SaveToken(token *oauth2.Token) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if token.AccessToken == "" {
		return fmt.Errorf("cannot save empty token with expiration %v", token.Expiry)
	}

	dir := filepath.Dir(t.credentialsFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating base directory (%s) for credentials: %s", dir, err.Error())
	}

	creds, err := t.getExistingCredentials()
	if err != nil {
		return err
	}
	creds[t.ServiceUser] = token

	tmp, err := os.CreateTemp("", "flytectl")
	if err != nil {
		return fmt.Errorf("creating tmp file for credentials update: %s", err.Error())
	}
	defer os.Remove(tmp.Name())

	b, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("marshalling credentials: %s", err.Error())
	}
	if _, err := tmp.Write(b); err != nil {
		return fmt.Errorf("writing updated credentials to tmp file: %s", err.Error())
	}

	if err = os.Rename(tmp.Name(), t.credentialsFile); err != nil {
		return fmt.Errorf("updating credentials via tmp file rename: %s", err.Error())
	}

	return nil
}

func (t *tokenCacheFilesystemProvider) GetToken() (*oauth2.Token, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	creds, err := t.getExistingCredentials()
	if err != nil {
		return nil, err
	}

	if token, ok := creds[t.ServiceUser]; ok {
		return token, nil
	}

	return nil, errors.New("token does not exist")
}

func (t *tokenCacheFilesystemProvider) getExistingCredentials() (credentials, error) {
	creds := credentials{}
	if _, err := os.Stat(t.credentialsFile); errors.Is(err, os.ErrNotExist) {
		return creds, nil
	}

	b, err := os.ReadFile(t.credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("reading existing credentials: %s", err.Error())
	}

	if err = json.Unmarshal(b, &creds); err != nil {
		return nil, fmt.Errorf("unmarshalling credentials: %s", err.Error())
	}

	return creds, nil
}
