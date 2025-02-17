resource "aws_ecs_cluster" "main" {
  name = "operation-cluster"
}

resource "aws_cloudwatch_log_group" "bench" {
  name = "/ecs/bench"
}

resource "aws_iam_role" "ecs_task" {
  name = "ecs-task-role"
  assume_role_policy = templatefile("../../files/policy/assume_role.json", {
    service_name = "ecs-tasks"
  })
  managed_policy_arns = [
    "arn:aws:iam::aws:policy/AmazonSQSFullAccess",
    "arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess",
  ]

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
  managed_policy_arns = [
    "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy",
  ]
}
