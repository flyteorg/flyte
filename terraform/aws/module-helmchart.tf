module helmchart {
  source = "app.terraform.io/unionpoc/union/aws//modules/aws_helmchart"

  env_name = var.env_name

  layer_name = var.layer_name

  module_name = helmchart

  chart = ../../charts/flyte-core

  release_name = ""

  repository = ""

  namespace = flyte

  create_namespace = true

  atomic = true

  cleanup_on_fail = true

  chart_version = ""

  values_files = [./../../charts/flyte-core/values-eks.yaml]

  values_file = ../../charts/flyte-core/values-eks.yaml

  values db datacatalog database {
    port = 5432

    username = ${module.postgres.db_user}

    host = ${module.postgres.db_host}

    dbname = ${module.postgres.db_name}
  }

  values db admin database {
    port = 5432

    username = ${module.postgres.db_user}

    host = ${module.postgres.db_host}

    dbname = ${module.postgres.db_name}
  }

  values common ingress {
    albSSLRedirect = false

    host = ${data.terraform_remote_state.parent.outputs.domain}

    annotations = {
      kubernetes.io/ingress.class = nginx

      nginx.ingress.kubernetes.io/app-root = /console
    }
  }

  values common databaseSecret secretManifest stringData {
    pass.txt = ${module.postgres.db_password}
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
      eks.amazonaws.com/role-arn = ${module.adminflyterole.role_arn}
    }
  }

  values datacatalog serviceAccount {
    create = true

    annotations = {
      eks.amazonaws.com/role-arn = ${module.adminflyterole.role_arn}
    }
  }

  values flytepropeller serviceAccount {
    create = true

    annotations = {
      eks.amazonaws.com/role-arn = ${module.adminflyterole.role_arn}
    }
  }

  values flytescheduler serviceAccount {
    create = true

    annotations = {
      eks.amazonaws.com/role-arn = ${module.adminflyterole.role_arn}
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
    rawoutput-prefix = s3://production-service-flyte
  }

  values cluster_resource_manager {
    enabled = true

    config cluster_resources {
      customData = {
        production defaultIamRole {
          value = ${module.userflyterole.role_arn}
        }

        production projectQuotaCpu {
          value = 6
        }

        production projectQuotaMemory {
          value = 6000Mi
        }
      }

      customData = {
        staging defaultIamRole {
          value = ${module.userflyterole.role_arn}
        }

        staging projectQuotaCpu {
          value = 6
        }

        staging projectQuotaMemory {
          value = 6000Mi
        }
      }

      customData = {
        development defaultIamRole {
          value = ${module.userflyterole.role_arn}
        }

        development projectQuotaCpu {
          value = 6
        }

        development projectQuotaMemory {
          value = 6000Mi
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