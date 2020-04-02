package models

// Workflow primary key
type WorkflowKey struct {
	Project string `gorm:"primary_key:workflow_project_domain_name_idx,workflow_project_domain_idx"`
	Domain  string `gorm:"primary_key;index:workflow_project_domain_name_idx,workflow_project_domain_idx"`
	Name    string `gorm:"primary_key;index:workflow_project_domain_name_idx"`
	Version string `gorm:"primary_key"`
}

// Database model to encapsulate a workflow.
type Workflow struct {
	BaseModel
	WorkflowKey
	TypedInterface          []byte
	RemoteClosureIdentifier string `gorm:"not null"`
	LaunchPlans             []LaunchPlan
	Executions              []Execution
	// Hash of the compiled workflow closure
	Digest []byte
	// GORM doesn't save the zero value for ints, so we use a pointer for the State field
	State *int32 `gorm:"default:0"`
}
