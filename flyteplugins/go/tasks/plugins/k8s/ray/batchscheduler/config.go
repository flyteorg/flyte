package batchscheduler

type BatchSchedulerConfig struct {
	Scheduler  string `json:"scheduler,omitempty"`
	Parameters string `json:"parameters,omitempty"`
}

func NewDefaultBatchSchedulerConfig() BatchSchedulerConfig {
	return BatchSchedulerConfig{
		Scheduler:  "",
		Parameters: "",
	}
}

func (b *BatchSchedulerConfig) GetScheduler() string {
	return b.Scheduler
}

func (b *BatchSchedulerConfig) GetParameters() string {
	return b.Parameters
}
