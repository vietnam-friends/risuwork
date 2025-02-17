resource "aws_db_subnet_group" "main" {
  name       = "default-subnet-group"
  subnet_ids = [aws_subnet.private_az1.id, aws_subnet.private_az4.id]
}

resource "aws_iam_role" "rds_monitoring" {
  name = "rds-monitoring-role"
  assume_role_policy = templatefile("../../files/policy/assume_role.json", {
    service_name = "monitoring.rds"
  })
}

resource "aws_iam_role_policy_attachment" "rds_monitoring" {
  role       = aws_iam_role.rds_monitoring.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
}

resource "aws_rds_cluster_parameter_group" "main" {
  name   = "db-cluster-pg"
  family = "aurora-mysql8.0"
}

resource "aws_db_parameter_group" "main" {
  name   = "db-instance-pg"
  family = "aurora-mysql8.0"

  parameter {
    name  = "slow_query_log"
    value = "1"
  }
  parameter {
    name  = "long_query_time"
    value = "0"
  }
}

resource "aws_rds_cluster" "main" {
  cluster_identifier     = "db-cluster"
  engine                 = "aurora-mysql"
  engine_version         = "8.0.mysql_aurora.3.05.2"
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_default_security_group.main.id]

  database_name                    = "risuwork"
  master_username                  = "root"
  master_password                  = local.secrets.db_password
  db_cluster_parameter_group_name  = aws_rds_cluster_parameter_group.main.name
  db_instance_parameter_group_name = aws_db_parameter_group.main.name

  enabled_cloudwatch_logs_exports = ["slowquery"]

  skip_final_snapshot = true
  apply_immediately   = true
}

resource "aws_rds_cluster_instance" "writer" {
  cluster_identifier   = aws_rds_cluster.main.id
  identifier           = "db-instance"
  engine               = aws_rds_cluster.main.engine
  engine_version       = aws_rds_cluster.main.engine_version
  instance_class       = "db.t4g.large"
  db_subnet_group_name = aws_rds_cluster.main.db_subnet_group_name
  availability_zone    = aws_subnet.private_az1.availability_zone

  db_parameter_group_name = aws_db_parameter_group.main.name

  monitoring_interval          = 1
  monitoring_role_arn          = aws_iam_role.rds_monitoring.arn
  performance_insights_enabled = true
}
