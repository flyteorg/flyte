package pkce

import (
	"context"
	// This import is used to embed the callback.html
	_ "embed"
	"fmt"
	"net/http"
	"text/template"

	"golang.org/x/oauth2"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/oauth"
)

//go:embed callback.html
var callbackHTML string

var callbackTemplate = template.Must(template.New("callback").Parse(callbackHTML))

type callbackData struct {
	Error            string
	ErrorHint        string
	ErrorDescription string
	NoCode           bool
	WrongState       bool
	AccessTokenError string
}

func getAuthServerCallbackHandler(c *oauth.Config, codeVerifier string, tokenChannel chan *oauth2.Token,
	errorChannel chan error, stateString string, client *http.Client) func(rw http.ResponseWriter, req *http.Request) {
	// this will be used only in Union CLI tools since it's a Union private Flyte fork
	return func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := callbackData{
			Error:            req.URL.Query().Get("error"),
			ErrorHint:        req.URL.Query().Get("error_hint"),
			ErrorDescription: req.URL.Query().Get("error_description"),
			NoCode:           req.URL.Query().Get("code") == "",
			WrongState:       req.URL.Query().Get("state") != stateString,
		}
		defer func() {
			_ = callbackTemplate.Execute(rw, data)
		}()

		if data.Error != "" {
			errorChannel <- fmt.Errorf("error on callback during authorization due to %v", data.Error)
			return
		}
		if data.NoCode {
			errorChannel <- fmt.Errorf("could not find the authorize code")
			return
		}
		if data.WrongState {
			errorChannel <- fmt.Errorf("possibly a csrf attack")
			return
		}

		// We'll check whether we sent a code+PKCE request, and if so, send the code_verifier along when requesting the access token.
		var opts []oauth2.AuthCodeOption
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
		ctx := context.WithValue(context.Background(), oauth2.HTTPClient, client)

		token, err := c.Exchange(ctx, req.URL.Query().Get("code"), opts...)
		if err != nil {
			errorChannel <- fmt.Errorf("error while exchanging auth code due to %v", err)
			data.AccessTokenError = err.Error()
			return
		}

		tokenChannel <- token
	}
}
