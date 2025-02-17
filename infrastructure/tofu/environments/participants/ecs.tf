resource "aws_ecs_cluster" "main" {
  name = "participants-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

resource "aws_cloudwatch_log_group" "app" {
  name = "/ecs/app"
}

resource "aws_cloudwatch_log_group" "nginx" {
  name = "/ecs/nginx"
}

# ECS Task Role
resource "aws_iam_role" "ecs_task" {
  name = "ecs-task-role"
  assume_role_policy = templatefile("../../files/policy/assume_role.json", {
    service_name = "ecs-tasks"
  })
  managed_policy_arns = ["arn:aws:iam::aws:policy/AWSXrayFullAccess"]

  inline_policy {
    name   = "allow-ecs-exec"
    policy = file("../../files/policy/ecs_task_allow_ecs_exec.json")
  }
}

resource "aws_iam_role" "ecs_task_exec" {
  name = "ecs-task-execution-role"
  assume_role_policy = templatefile("../../files/policy/assume_role.json", {
    service_name = "ecs-tasks"
  })
  managed_policy_arns = ["arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"]

  inline_policy {
    name   = "use-pull-through-cache-repo"
    policy = file("./files/policy/ecs_task_exec_use_pull_through_cache_repo.json")
  }
}

# Node Resource
data "aws_ssm_parameter" "ecs_ami" {
  name = "/aws/service/ecs/optimized-ami/amazon-linux-2023/recommended/image_id"
}

resource "aws_iam_role" "ecs_node" {
  name = "ecs-node-role"
  assume_role_policy = templatefile("../../files/policy/assume_role.json", {
    service_name = "ec2"
  })
  managed_policy_arns = [
    "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role",
    "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore",
  ]
}

resource "aws_iam_instance_profile" "ecs_node" {
  role = aws_iam_role.ecs_node.id
}
