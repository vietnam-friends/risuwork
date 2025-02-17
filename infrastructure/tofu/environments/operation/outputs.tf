output "account_id" {
  value = data.aws_caller_identity.main.account_id
}

output "vpc_id" {
  value = data.aws_vpc.main.id
}

output "sg_id" {
  value = data.aws_security_group.main.id
}

output "trigger_bench_endpoint" {
  value = aws_lambda_function_url.trigger_bench.function_url
}
