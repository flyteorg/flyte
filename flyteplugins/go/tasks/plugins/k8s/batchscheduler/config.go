package batchscheduler

type Config struct {
	Scheduler string                      `json:"scheduler,omitempty" pflag:", Specify batch scheduler to"`
	Default   SchedulingConfig            `json:"default,omitempty" pflag:", Specify default scheduling config which batch scheduler adopts"`
	NameSpace map[string]SchedulingConfig `json:"Namespace,omitempty" pflag:"-, Specify namespace scheduling config"`
	Domain    map[string]SchedulingConfig `json:"Domain,omitempty" pflag:"-, Specify domain scheduling config"`
}

type SchedulingConfig struct {
	KueueConfig    `json:"Kueue,omitempty" pflag:", Specify Kueue scheduling scheduling config"`
	YunikornConfig `json:"Yunikorn,omitempty" pflag:", Yunikorn scheduling config"`
}

type KueueConfig struct {
	PriorityClassName string `json:"Priority,omitempty" pflag:", Kueue Prioty class"`
	Queue             string `json:"Queue,omitempty" pflag:", Specify batch scheduler to"`
}

type YunikornConfig struct {
	Parameters string `json:"parameters,omitempty" pflag:", Specify gangscheduling policy"`
	Queue      string `json:"queue,omitempty" pflag:", Specify leaf queue to submit to"`
}
