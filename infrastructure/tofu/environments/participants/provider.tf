terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.54.1"
    }
    awscc = {
      source  = "hashicorp/awscc"
      version = "~> 1.10.0"
    }
    github = {
      source  = "integrations/github"
      version = "~> 6.2.2"
    }
    slack = {
      source  = "pablovarela/slack"
      version = "~> 1.2.2"
    }
  }

  backend "s3" {
    region  = "ap-northeast-1"
    bucket  = "risuwork-tofu-state-test"
    key     = "participants/tofu.tfstate"
    profile = "risuwork"
  }
}

provider "aws" {
  region  = "ap-northeast-1"
  profile = "risuwork-${terraform.workspace}"
  default_tags {
    tags = {
      Project = "risuwork"
    }
  }
}

provider "awscc" {
  region  = "ap-northeast-1"
  profile = "risuwork-${terraform.workspace}"
}

provider "github" {
  owner = "risuwork"
  token = local.secrets.github_pat
}

provider "slack" {
  token = local.secrets.slack_token
}
