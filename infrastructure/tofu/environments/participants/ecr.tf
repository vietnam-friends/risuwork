module "app_repo" {
  source = "../../modules/ecr"
  name   = "application-repository"
}

module "nginx_repo" {
  source = "../../modules/ecr"
  name   = "nginx-repository"
}
