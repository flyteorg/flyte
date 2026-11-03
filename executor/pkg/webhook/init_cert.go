package webhook

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path"
	"time"

	corev1 "k8s.io/api/core/v1"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	webhookConfig "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

type webhookCerts struct {
	CaPEM         *bytes.Buffer
	ServerPEM     *bytes.Buffer
	PrivateKeyPEM *bytes.Buffer
}

const (
	CaCertKey            = "ca.crt"
	ServerCertKey        = "tls.crt"
	ServerCertPrivateKey = "tls.key"
	permission           = 0644
	folderPerm           = 0755
)

// certRenewalWindow is how close to expiry an existing server cert may get before InitCerts
// replaces it. Generated certs are valid for 99 years, so in practice this only triggers for
// certs issued by something else.
const certRenewalWindow = 30 * 24 * time.Hour

// InitCerts makes sure the webhook's TLS cert Secret exists and is usable, and (with LocalCert)
// writes its contents to the cert dir.
//
// An existing Secret is kept as-is when it holds a CA, server cert and key that are mutually
// consistent, unexpired, and valid for the service's DNS names. Keeping it matters: the CA bundle
// in the MutatingWebhookConfiguration and the certs mounted into other running webhook replicas
// come from this Secret, so replacing it on every start would break admission until everything
// converges. Only a missing, incomplete or unusable Secret is (re)generated.
//
// podNamespace must be the namespace the webhook service runs in — the cert's DNS
// names are derived from it.
func InitCerts(ctx context.Context, kubeClient kubernetes.Interface, cfg *webhookConfig.Config, podNamespace string) error {
	generate := func() (webhookCerts, error) {
		logger.Infof(ctx, "Issuing certs")
		return createCerts(cfg.ServiceName, podNamespace)
	}

	logger.Infof(ctx, "Ensuring secret [%v] in Namespace [%v]", cfg.SecretName, podNamespace)
	data, err := ensureWebhookSecret(ctx, podNamespace, cfg, kubeClient.CoreV1().Secrets(podNamespace), generate)
	if err != nil {
		return err
	}

	// TODO(alex): This LocalCert tag is only for flyte running in single binary mode.
	// In full deployment the webhook should be running in a single pod and an init container will generate and inject the secret data
	if cfg.LocalCert {
		return writeLocalCerts(cfg.ExpandCertDir(), data)
	}
	return nil
}

// ensureWebhookSecret returns the cert data the webhook should serve with, creating or replacing
// the Secret only when the existing one is missing or unusable. The returned data always matches
// what is stored in the Secret, so local cert files and the CA bundle never diverge from it.
func ensureWebhookSecret(ctx context.Context, namespace string, cfg *webhookConfig.Config,
	secretsClient v1.SecretInterface, generate func() (webhookCerts, error)) (map[string][]byte, error) {

	serviceName := cfg.ServiceName
	existing, err := secretsClient.Get(ctx, cfg.SecretName, metav1.GetOptions{})
	switch {
	case err == nil:
		reason := validateCertData(existing.Data, serviceName, namespace, time.Now())
		if reason == "" {
			logger.Infof(ctx, "Secret [%v] already exists with valid certs; keeping it.", cfg.SecretName)
			return existing.Data, nil
		}
		logger.Infof(ctx, "Secret [%v] exists but %s; regenerating.", cfg.SecretName, reason)
	case kubeErrors.IsNotFound(err):
		existing = nil
	default:
		return nil, fmt.Errorf("failed to get secret [%v]: %w", cfg.SecretName, err)
	}

	certs, err := generate()
	if err != nil {
		return nil, err
	}
	isImmutable := true
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.SecretName,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			CaCertKey:            certs.CaPEM.Bytes(),
			ServerCertKey:        certs.ServerPEM.Bytes(),
			ServerCertPrivateKey: certs.PrivateKeyPEM.Bytes(),
		},
		Immutable: &isImmutable,
	}

	if existing != nil {
		// The Secret is immutable, so it must be replaced rather than updated. Preconditions make
		// sure we only delete the exact object we judged unusable; if another replica replaced it
		// in the meantime, adopt theirs instead.
		uid := existing.UID
		err := secretsClient.Delete(ctx, cfg.SecretName, metav1.DeleteOptions{
			Preconditions: &metav1.Preconditions{UID: &uid},
		})
		if err != nil && !kubeErrors.IsNotFound(err) && !kubeErrors.IsConflict(err) {
			return nil, fmt.Errorf("failed to delete secret [%v]: %w", cfg.SecretName, err)
		}
	}

	_, err = secretsClient.Create(ctx, secret, metav1.CreateOptions{})
	if err == nil {
		logger.Infof(ctx, "Created secret [%v]", cfg.SecretName)
		return secret.Data, nil
	}
	if !kubeErrors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("failed to create secret [%v]: %w", cfg.SecretName, err)
	}

	// Another replica created it first. Use theirs so every replica serves the same certs.
	logger.Infof(ctx, "Secret [%v] was created concurrently; adopting it.", cfg.SecretName)
	winner, err := secretsClient.Get(ctx, cfg.SecretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get concurrently created secret [%v]: %w", cfg.SecretName, err)
	}
	if reason := validateCertData(winner.Data, serviceName, namespace, time.Now()); reason != "" {
		return nil, fmt.Errorf("concurrently created secret [%v] is unusable: %s", cfg.SecretName, reason)
	}
	return winner.Data, nil
}

