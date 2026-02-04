package secret

import (
	"crypto/md5" // #nosec
	"fmt"
)

// A Kubernetes Secret object's name must be a valid DNS subdomain, which prohibits the use of "_".
// However, Union uses "__" to delimit contextual information.
// There are no restrictions on key names within a Kubernetes Secret object,
// so we can store the original secret name as a key inside the Secret.
// To ensure the Kubernetes Secret name complies with DNS subdomain rules,
// we can use a hash of the original secret name.
// Since this is just for avoid Kubernetes naming restrictions and not for security,
// we can use a simple hash function like MD5.
func EncodeK8sSecretName(secretName string) string {
	hash := md5.Sum([]byte(secretName)) // #nosec
	hashHex := fmt.Sprintf("%x", hash)
	return hashHex
}
