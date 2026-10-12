package webhook

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptorand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	webhookConfig "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
)

const (
	testServiceName = "flyte-pod-webhook"
	testNamespace   = "flyte"
	testSecretName  = "flyte-pod-webhook"
	testCertOrg     = "flyte.org"
)

// makeTestCerts issues a CA + server cert quickly (ECDSA) with the given DNS names and expiry, so
// tests can build valid, expired, or mis-named cert sets without 4096-bit RSA keygen.
func makeTestCerts(t *testing.T, dnsNames []string, notAfter time.Time) webhookCerts {
	t.Helper()
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), cryptorand.Reader)
	require.NoError(t, err)
	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{Organization: []string{testCertOrg}},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caDER, err := x509.CreateCertificate(cryptorand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	require.NoError(t, err)

	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), cryptorand.Reader)
	require.NoError(t, err)
	serverTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		DNSNames:     dnsNames,
		NotBefore:    time.Now().Add(-2 * time.Hour),
		NotAfter:     notAfter,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	serverDER, err := x509.CreateCertificate(cryptorand.Reader, serverTmpl, caTmpl, &serverKey.PublicKey, caKey)
	require.NoError(t, err)
	keyDER, err := x509.MarshalECPrivateKey(serverKey)
	require.NoError(t, err)

	encode := func(typ string, b []byte) *bytes.Buffer {
		buf := new(bytes.Buffer)
		require.NoError(t, pem.Encode(buf, &pem.Block{Type: typ, Bytes: b}))
		return buf
	}
	return webhookCerts{
		CaPEM:         encode("CERTIFICATE", caDER),
		ServerPEM:     encode("CERTIFICATE", serverDER),
		PrivateKeyPEM: encode("EC PRIVATE KEY", keyDER),
	}
}

func certData(c webhookCerts) map[string][]byte {
	return map[string][]byte{
		CaCertKey:            c.CaPEM.Bytes(),
		ServerCertKey:        c.ServerPEM.Bytes(),
		ServerCertPrivateKey: c.PrivateKeyPEM.Bytes(),
	}
}

func certSecret(data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: testSecretName, Namespace: testNamespace, UID: "existing-uid"},
		Data:       data,
	}
}

func testCfg(t *testing.T, localCert bool) *webhookConfig.Config {
	return &webhookConfig.Config{
		ServiceName: testServiceName,
		SecretName:  testSecretName,
		CertDir:     t.TempDir(),
		LocalCert:   localCert,
	}
}

func assertLocalCertsMatch(t *testing.T, dir string, want map[string][]byte) {
	t.Helper()
	for _, key := range []string{CaCertKey, ServerCertKey, ServerCertPrivateKey} {
		got, err := os.ReadFile(filepath.Join(dir, key))
		require.NoError(t, err)
		assert.Equal(t, want[key], got, "local %s must match the Secret", key)
	}
}

func getSecret(t *testing.T, client *fake.Clientset) *corev1.Secret {
	t.Helper()
	s, err := client.CoreV1().Secrets(testNamespace).Get(context.Background(), testSecretName, metav1.GetOptions{})
	require.NoError(t, err)
	return s
}

func TestInitCerts_CreatesSecretWhenMissing(t *testing.T) {
	// Uses the real generator once, which also proves createCerts output passes validation.
	ctx := context.Background()
	client := fake.NewClientset()
	cfg := testCfg(t, true)

	require.NoError(t, InitCerts(ctx, client, cfg, testNamespace))

	s := getSecret(t, client)
	require.NotNil(t, s.Immutable)
	assert.True(t, *s.Immutable)
	assert.Empty(t, validateCertData(s.Data, testServiceName, testNamespace, time.Now()))
	assertLocalCertsMatch(t, cfg.CertDir, s.Data)

	// A second start (e.g. pod restart, another replica) keeps the same certs.
	require.NoError(t, InitCerts(ctx, client, cfg, testNamespace))
	assert.Equal(t, s.Data, getSecret(t, client).Data)
}

