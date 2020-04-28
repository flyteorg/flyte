output "admin_rds_name" {
  value = aws_rds_cluster.flyteadmin.endpoint
}

output "admin_rds_instance_id" {
  value = aws_rds_cluster_instance.flyte_instances[0].identifier
}

