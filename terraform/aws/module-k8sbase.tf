module k8sbase {
  source = "app.terraform.io/unionpoc/union/aws//modules/aws_k8sbase"

  env_name = var.env_name

  layer_name = var.layer_name

  module_name = k8sbase

  k8s_cluster_name = ${module.k8scluster.k8s_cluster_name}

  k8s_version = ${module.k8scluster.k8s_version}

  eks_cluster_name = ${module.k8scluster.k8s_cluster_name}

  s3_log_bucket_name = ${module.base.s3_log_bucket_name}

  openid_provider_url = ${module.k8scluster.k8s_openid_provider_url}

  openid_provider_arn = ${module.k8scluster.k8s_openid_provider_arn}

  nginx_high_availability = false

  linkerd_high_availability = false

  linkerd_enabled = true

  nginx_enabled = true

  admin_arns = []

  nginx_config = {}

  nginx_extra_tcp_ports = {}

  nginx_extra_tcp_ports_tls = []

  expose_self_signed_ssl = false

  cert_manager_values = {}

  linkerd_values = {}

  ingress_nginx_values = {}

  enable_auto_dns = true

  domain = ${module.dns.domain}

  zone_id = ${module.dns.zone_id}

  cert_arn = ${module.dns.cert_arn}
}