module helmchart {
	source = "unionai-oss/opta/unionai//modules/helm_chart"

	env_name = var.env_name

	layer_name = var.layer_name

	module_name = helmchart

	chart = var.chart

	release_name = var.release_name

	repository = var.repository

	namespace = var.namespace

	create_namespace = true

	atomic = true

	cleanup_on_fail = true

	chart_version = ""

	values_files = var.values_files

	values_file = var.values_file

	values db datacatalog database {
		port = 5432

		username = "${module.postgres.db_user}"

		host = "${module.postgres.db_host}"

		dbname = "${module.postgres.db_name}"
	}

	values db admin database {
		port = 5432

		username = "${module.postgres.db_user}"

		host = "${module.postgres.db_host}"

		dbname = "${module.postgres.db_name}"
	}

	values common ingress {
		albSSLRedirect = false

		host = "${data.terraform_remote_state.parent.outputs.domain}"

		annotations = {
			"kubernetes.io/ingress.class" = "nginx"

			"nginx.ingress.kubernetes.io/app-root" = "/console"
		}
	}

	values common databaseSecret secretManifest stringData {
		# pass.txt = "${module.postgres.db_password}"
	}

	values storage {
		bucketName = production-service-flyte

		s3 = {
			region = us-east-2
		}
	}

	values flyteadmin serviceAccount {
		create = true

		annotations = {
			eks.amazonaws.com/role-arn = "${module.adminflyterole.role_arn}"
		}
	}

	values datacatalog serviceAccount {
		create = true

		annotations = {
			eks.amazonaws.com/role-arn = "${module.adminflyterole.role_arn}"
		}
	}

	values flytepropeller serviceAccount {
		create = true

		annotations = {
			eks.amazonaws.com/role-arn = "${module.adminflyterole.role_arn}"
		}
	}

	values flytescheduler serviceAccount {
		create = true

		annotations = {
			eks.amazonaws.com/role-arn = "${module.adminflyterole.role_arn}"
		}
	}

	values workflow_scheduler {
		enabled = true
	}

	values workflow_notifications {
		enabled = false
	}

	values configmap remoteData remoteData {
		region = us-east-2

		scheme = aws

		signedUrls = {
			durationMinutes = 3
		}
	}

	values configmap task_logs plugins logs {
		cloudwatch-region = us-east-2
	}

	values configmap core propeller {
		rawoutput-prefix = "s3://production-service-flyte"
	}

	values cluster_resource_manager {
		enabled = true

		config cluster_resources {
			customData = {
				production = {
					defaultIamRole = {
						value = "${module.iam_role.role_arn}"
					}
				}
				production = {
					projectQuotaCpu = {
						value = 6
					}
				}
				projectQuotaMemory = {
					defaultIamRole = {
						value = "6000Mi"
					}
				}

				staging = {
					defaultIamRole = {
						value = "${module.iam_role.role_arn}"
					}
				}
				staging = {
					projectQuotaCpu = {
						value = 6
					}
				}
				staging = {
					defaultIamRole = {
						value = "6000Mi"
					}
				}

			}

		}
	}

	timeout = 600

	dependency_update = true

	wait = true

	wait_for_jobs = false

	max_history = 25
}
