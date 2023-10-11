package artifacts

type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Insecure bool   `json:"insecure"`
}
