package config

//go:generate enumer --type=CloudDeployment -json -yaml -trimprefix=CloudDeployment
type CloudDeployment uint8

const (
	CloudDeploymentNone CloudDeployment = iota
	CloudDeploymentAWS
	CloudDeploymentGCP
	CloudDeploymentSandbox
	CloudDeploymentLocal
)
