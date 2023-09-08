package pkce

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/flyteorg/flyteidl/clients/go/admin/oauth"
)

func getAuthServerCallbackHandler(c *oauth.Config, codeVerifier string, tokenChannel chan *oauth2.Token,
	errorChannel chan error, stateString string) func(rw http.ResponseWriter, req *http.Request) {

	return func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write([]byte(`<h1>Flyte Authentication</h1>`))
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		if req.URL.Query().Get("error") != "" {
			errorChannel <- fmt.Errorf("error on callback during authorization due to %v", req.URL.Query().Get("error"))
			_, _ = rw.Write([]byte(fmt.Sprintf(`<h1>Error!</h1>
			Error: %s<br>
			Error Hint: %s<br>
			Description: %s<br>
			<br>`,
				req.URL.Query().Get("error"),
				req.URL.Query().Get("error_hint"),
				req.URL.Query().Get("error_description"),
			)))
			return
		}
		if req.URL.Query().Get("code") == "" {
			errorChannel <- fmt.Errorf("could not find the authorize code")
			_, _ = rw.Write([]byte(fmt.Sprintln(`<p>Could not find the authorize code.</p>`)))
			return
		}
		if req.URL.Query().Get("state") != stateString {
			errorChannel <- fmt.Errorf("possibly a csrf attack")
			_, _ = rw.Write([]byte(fmt.Sprintln(`<p>Sorry we can't serve your request'.</p>`)))
			return
		}
		// We'll check whether we sent a code+PKCE request, and if so, send the code_verifier along when requesting the access token.
		var opts []oauth2.AuthCodeOption
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", codeVerifier))

		token, err := c.Exchange(context.Background(), req.URL.Query().Get("code"), opts...)
		if err != nil {
			errorChannel <- fmt.Errorf("error while exchanging auth code due to %v", err)
			_, _ = rw.Write([]byte(fmt.Sprintf(`<p>Couldn't get access token due to error: %s</p>`, err.Error())))
			return
		}
		_, _ = rw.Write([]byte(`<p>Cool! Your authentication was successful and you can close the window.<p>`))
		tokenChannel <- token
	}
}