func TestEnsureWebhookSecret(t *testing.T) {
	validDNS := serviceDNSNames(testServiceName, testNamespace)
	farFuture := time.Now().AddDate(5, 0, 0)

	tests := []struct {
		name         string
		existing     map[string][]byte // nil = no Secret
		wantKeep     bool
		wantGenerate bool
	}{
		{
			name:         "no secret: generate",
			existing:     nil,
			wantGenerate: true,
		},
		{
			name:     "valid secret: keep without generating",
			existing: certData(makeTestCerts(t, validDNS, farFuture)),
			wantKeep: true,
		},
		{
			name: "missing server key: regenerate",
			existing: func() map[string][]byte {
				d := certData(makeTestCerts(t, validDNS, farFuture))
				delete(d, ServerCertPrivateKey)
				return d
			}(),
			wantGenerate: true,
		},
		{
			name: "missing CA: regenerate",
			existing: func() map[string][]byte {
				d := certData(makeTestCerts(t, validDNS, farFuture))
				delete(d, CaCertKey)
				return d
			}(),
			wantGenerate: true,
		},
		{
			name:         "expired server cert: regenerate",
			existing:     certData(makeTestCerts(t, validDNS, time.Now().Add(-time.Hour))),
			wantGenerate: true,
		},
		{
			name:         "server cert inside renewal window: regenerate",
			existing:     certData(makeTestCerts(t, validDNS, time.Now().Add(certRenewalWindow/2))),
			wantGenerate: true,
		},
		{
			name:         "cert for another namespace: regenerate",
			existing:     certData(makeTestCerts(t, serviceDNSNames(testServiceName, "other"), farFuture)),
			wantGenerate: true,
		},
		{
			name: "server cert from a different CA: regenerate",
			existing: func() map[string][]byte {
				d := certData(makeTestCerts(t, validDNS, farFuture))
				d[CaCertKey] = makeTestCerts(t, validDNS, farFuture).CaPEM.Bytes()
				return d
			}(),
			wantGenerate: true,
		},
		{
			name: "key does not match cert: regenerate",
			existing: func() map[string][]byte {
				d := certData(makeTestCerts(t, validDNS, farFuture))
				d[ServerCertPrivateKey] = makeTestCerts(t, validDNS, farFuture).PrivateKeyPEM.Bytes()
				return d
			}(),
			wantGenerate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var objs []runtime.Object
			if tt.existing != nil {
				objs = append(objs, certSecret(tt.existing))
			}
			client := fake.NewClientset(objs...)
			cfg := testCfg(t, false)

			fresh := makeTestCerts(t, validDNS, farFuture)
			generated := false
			generate := func() (webhookCerts, error) {
				generated = true
				return fresh, nil
			}

			data, err := ensureWebhookSecret(ctx, testNamespace, cfg, client.CoreV1().Secrets(testNamespace), generate)
			require.NoError(t, err)
			assert.Equal(t, tt.wantGenerate, generated)

			stored := getSecret(t, client).Data
			assert.Equal(t, stored, data, "returned data must match the stored Secret")
			if tt.wantKeep {
				assert.Equal(t, tt.existing, stored)
			} else {
				assert.Equal(t, certData(fresh), stored)
			}
		})
	}
}

func TestInitCerts_LocalCertWritesExistingSecret(t *testing.T) {
	// With LocalCert, the files on disk must be the Secret's certs (which the CA bundle and other
	// replicas use), not a freshly generated set.
	existing := certData(makeTestCerts(t, serviceDNSNames(testServiceName, testNamespace), time.Now().AddDate(5, 0, 0)))
	client := fake.NewClientset(certSecret(existing))
	cfg := testCfg(t, true)

	require.NoError(t, InitCerts(context.Background(), client, cfg, testNamespace))
	assert.Equal(t, existing, getSecret(t, client).Data)
	assertLocalCertsMatch(t, cfg.CertDir, existing)
}

func TestEnsureWebhookSecret_AdoptsConcurrentlyCreatedSecret(t *testing.T) {
	validDNS := serviceDNSNames(testServiceName, testNamespace)
	winner := certData(makeTestCerts(t, validDNS, time.Now().AddDate(5, 0, 0)))
	client := fake.NewClientset(certSecret(winner))

	// The first Get misses (another replica hasn't created it yet), so our Create races and
	// loses with AlreadyExists.
	firstGet := true
	client.PrependReactor("get", "secrets", func(k8stesting.Action) (bool, runtime.Object, error) {
		if firstGet {
			firstGet = false
			return true, nil, kubeErrors.NewNotFound(corev1.Resource("secrets"), testSecretName)
		}
		return false, nil, nil
	})

	generate := func() (webhookCerts, error) { return makeTestCerts(t, validDNS, time.Now().AddDate(5, 0, 0)), nil }
	data, err := ensureWebhookSecret(context.Background(), testNamespace, testCfg(t, false),
		client.CoreV1().Secrets(testNamespace), generate)
	require.NoError(t, err)
	assert.Equal(t, winner, data)
	assert.Equal(t, winner, getSecret(t, client).Data)
}

func TestEnsureWebhookSecret_GetErrorIsReturned(t *testing.T) {
	client := fake.NewClientset()
	client.PrependReactor("get", "secrets", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, kubeErrors.NewForbidden(corev1.Resource("secrets"), testSecretName, errors.New("denied"))
	})
	_, err := ensureWebhookSecret(context.Background(), testNamespace, testCfg(t, false),
		client.CoreV1().Secrets(testNamespace), func() (webhookCerts, error) {
			t.Fatal("must not generate when the Secret state is unknown")
			return webhookCerts{}, nil
		})
	assert.Error(t, err)
}
