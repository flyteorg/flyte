package artifacts

// gatepr: add proper config bits for this
// eduardo to consider moving to idl clients.
type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Insecure bool   `json:"insecure"`
}
