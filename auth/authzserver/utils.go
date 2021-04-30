package authzserver

import (
	"context"
	"encoding/base64"
	"net/http"

	"github.com/flyteorg/flyteadmin/auth"
	"github.com/flyteorg/flyteadmin/auth/config"
	"github.com/gtank/cryptopasta"
	"github.com/ory/fosite"
)

func interfaceSliceToStringSlice(raw []interface{}) []string {
	res := make([]string, 0, len(raw))
	for _, item := range raw {
		res = append(res, item.(string))
	}

	return res
}

func toClientIface(clients map[string]*fosite.DefaultClient) map[string]fosite.Client {
	res := make(map[string]fosite.Client, len(clients))
	for clientID, client := range clients {
		res[clientID] = client
	}

	return res
}

func GetIssuer(ctx context.Context, req *http.Request, cfg *config.Config) string {
	if configIssuer := cfg.AppAuth.SelfAuthServer.Issuer; len(configIssuer) > 0 {
		return configIssuer
	}

	return auth.GetPublicURL(ctx, req, cfg).String()
}

func encryptString(plainTextCode string, blockKey [auth.SymmetricKeyLength]byte) (string, error) {
	cypher, err := cryptopasta.Encrypt([]byte(plainTextCode), &blockKey)
	if err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(cypher), nil
}

func decryptString(encryptedEncoded string, blockKey [auth.SymmetricKeyLength]byte) (string, error) {
	cypher, err := base64.RawStdEncoding.DecodeString(encryptedEncoded)
	if err != nil {
		return "", err
	}

	raw, err := cryptopasta.Decrypt(cypher, &blockKey)
	return string(raw), err
}
