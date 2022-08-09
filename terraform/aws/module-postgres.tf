module postgres {
	source = "unionai-oss/opta/unionai//modules/aws_postgres"

	env_name = var.env_name

	layer_name = var.layer_name

	module_name = postgres

	instance_class = "db.t3.medium"

	engine_version = 11.9

	multi_az = false

	safety = false

	backup_retention_days = 7

	extra_security_groups_ids = []

	create_global_database = false

	database_name = app

}
