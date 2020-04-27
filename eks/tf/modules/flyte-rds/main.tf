resource "aws_rds_cluster_instance" "flyte_instances" {
  count              = 1
  identifier         = "flyteadmin-instances-${count.index}"
  cluster_identifier = aws_rds_cluster.flyteadmin.id
  instance_class     = "db.t3.medium"
  engine               = "aurora-postgresql"
  engine_version       = "11.6"
}

resource "aws_rds_cluster" "flyteadmin" {
  cluster_identifier      = "flyteadmin-cluster"
  engine                  = "aurora-postgresql"
  engine_version          = "11.6"
  availability_zones      = ["us-west-2a", "us-west-2b", "us-west-2c"]
  database_name           = "flyteadmin"
  master_username         = "flyteadmin"
  master_password         = "spongebob"
  backup_retention_period = 1
  preferred_backup_window = "07:00-09:00"
  skip_final_snapshot     = true
}

