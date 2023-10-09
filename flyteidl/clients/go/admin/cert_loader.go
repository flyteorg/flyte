package admin

import (
	"fmt"
	"io/ioutil"

	"crypto/x509"
)

// readCACerts from the passed in file at certLoc and return certpool.
func readCACerts(certLoc string) (*x509.CertPool, error) {
	rootPEM, err := ioutil.ReadFile(certLoc)
	if err != nil {
		return nil, fmt.Errorf("unable to read from %v file due to %v", certLoc, err)
	}
	rootCertPool := x509.NewCertPool()
	ok := rootCertPool.AppendCertsFromPEM(rootPEM)
	if !ok {
		return nil, fmt.Errorf("failed to parse root certificate file %v due to %v", certLoc, err)
	}
	return rootCertPool, err
}
