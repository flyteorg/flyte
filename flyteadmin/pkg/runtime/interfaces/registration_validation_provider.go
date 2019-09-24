package interfaces

type RegistrationValidationConfig struct {
	MaxWorkflowNodes     int    `json:"maxWorkflowNodes"`
	MaxLabelEntries      int    `json:"maxLabelEntries"`
	MaxAnnotationEntries int    `json:"maxAnnotationEntries"`
	WorkflowSizeLimit    string `json:"workflowSizeLimit"`
}

// Provides validation limits used at entity registration
type RegistrationValidationConfiguration interface {
	GetWorkflowNodeLimit() int
	GetMaxLabelEntries() int
	GetMaxAnnotationEntries() int
	GetWorkflowSizeLimit() string
}
