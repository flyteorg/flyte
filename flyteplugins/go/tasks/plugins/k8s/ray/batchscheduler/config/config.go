package config

type Config struct {
	Scheduler  string `json:"scheduler,omitempty"`
	Parameters string `json:"parameters,omitempty"`
}

func NewConfig() Config {
	return Config{
		Scheduler:  "",
		Parameters: "",
	}
}

func (b *Config) GetScheduler() string {
	return b.Scheduler
}

func (b *Config) GetParameters() string {
	return b.Parameters
}
