package auth

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/coreos/go-oidc"
	"github.com/stretchr/testify/assert"
)

// Until the go-oidc library uses typed errors, we are left to check expiration with a string match.
// This test ensures that the string we depend on remains in the error message.
func TestExpiredToken(t *testing.T) {
	// #nosec
	expiredToken := "eyJraWQiOiItY2FQXzgyX1o0cVVDMnUtWkRPS2pPYVVIa2RkaWN3YUJOMGJjMTZIb3ZFIiwiYWxnIjoiUlMyNTYifQ.eyJ2ZXIiOjEsImp0aSI6IkFULmRObmh1WTZPMm80QU1sOGNmYVRTWkNqQl9jNW1hYUNZUFZVdzRDZzJOTGcuWlFxMG5Tb0ZDTWl6RUlRUTNlL3gyaXoyY1EwNDdJemgrTDE5UTQ2Q3U4OD0iLCJpc3MiOiJodHRwczovL2x5ZnQub2t0YS5jb20vb2F1dGgyL2RlZmF1bHQiLCJhdWQiOiJhcGk6Ly9kZWZhdWx0IiwiaWF0IjoxNTcwNDAxNTU3LCJleHAiOjE1NzA0MDUxNTcsImNpZCI6IjBvYWJzNTJna3VYNFZ6WXRKMXQ3IiwidWlkIjoiMDB1MXAyMjB2NmRtc0N3Y28xdDciLCJzY3AiOlsib3BlbmlkIiwib2ZmbGluZV9hY2Nlc3MiXSwic3ViIjoieXRvbmdAbHlmdC5jb20ifQ.iTAwHhWZgdu-k6am8QH85Kt2l5duEiKuex5d2l0uAhppvVBKmyL1n2r55xxea5Q9AtQo_dFtUnceqmi6XVlksppZ__NAAG8v6vHzrUpin1G2XW4Ycv2_HMNSCAZCVQ_JCGIX9INxTb1K43sknZo0eMpEaf4Do24MtkJxQEZpCrPyt_qsMVxyRJBIsUcW3wzd8mMyvn5rX-EWIvDuaQYwrW2egCdw6I2IzWjZ923F9S-neHKkuf3E4fVnp8WUYoWs3BRR9LzZQPwDCES10sixLb7v6khopP4bW2L5yooccDp0Sdied08aFw63Uu4vzFRwIu_vEzEhSaDY7KOgccJjcA"

	myClient := &http.Client{}
	ctx := oidc.ClientContext(context.Background(), myClient)
	provider, err := oidc.NewProvider(ctx, "https://lyft.okta.com/oauth2/default")
	if err != nil {
		panic(err)
	}
	var verifier = provider.Verifier(&oidc.Config{ClientID: "api://default"})

	x, err := verifier.Verify(context.Background(), expiredToken)
	assert.Nil(t, x)
	assert.Error(t, err)
	t.Log(err.Error())
	assert.True(t, strings.Contains(err.Error(), "token is expired"))
}
