module base {
	source = "unionai-oss/opta/unionai//modules/aws_base"

	env_name = var.env_name

	layer_name = var.layer_name

	module_name = base

	total_ipv4_cidr_block = "10.0.0.0/16"

	vpc_log_retention = 90

	private_ipv4_cidr_blocks = ["10.0.128.0/21", "10.0.136.0/21", "10.0.144.0/21"]

	public_ipv4_cidr_blocks = ["10.0.0.0/21", "10.0.8.0/21", "10.0.16.0/21"]

}
