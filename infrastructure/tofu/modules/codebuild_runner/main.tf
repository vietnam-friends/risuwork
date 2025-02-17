resource "aws_iam_role" "main" {
  name = "codebuild-service-role"
  assume_role_policy = templatefile("../../files/policy/assume_role.json", {
    service_name = "codebuild"
  })
  managed_policy_arns = [
    "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPowerUser",
    "arn:aws:iam::aws:policy/CloudWatchLogsFullAccess",
    "arn:aws:iam::aws:policy/AmazonECS_FullAccess",
    "arn:aws:iam::aws:policy/AmazonEC2FullAccess",
  ]
}

resource "aws_codebuild_source_credential" "main" {
  auth_type   = "PERSONAL_ACCESS_TOKEN"
  server_type = "GITHUB"
  token       = var.github_pat
}

resource "aws_codebuild_project" "main" {
  name          = "self-hosted-runner"
  build_timeout = 30
  service_role  = aws_iam_role.main.arn
  environment {
    compute_type    = "BUILD_GENERAL1_SMALL"
    image           = "aws/codebuild/amazonlinux2-x86_64-standard:5.0"
    type            = "LINUX_CONTAINER"
    privileged_mode = true
  }
  source {
    type            = "GITHUB"
    location        = var.repository_url
    git_clone_depth = 1
  }
  dynamic "vpc_config" {
    for_each = var.use_fleet ? [] : [1]
    content {
      vpc_id             = var.vpc_id
      subnets            = var.subnet_ids
      security_group_ids = var.security_group_ids
    }
  }
  artifacts {
    type = "NO_ARTIFACTS"
  }
  cache {
    type  = "LOCAL"
    modes = ["LOCAL_DOCKER_LAYER_CACHE", "LOCAL_SOURCE_CACHE"]
  }
}

resource "aws_codebuild_webhook" "main" {
  project_name = aws_codebuild_project.main.name
  build_type   = "BUILD"
  filter_group {
    filter {
      type    = "EVENT"
      pattern = "WORKFLOW_JOB_QUEUED"
    }
  }
}

# MEMO: 構築当時はcodebuild_projectとfleetの紐付けが未対応だったため、紐付けは構築後に手動で行う必要があった。
# 現在は aws_codebuild_project.environment.fleet で紐付け可能
resource "awscc_codebuild_fleet" "main" {
  count            = var.use_fleet ? 1 : 0
  base_capacity    = 1
  compute_type     = "BUILD_GENERAL1_SMALL"
  environment_type = "LINUX_CONTAINER"

  fleet_vpc_config = {
    vpc_id             = var.vpc_id
    subnets            = [var.subnet_ids[0]]
    security_group_ids = var.security_group_ids
  }
  fleet_service_role = aws_iam_role.main.arn
}
