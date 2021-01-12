terraform {
  required_version = ">= 0.12.0"
}

provider "aws" {
  profile = "default"
  region  = var.region
}

# Use the internal module to create an RDS instance with a user-supplied VPC
module "flyte_rds" {
  source = "./modules/flyte-rds"

  rds_vpc = var.rds_vpc
}

# Use the internal module to create an EKS cluster, which has its own VPC
module "flyte_eks" {
  source           = "./modules/flyte-eks"
  eks_cluster_name = var.eks_cluster_name
}

# Get information about the two VPCs
data "aws_vpc" "rds_vpc" {
  id = var.rds_vpc
}

data "aws_vpc" "eks_vpc" {
  id         = module.flyte_eks.eks_vpc_id
  depends_on = ["module.flyte_eks"]
}

# Get information about the RDS instance
data "aws_db_instance" "flyte_rds" {
  db_instance_identifier = module.flyte_rds.admin_rds_instance_id
  depends_on             = ["module.flyte_rds.admin_rds_name"]
}

resource "aws_vpc_peering_connection" "eks_to_main_peering" {
  peer_vpc_id = var.rds_vpc
  vpc_id      = module.flyte_eks.eks_vpc_id
  auto_accept = true

  tags = {
    Name = "VPC peering connection between Flyte RDS and EKS"
  }

  accepter {
    allow_remote_vpc_dns_resolution = true
  }

  requester {
    allow_remote_vpc_dns_resolution = true
  }
}

data "aws_route_table" "eks_public_route_table" {
  vpc_id = module.flyte_eks.eks_vpc_id
  filter {
    name   = "tag:Name"
    values = ["${var.eks_cluster_name}-vpc-public"]
  }
}

resource "aws_route" "route_rds_cidr" {
  route_table_id            = data.aws_route_table.eks_public_route_table.id
  destination_cidr_block    = data.aws_vpc.rds_vpc.cidr_block
  vpc_peering_connection_id = aws_vpc_peering_connection.eks_to_main_peering.id
}

# Add a rule to the RDS security group to allow access from the EKS VPC
resource "aws_security_group_rule" "allow_eks_to_rds" {
  depends_on        = ["module.flyte_rds.admin_rds_name"]
  type              = "ingress"
  from_port         = 5432
  to_port           = 5432
  protocol          = "tcp"
  cidr_blocks       = [data.aws_vpc.eks_vpc.cidr_block]
  security_group_id = data.aws_db_instance.flyte_rds.vpc_security_groups[0]
}

# The following implements the instructions set forth by:
#   https://github.com/aws/amazon-eks-pod-identity-webhook/blob/95808cffe6d801822dae122f2f2c87a258d70bb8/README.md
# This is a webhook that will allow pods to assume arbitrarily constrained roles via their service account.
# TODO: This should be moved into a separate module probably but will require further refactoring as the assume role
#       policy used is also used further below in the ALB ingress module.

# Create an oidc provider using the EKS cluster's public OIDC discovery endpoint
resource "aws_iam_openid_connect_provider" "eks_oidc_connection" {
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = ["9e99a48a9960b14926bb7f3b02e22da2b0ab7280"]
  url             = module.flyte_eks.eks_oidc_issuer
}

locals {
  issuer_parsed = regex("^arn.*(?P<trailing>oidc.eks.*)", aws_iam_openid_connect_provider.eks_oidc_connection.arn)
}

# This is the trust document that will allow pods to use roles that they specify in their service account
data "aws_iam_policy_document" "let_pods_assume_roles" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.eks_oidc_connection.arn]
    }

    condition {
      test     = "StringLike"
      variable = "${local.issuer_parsed.trailing}:sub"

      values = [
        "system:serviceaccount:*:*",
      ]
    }
  }
}

# Make a role for Flyte components themselves to use
resource "aws_iam_role" "flyte_operator" {
  name               = "flyte-operator"
  assume_role_policy = data.aws_iam_policy_document.let_pods_assume_roles.json
}


# Make a policy document
# TODO: Scope this down later
data "aws_iam_policy_document" "all_s3_access" {
  statement {
    actions = [
      "s3:*",
    ]

    resources = [
      "*",
    ]
  }
}

# Use the policy document to create a policy
resource "aws_iam_policy" "flyte_operator_s3_access" {
  name   = "flyte_operator_s3_access"
  path   = "/"
  policy = data.aws_iam_policy_document.all_s3_access.json
}

# Attach the policy to the flyte operator role
resource "aws_iam_role_policy_attachment" "flyte_operator_s3_attach" {
  role       = aws_iam_role.flyte_operator.name
  policy_arn = aws_iam_policy.flyte_operator_s3_access.arn
}

module "alb_ingress" {
  source = "./modules/alb-ingress"

  region                    = var.region
  eks_cluster_name          = var.eks_cluster_name
  cluster_id                = module.flyte_eks.cluster_id
  eks_vpc_id                = module.flyte_eks.eks_vpc_id
  assume_role_policy_string = data.aws_iam_policy_document.let_pods_assume_roles.json
}

