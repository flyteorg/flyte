package kubernetes

var (
	DefaultScheduler = "default"
)

type Plugin struct{}

func NewPlugin() *Plugin {
	return &Plugin{}
}

func (p *Plugin) Process(app interface{}) error { return nil }
