module "codebuild_runner" {
  source             = "../../modules/codebuild_runner"
  repository_url     = local.repository_url
  github_pat         = local.secrets.github_pat
  vpc_id             = data.aws_vpc.main.id
  subnet_ids         = [data.aws_subnet.private_az1.id, data.aws_subnet.private_az4.id]
  security_group_ids = [data.aws_security_group.main.id]
}
