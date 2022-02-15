package cmd

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"

	"github.com/flyteorg/flytepropeller/pkg/webhook"

	webhookConfig "github.com/flyteorg/flytepropeller/pkg/webhook/config"

	"github.com/flyteorg/flytestdlib/logger"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytepropeller/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/spf13/cobra"
)

const (
	CaCertKey            = "ca.crt"
	ServerCertKey        = "tls.crt"
	ServerCertPrivateKey = "tls.key"
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
		return runCertsCmd(context.Background(), config.GetConfig(), webhookConfig.GetConfig())
	},
}

type webhookCerts struct {
	// base64 Encoded CA Cert
	CaPEM *bytes.Buffer
	// base64 Encoded Server Cert
	ServerPEM *bytes.Buffer
	// base64 Encoded Server Cert Key
	PrivateKeyPEM *bytes.Buffer
}

func init() {
	webhookCmd.AddCommand(initCertsCmd)
}

func runCertsCmd(ctx context.Context, propellerCfg *config.Config, cfg *webhookConfig.Config) error {
	podNamespace, found := os.LookupEnv(webhook.PodNamespaceEnvVar)
	if !found {
		podNamespace = podDefaultNamespace
	}

	logger.Infof(ctx, "Issuing certs")
	certs, err := createCerts(podNamespace)
	if err != nil {
		return err
	}

	kubeClient, _, err := utils.GetKubeConfig(ctx, propellerCfg)
	if err != nil {
		return err
	}

	logger.Infof(ctx, "Creating secret [%v] in Namespace [%v]", cfg.SecretName, podNamespace)
	err = createWebhookSecret(ctx, podNamespace, cfg, certs, kubeClient.CoreV1().Secrets(podNamespace))
	if err != nil {
		return err
	}

	return nil
}

func createWebhookSecret(ctx context.Context, namespace string, cfg *webhookConfig.Config, certs webhookCerts, secretsClient v1.SecretInterface) error {
	isImmutable := true
	secretData := map[string][]byte{
		CaCertKey:            certs.CaPEM.Bytes(),
		ServerCertKey:        certs.ServerPEM.Bytes(),
		ServerCertPrivateKey: certs.PrivateKeyPEM.Bytes(),
	}

	secret := &corev1.Secret{
		ObjectMeta: v12.ObjectMeta{
			Name:      cfg.SecretName,
			Namespace: namespace,
		},
		Type:      corev1.SecretTypeOpaque,
		Data:      secretData,
		Immutable: &isImmutable,
	}

	_, err := secretsClient.Create(ctx, secret, v12.CreateOptions{})
	if err == nil {
		logger.Infof(ctx, "Created secret [%v]", cfg.SecretName)
		return nil
	}

	if kubeErrors.IsAlreadyExists(err) {
		logger.Infof(ctx, "A secret already exists with the same name. Validating.")
		s, err := secretsClient.Get(ctx, cfg.SecretName, v12.GetOptions{})
		if err != nil {
			return err
		}

		// If ServerCertKey or ServerCertPrivateKey are missing, update
		requiresUpdate := false
		for key := range secretData {
			if key == CaCertKey {
				continue
			}

			if _, exists := s.Data[key]; !exists {
				requiresUpdate = true
				break
			}
		}

		if requiresUpdate {
			logger.Infof(ctx, "The existing secret is missing one or more keys.")
			secret.Annotations = map[string]string{
				"flyteLastUpdate": "system-updated",
				"flyteUpdatedAt":  time.Now().String(),
			}

			_, err = secretsClient.Update(ctx, secret, v12.UpdateOptions{})
			if err != nil && kubeErrors.IsConflict(err) {
				logger.Infof(ctx, "Another instance of flyteadmin has updated the same secret. Ignoring this update")
				err = nil
			}

			return err
		}

		return nil
	}

	return err
}

func createCerts(serviceNamespace string) (certs webhookCerts, err error) {
	// CA config
	caRequest := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization: []string{"flyte.org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// CA private key
	caPrivateKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return webhookCerts{}, err
	}

	// Self signed CA certificate
	caCert, err := x509.CreateCertificate(cryptorand.Reader, caRequest, caRequest, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return webhookCerts{}, err
	}

	// PEM encode CA cert
	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCert,
	})
	if err != nil {
		return webhookCerts{}, err
	}

	dnsNames := []string{"flyte-pod-webhook",
		"flyte-pod-webhook." + serviceNamespace, "flyte-pod-webhook." + serviceNamespace + ".svc"}
	commonName := "flyte-pod-webhook." + serviceNamespace + ".svc"

	// server cert config
	certRequest := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"flyte.org"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// server private key
	serverPrivateKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return webhookCerts{}, err
	}

	// sign the server cert
	cert, err := x509.CreateCertificate(cryptorand.Reader, certRequest, caRequest, &serverPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return webhookCerts{}, err
	}

	// PEM encode the  server cert and key
	serverCertPEM := new(bytes.Buffer)
	err = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})

	if err != nil {
		return webhookCerts{}, fmt.Errorf("failed to Encode CertPEM. Error: %w", err)
	}

	serverPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(serverPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivateKey),
	})

	if err != nil {
		return webhookCerts{}, fmt.Errorf("failed to Encode Cert Private Key. Error: %w", err)
	}

	return webhookCerts{
		CaPEM:         caPEM,
		ServerPEM:     serverCertPEM,
		PrivateKeyPEM: serverPrivKeyPEM,
	}, nil
}
