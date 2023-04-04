package launchplan

//go:generate pflags UpdateConfig --default-var UConfig --bind-default-var
var (
	UConfig = &UpdateConfig{}
)

// Config
type UpdateConfig struct {
	Archive  bool   `json:"archive" pflag:",disable the launch plan schedule (if it has an active schedule associated with it)."`
	Activate bool   `json:"activate" pflag:",activate launchplan."`
	DryRun   bool   `json:"dryRun" pflag:",execute command without making any modifications."`
	Version  string `json:"version" pflag:",version of the launchplan to be fetched."`
}
