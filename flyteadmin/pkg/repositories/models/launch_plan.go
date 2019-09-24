package models

// Launch plan primary key
type LaunchPlanKey struct {
	Project string `gorm:"primary_key;index:lp_project_domain_name_idx,lp_project_domain_idx"`
	Domain  string `gorm:"primary_key;index:lp_project_domain_name_idx,lp_project_domain_idx"`
	Name    string `gorm:"primary_key;index:lp_project_domain_name_idx"`
	Version string `gorm:"primary_key"`
}

type LaunchPlanScheduleType string

const (
	// LaunchPlanScheduleTypeNONE is the const representing the launch plan does not have a schedule
	LaunchPlanScheduleTypeNONE LaunchPlanScheduleType = "NONE"
	// LaunchPlanScheduleTypeCRON is the const representing the launch plan has a CRON type of schedule
	LaunchPlanScheduleTypeCRON LaunchPlanScheduleType = "CRON"
	// LaunchPlanScheduleTypeRATE is the launch plan has a RATE type of schedule
	LaunchPlanScheduleTypeRATE LaunchPlanScheduleType = "RATE"
)

// Database model to encapsulate a launch plan.
type LaunchPlan struct {
	BaseModel
	LaunchPlanKey
	Spec       []byte `gorm:"not null"`
	WorkflowID uint   `gorm:"index"`
	Closure    []byte `gorm:"not null"`
	// GORM doesn't save the zero value for ints, so we use a pointer for the State field
	State      *int32 `gorm:"default:0"`
	Executions []Execution
	// Hash of the launch plan
	Digest       []byte
	ScheduleType LaunchPlanScheduleType
}
