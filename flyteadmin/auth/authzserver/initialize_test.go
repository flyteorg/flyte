package authzserver

import (
	"testing"

	"github.com/ory/fosite/storage"

	"github.com/ory/fosite/compose"

	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyteadmin/auth"
	"github.com/flyteorg/flyteadmin/auth/interfaces/mocks"
)

func TestRegisterHandlers(t *testing.T) {
	t.Run("No OAuth2 Provider, no registration required", func(t *testing.T) {
		registerer := &mocks.HandlerRegisterer{}
		RegisterHandlers(registerer, auth.Context{})
	})

	t.Run("Register 4 endpoints", func(t *testing.T) {
		registerer := &mocks.HandlerRegisterer{}
		registerer.On("HandleFunc", "/oauth2/authorize", mock.Anything)
		registerer.On("HandleFunc", "/oauth2/authorize_callback", mock.Anything)
		registerer.On("HandleFunc", "/oauth2/jwks", mock.Anything)
		registerer.On("HandleFunc", "/oauth2/token", mock.Anything)
		authCtx := &mocks.AuthenticationContext{}
		oauth2Provider := &mocks.OAuth2Provider{}
		authCtx.OnOAuth2Provider().Return(oauth2Provider)
		RegisterHandlers(registerer, authCtx)
	})
}

func Test_composeOAuth2Provider(t *testing.T) {
	composeOAuth2Provider(nil, &compose.Config{}, &storage.MemoryStore{}, nil)
}
