terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.54.1"
    }
    slack = {
      source  = "pablovarela/slack"
      version = "~> 1.2.2"
    }
  }

  backend "s3" {
    region = "ap-northeast-1"
    bucket = "risuwork-tofu-state-test"
    key    = "operation/tofu.tfstate"
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

provider "aws" {
  alias  = "us_east"
  region = "us-east-1"
  default_tags {
    tags = {
      Project = "risuwork"
    }
  }
}

provider "slack" {
  token = local.secrets.slack_token
}
