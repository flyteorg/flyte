package launchplan

//go:generate pflags UpdateConfig --default-var UConfig --bind-default-var
var (
	UConfig = &UpdateConfig{}
)

// Config
type UpdateConfig struct {
	Activate   bool   `json:"activate" pflag:",activate launchplan."`
	Archive    bool   `json:"archive" pflag:",(Deprecated) disable the launch plan schedule (if it has an active schedule associated with it)."`
	Deactivate bool   `json:"deactivate" pflag:",disable the launch plan schedule (if it has an active schedule associated with it)."`
	DryRun     bool   `json:"dryRun" pflag:",execute command without making any modifications."`
	Force      bool   `json:"force" pflag:",do not ask for an acknowledgement during updates."`
	Version    string `json:"version" pflag:",version of the launchplan to be fetched."`
}
