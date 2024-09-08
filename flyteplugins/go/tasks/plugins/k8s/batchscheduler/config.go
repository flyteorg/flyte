package batchscheduler

type Config struct {
	Scheduler  string `json:"scheduler,omitempty" pflag:", Specify batch scheduler to"`
	Parameters string `json:"parameters,omitempty" pflag:", Specify static parameters"`
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
