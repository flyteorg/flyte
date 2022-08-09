module dns {
  source = "app.terraform.io/unionpoc/union/aws//modules/aws_dns"
  env_name = var.env_name
  layer_name = var.layer_name
  module_name = dns
  domain = var.domain
  delegated = var.delegated
  upload_cert = false
  cert_chain_included = false
  force_update = false
}