// validateCertData returns "" when data holds a usable webhook cert set, or a short reason why
// it does not. Usable means: all three keys present, the key matches the server cert, the server
// cert chains to the CA, is valid for every service DNS name, and is not within
// certRenewalWindow of expiry.
func validateCertData(data map[string][]byte, serviceName, namespace string, now time.Time) string {
	for _, key := range []string{CaCertKey, ServerCertKey, ServerCertPrivateKey} {
		if len(data[key]) == 0 {
			return fmt.Sprintf("is missing key %q", key)
		}
	}

	keyPair, err := tls.X509KeyPair(data[ServerCertKey], data[ServerCertPrivateKey])
	if err != nil {
		return fmt.Sprintf("has a server cert/key that do not form a valid pair (%v)", err)
	}
	serverCert, err := x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return fmt.Sprintf("has an unparsable server cert (%v)", err)
	}

	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(data[CaCertKey]) {
		return "has an unparsable CA cert"
	}

	if now.Add(certRenewalWindow).After(serverCert.NotAfter) {
		return fmt.Sprintf("has a server cert expiring at %v", serverCert.NotAfter)
	}

	for _, dnsName := range serviceDNSNames(serviceName, namespace) {
		if _, err := serverCert.Verify(x509.VerifyOptions{
			DNSName:     dnsName,
			Roots:       roots,
			CurrentTime: now,
			KeyUsages:   []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		}); err != nil {
			return fmt.Sprintf("has a server cert not valid for %q (%v)", dnsName, err)
		}
	}
	return ""
}

func serviceDNSNames(serviceName, namespace string) []string {
	return []string{
		serviceName,
		serviceName + "." + namespace,
		serviceName + "." + namespace + ".svc",
	}
}

func writeLocalCerts(certPath string, data map[string][]byte) error {
	if err := os.MkdirAll(certPath, folderPerm); err != nil {
		return err
	}
	for _, key := range []string{CaCertKey, ServerCertKey, ServerCertPrivateKey} {
		if err := os.WriteFile(path.Join(certPath, key), data[key], permission); err != nil {
			return err
		}
	}
	return nil
}

func createCerts(serviceName string, serviceNamespace string) (certs webhookCerts, err error) {
	caRequest := &x509.Certificate{
		SerialNumber:          big.NewInt(2021),
		Subject:               pkix.Name{Organization: []string{"flyte.org"}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(99, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivateKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return webhookCerts{}, err
	}

	caCert, err := x509.CreateCertificate(cryptorand.Reader, caRequest, caRequest, &caPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return webhookCerts{}, err
	}

	caPEM := new(bytes.Buffer)
	if err = pem.Encode(caPEM, &pem.Block{Type: "CERTIFICATE", Bytes: caCert}); err != nil {
		return webhookCerts{}, err
	}

	dnsNames := serviceDNSNames(serviceName, serviceNamespace)
	commonName := serviceName + "." + serviceNamespace + ".svc"

	certRequest := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(1658),
		Subject:      pkix.Name{CommonName: commonName, Organization: []string{"flyte.org"}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(99, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	serverPrivateKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return webhookCerts{}, err
	}

	cert, err := x509.CreateCertificate(cryptorand.Reader, certRequest, caRequest, &serverPrivateKey.PublicKey, caPrivateKey)
	if err != nil {
		return webhookCerts{}, err
	}

	serverCertPEM := new(bytes.Buffer)
	if err = pem.Encode(serverCertPEM, &pem.Block{Type: "CERTIFICATE", Bytes: cert}); err != nil {
		return webhookCerts{}, fmt.Errorf("failed to encode CertPEM: %w", err)
	}

	serverPrivKeyPEM := new(bytes.Buffer)
	if err = pem.Encode(serverPrivKeyPEM, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverPrivateKey)}); err != nil {
		return webhookCerts{}, fmt.Errorf("failed to encode cert private key: %w", err)
	}

	return webhookCerts{
		CaPEM:         caPEM,
		ServerPEM:     serverCertPEM,
		PrivateKeyPEM: serverPrivKeyPEM,
	}, nil
}
