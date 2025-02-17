resource "github_repository" "main" {
  name       = local.team.name
  visibility = "private"
}

resource "github_repository_collaborators" "main" {
  repository = github_repository.main.name

  dynamic "user" {
    for_each = { for member in local.team.members : member.github_username => member }
    content {
      username   = user.key
      permission = "admin"
    }
  }
}

resource "github_actions_variable" "trigger_bench_endpoint" {
  repository    = github_repository.main.name
  variable_name = "TRIGGER_BENCH_ENDPOINT"
  value         = data.terraform_remote_state.operation.outputs.trigger_bench_endpoint
}

resource "github_actions_variable" "team_name" {
  repository    = github_repository.main.name
  variable_name = "TEAM_NAME"
  value         = local.team.name
}

resource "github_actions_variable" "application_endpoint" {
  repository    = github_repository.main.name
  variable_name = "APPLICATION_ENDPOINT"
  value         = aws_lb.main.dns_name
}

data "aws_caller_identity" "main" {}

resource "github_actions_variable" "aws_account_id" {
  repository    = github_repository.main.name
  variable_name = "AWS_ACCOUNT_ID"
  value         = data.aws_caller_identity.main.account_id
}

resource "github_actions_variable" "db_cluster_host" {
  repository    = github_repository.main.name
  variable_name = "DB_CLUSTER_HOST"
  value         = aws_rds_cluster.main.endpoint
}

resource "github_actions_variable" "alb_tg_arn" {
  repository    = github_repository.main.name
  variable_name = "ALB_TG_ARN"
  value         = aws_lb_target_group.main.arn
}

resource "github_actions_variable" "alb_sg_id" {
  repository    = github_repository.main.name
  variable_name = "ALB_SG_ID"
  value         = aws_default_security_group.main.id
}

resource "github_actions_variable" "alb_subnet_id" {
  repository    = github_repository.main.name
  variable_name = "ALB_SUBNET_ID"
  value         = aws_subnet.private_az1.id
}
