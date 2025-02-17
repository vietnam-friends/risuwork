# 過去に作成済みのリソースを流用
data "aws_vpc" "main" {
  filter {
    name   = "tag:Name"
    values = ["test-vpc"]
  }
}

data "aws_security_group" "main" {
  vpc_id = data.aws_vpc.main.id

  filter {
    name   = "group-name"
    values = ["default"]
  }
}

data "aws_subnet" "private_az1" {
  vpc_id               = data.aws_vpc.main.id
  availability_zone_id = "apne1-az1"
  tags = {
    Name = "test-subnet-private2-ap-northeast-1c"
  }
}

data "aws_subnet" "private_az4" {
  vpc_id               = data.aws_vpc.main.id
  availability_zone_id = "apne1-az4"
  tags = {
    Name = "test-subnet-private1-ap-northeast-1a"
  }
}

data "aws_route_table" "private_az1" {
  filter {
    name   = "tag:Name"
    values = ["test-rtb-private2-ap-northeast-1c"]
  }
}

data "aws_route_table" "private_az4" {
  filter {
    name   = "tag:Name"
    values = ["test-rtb-private1-ap-northeast-1a"]
  }
}

// VPC Peering
data "terraform_remote_state" "participants" {
  for_each  = local.teams
  backend   = "s3"
  workspace = each.key

  config = {
    region = "ap-northeast-1"
    bucket = "risuwork-tofu-state-test"
    key    = "participants/tofu.tfstate"
  }
}

resource "aws_vpc_peering_connection_accepter" "main" {
  for_each                  = local.teams
  vpc_peering_connection_id = data.terraform_remote_state.participants[each.key].outputs.vpc_peering_id
  auto_accept               = true
}

resource "aws_route" "private_az1" {
  for_each                  = local.teams
  route_table_id            = data.aws_route_table.private_az1.id
  destination_cidr_block    = data.terraform_remote_state.participants[each.key].outputs.vpc_cider_block
  vpc_peering_connection_id = data.terraform_remote_state.participants[each.key].outputs.vpc_peering_id
}

resource "aws_route" "private_az4" {
  for_each                  = local.teams
  route_table_id            = data.aws_route_table.private_az4.id
  destination_cidr_block    = data.terraform_remote_state.participants[each.key].outputs.vpc_cider_block
  vpc_peering_connection_id = data.terraform_remote_state.participants[each.key].outputs.vpc_peering_id
}
