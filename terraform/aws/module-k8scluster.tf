module k8scluster {
  source = "app.terraform.io/unionpoc/union/aws//modules/aws_k8scluster"

  env_name = var.env_name

  layer_name = var.layer_name

  module_name = k8scluster

  cluster_name = var.cluster_name

  kms_account_key_arn = ${module.base.kms_account_key_arn}

  private_subnet_ids = ${module.base.private_subnet_ids}

  vpc_id = ${module.base.vpc_id}

  eks_log_retention = 7

  max_nodes = var.max_nodes

  min_nodes = var.min_nodes

  node_disk_size = var.node_disk_size

  node_instance_type = var.node_instance_type

  k8s_version = var.k8s_version

  control_plane_security_groups = []

  spot_instances = var.spot_instances

  enable_metrics = false

  node_launch_template = {}

  ami_type = AL2_x86_64
}