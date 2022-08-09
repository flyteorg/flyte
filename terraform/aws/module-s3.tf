module s3 {
	source = "unionai-oss/opta/unionai//modules/aws_s3"

	env_name = var.env_name

	layer_name = var.layer_name

	module_name = s3

	s3_log_bucket_name = "${data.terraform_remote_state.parent.outputs.s3_log_bucket_name}"

	bucket_name = production-service-flyte

	same_region_replication = false

	block_public = true
}
