# ---------------------------------------------------------------------------------------------------------------------
# ENVIRONMENT VARIABLES
# Define these secrets as environment variables
# ---------------------------------------------------------------------------------------------------------------------

# ---------------------------------------------------------------------------------------------------------------------
# OPTIONAL PARAMETERS
# These parameters have reasonable defaults.
# ---------------------------------------------------------------------------------------------------------------------

variable "domain" {
	description = "Domain Name"
	type        = string
	default     = null
}

variable "delegated" {
	description = ""
	type        = bool
	default     = false
}


variable "env_name" {
	description = "Env Name"
	type        = string
	default     = null
}

variable "cluster_name" {
	description = "What to name the Consul cluster and all of its associated resources"
	type        = string
	default     = "flyte-example"
}

variable "layer_name" {
	description = "service name that will be deployed in env"
	type        = string
	default     = "flyte"
}

variable "num_clients" {
	description = "The number of Consul client nodes to deploy. You typically run the Consul client alongside your apps, so set this value to however many Instances make sense for your app code."
	type        = number
	default     = 6
}

variable "max_nodes" {
	description = "Max node for eks default node group "
	type        = number
	default     = 10
}

variable "min_nodes" {
	description = "Min node for eks default node group"
	type        = number
	default     = 5
}

variable "node_disk_size" {
	description = "The ID of the VPC in which the nodes will be deployed.  Uses default VPC if not supplied."
	type        = number
	default     = 20
}

variable "node_instance_type" {
	description = "Node instance type"
	type        = string
	default     = null
}

variable "k8s_version" {
	description = "K8s version upgrade"
	type        = string
	default     = null
}

variable "spot_instances" {
	description = "If set to true, eks will use spot instance ."
	type        = bool
	default     = false
}

variable "release_name" {
	description = "Release name for helm chart"
	type        = string
	default     = null
}

variable "chart" {
	description = "Chart name for deployment"
	type        = string
	default     = null
}

variable "repository" {
	description = "Helm repository url"
	type        = string
	default     = null
}

variable "namespace" {
	description = "Namespace name for helm deployment"
	type        = string
	default     = null
}


