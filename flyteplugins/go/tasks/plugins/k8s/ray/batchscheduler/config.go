package batchscheduler

type BatchSchedulerConfig struct {
	Scheduler  string `json:"scheduler"`
	Parameters string `json:"parameters,omitempty`
}
