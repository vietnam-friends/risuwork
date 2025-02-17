module "codebuild_runner" {
  source             = "../../modules/codebuild_runner"
  repository_url     = github_repository.main.http_clone_url
  github_pat         = local.secrets.github_pat
  vpc_id             = aws_vpc.main.id
  subnet_ids         = [aws_subnet.private_az1.id]
  security_group_ids = [aws_default_security_group.main.id]
  use_fleet          = true
}
