package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"

	"github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/logger"

	"io/ioutil"
)

const (
	ErrCertificate errors.ErrorCode = "CERTIFICATE_FAILURE"
)

func GetSslCredentials(ctx context.Context, certFile, keyFile string) (*x509.CertPool, *tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, nil, errors.Wrapf(ErrCertificate, err, "failed to load X509 key pair: %s", certFile)
	}
	logger.Infof(ctx, "Constructing SSL credentials")

	certPool := x509.NewCertPool()
	data, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, nil, errors.Wrapf(ErrCertificate, err, "failed to read server cert file: %s", certFile)
	}
	if ok := certPool.AppendCertsFromPEM(data); !ok {
		return nil, nil, errors.Errorf(ErrCertificate, "failed to load certificate into the pool")
	}

	return certPool, &cert, nil
}
