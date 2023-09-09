package webhook

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path"
	"time"

	"github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytepropeller/pkg/utils"
	webhookConfig "github.com/flyteorg/flytepropeller/pkg/webhook/config"
	"github.com/flyteorg/flytestdlib/logger"
	corev1 "k8s.io/api/core/v1"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type webhookCerts struct {
	// base64 Encoded CA Cert
	CaPEM *bytes.Buffer
	// base64 Encoded Server Cert
	ServerPEM *bytes.Buffer
	// base64 Encoded Server Cert Key
	PrivateKeyPEM *bytes.Buffer
}

const (
	CaCertKey            = "ca.crt"
	ServerCertKey        = "tls.crt"
	ServerCertPrivateKey = "tls.key"
	podDefaultNamespace  = "flyte"
	permission           = 0644
	folderPerm           = 0755
)

func InitCerts(ctx context.Context, propellerCfg *config.Config, cfg *webhookConfig.Config) error {
	podNamespace, found := os.LookupEnv(PodNamespaceEnvVar)
	if !found {
		podNamespace = podDefaultNamespace
	}

	logger.Infof(ctx, "Issuing certs")
	certs, err := createCerts(cfg.ServiceName, podNamespace)
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

	if cfg.LocalCert {
		if _, err := os.Stat(cfg.CertDir); os.IsNotExist(err) {
			if err := os.Mkdir(cfg.CertDir, folderPerm); err != nil {
				return err
			}
		}

		if err := os.WriteFile(path.Join(cfg.CertDir, CaCertKey), certs.CaPEM.Bytes(), permission); err != nil {
			return err
		}

		if err := os.WriteFile(path.Join(cfg.CertDir, ServerCertKey), certs.ServerPEM.Bytes(), permission); err != nil {
			return err
		}

		if err := os.WriteFile(path.Join(cfg.CertDir, ServerCertPrivateKey), certs.PrivateKeyPEM.Bytes(), permission); err != nil {
			return err
		}
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.SecretName,
			Namespace: namespace,
		},
		Type:      corev1.SecretTypeOpaque,
		Data:      secretData,
		Immutable: &isImmutable,
	}

	_, err := secretsClient.Create(ctx, secret, metav1.CreateOptions{})
	if err == nil {
		logger.Infof(ctx, "Created secret [%v]", cfg.SecretName)
		return nil
	}

	if kubeErrors.IsAlreadyExists(err) {
		logger.Infof(ctx, "A secret already exists with the same name. Validating.")
		s, err := secretsClient.Get(ctx, cfg.SecretName, metav1.GetOptions{})
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

			_, err = secretsClient.Update(ctx, secret, metav1.UpdateOptions{})
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

func createCerts(serviceName string, serviceNamespace string) (certs webhookCerts, err error) {
	// CA config
	caRequest := &x509.Certificate{
		SerialNumber: big.NewInt(2021),
		Subject: pkix.Name{
			Organization: []string{"flyte.org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(99, 0, 0),
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

	dnsNames := []string{serviceName,
		serviceName + "." + serviceNamespace, serviceName + "." + serviceNamespace + ".svc"}
	commonName := serviceName + "." + serviceNamespace + ".svc"

	// server cert config
	certRequest := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"flyte.org"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(99, 0, 0),
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
