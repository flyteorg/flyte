package cmd

import (
	"context"

	"github.com/flyteorg/flytepropeller/pkg/controller/config"

	webhookConfig "github.com/flyteorg/flytepropeller/pkg/webhook/config"

	"github.com/flyteorg/flytepropeller/pkg/webhook"

	"github.com/spf13/cobra"
)

// initCertsCmd initializes x509 TLS Certificates and saves them to a secret.
var initCertsCmd = &cobra.Command{
	Use:     "init-certs",
	Aliases: []string{"init-cert"},
	Short:   "Generates CA, Cert and cert key and saves them into a secret using the configured --webhook.secretName",
	Long: `
K8s API Server Webhooks' Services are required to serve traffic over SSL. Ideally the SSL certificate is issued by a 
known Certificate Authority that's already trusted by API Server (e.g. using Let's Encrypt Cluster Certificate Controller).
Otherwise, a self-issued certificate can be used to serve traffic as long as the CA for that certificate is stored in
the MutatingWebhookConfiguration object that registers the webhook with API Server.

init-certs generates 4096-bit X509 Certificates for Certificate Authority as well as a derived Server cert and its private
key. It serializes all of them in PEM format (base64 encoding-compatible) and stores them into a new kubernetes secret.
If a secret with the same name already exist, it'll not update it.

POD_NAMESPACE is an environment variable that can, optionally, be set on the Pod that runs init-certs command. If set, 
the secret will be created in that namespace. It's critical that this command creates the secret in the right namespace
for the Webhook command to mount and read correctly.
`,
	Example: "flytepropeller webhook init-certs",
	RunE: func(cmd *cobra.Command, args []string) error {
		return webhook.InitCerts(context.Background(), config.GetConfig(), webhookConfig.GetConfig())
	},
}

func init() {
	webhookCmd.AddCommand(initCertsCmd)
}
