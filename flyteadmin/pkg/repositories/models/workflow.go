package models

// Workflow primary key
type WorkflowKey struct {
	Project string `gorm:"primary_key;index:workflow_project_domain_name_idx;index:workflow_project_domain_idx"  valid:"length(0|255)"`
	Domain  string `gorm:"primary_key;index:workflow_project_domain_name_idx;index:workflow_project_domain_idx"  valid:"length(0|255)"`
	Name    string `gorm:"primary_key;index:workflow_project_domain_name_idx"  valid:"length(0|255)"`
	Version string `gorm:"primary_key"`
}

// Database model to encapsulate a workflow.
type Workflow struct {
	BaseModel
	WorkflowKey
	TypedInterface          []byte
	RemoteClosureIdentifier string `gorm:"not null" valid:"length(0|255)"`
	// Hash of the compiled workflow closure
	Digest []byte
	// ShortDescription for the workflow.
	ShortDescription string
}

var WorkflowColumns = modelColumns(Workflow{})
