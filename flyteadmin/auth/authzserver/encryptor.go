package authzserver

type Encryptor interface {
	Encrypt(raw string) (cypher string, err error)
	Decrypt(cypher string) (raw string, err error)
}
