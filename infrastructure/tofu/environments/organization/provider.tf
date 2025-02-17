terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.54.1"
    }
    github = {
      source  = "integrations/github"
      version = "~> 6.2.2"
    }
  }

  backend "s3" {
    region = "ap-northeast-1"
    bucket = "risuwork-tofu-state-test"
    key    = "organization/tofu.tfstate"
  }
}

provider "aws" {
  region = "ap-northeast-1"
  default_tags {
    tags = {
      Project = "risuwork"
    }
  }
}

provider "github" {
  owner = "risuwork"
  token = local.secrets.github_pat
}
