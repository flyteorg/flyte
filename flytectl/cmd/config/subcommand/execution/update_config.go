package execution

//go:generate pflags UpdateConfig --default-var UConfig --bind-default-var
var (
	UConfig = &UpdateConfig{}
)

// UpdateConfig
type UpdateConfig struct {
	Archive  bool `json:"archive" pflag:",archive execution."`
	Activate bool `json:"activate" pflag:",activate execution."`
	DryRun   bool `json:"dryRun" pflag:",execute command without making any modifications."`
	Force    bool `json:"force" pflag:",do not ask for an acknowledgement during updates."`
}
