module userflyterole {
  source = "unionai-oss/opta/unionai//modules/aws_iam_role"
  env_name = var.env_name

  layer_name = var.layer_name

  module_name = userflyterole

  iam_policy = {
    Version = "2012-10-17"

    Statement = {
      Sid = PolicySimulatorAPI

      Action = ["iam:GetContextKeysForCustomPolicy", "iam:GetContextKeysForPrincipalPolicy", "iam:SimulateCustomPolicy", "iam:SimulatePrincipalPolicy"]

      Effect = Allow

      Resource = "*"
    }

    Statement = {
      Sid = PolicySimulatorConsole

	  Action = ["iam:GetGroup", "iam:GetGroupPolicy", "iam:GetPolicy", "iam:GetPolicyVersion", "iam:GetRole", "iam:GetRolePolicy", "iam:GetUser", "iam:GetUserPolicy", "iam:ListAttachedGroupPolicies", "iam:ListAttachedRolePolicies", "iam:ListAttachedUserPolicies", "iam:ListGroups", "iam:ListGroupPolicies", "iam:ListGroupsForUser", "iam:ListRolePolicies", "iam:ListRoles", "iam:ListUserPolicies", "iam:ListUsers"]

	  Effect = Allow

      Resource = "*"
    }

    Statement = {
      Sid = WriteBuckets

      Action = ["s3:GetObject*", "s3:PutObject*", "s3:DeleteObject*", "s3:ListBucket"]

      Effect = Allow

      Resource = ["arn:aws:s3:::production-service-flyte, arn:aws:s3:::production-service-flyte/*"]
    }
  }

  allowed_k8s_services = {
    namespace = "*"

    service_name = "*"
  }

  allowed_iams = []

  extra_iam_policies = ["arn:aws:iam::aws:policy/CloudWatchEventsFullAccess"]

  links = {
    s3 = [write]
  }

  kubernetes_trusts = {
    open_id_url = "${data.terraform_remote_state.parent.outputs.k8s_openid_provider_url}"

    open_id_arn = "${data.terraform_remote_state.parent.outputs.k8s_openid_provider_arn}"

    service_name = "*"

    namespace = "*"
  }
}
