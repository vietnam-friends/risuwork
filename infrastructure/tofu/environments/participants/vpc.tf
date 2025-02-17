resource "aws_default_vpc" "main" {
  force_destroy = true

  tags = {
    Name = "deprecated-vpc"
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.${local.team.id}.0.0/16"

  tags = {
    Name = "participants-vpc"
  }
}

resource "aws_default_security_group" "main" {
  vpc_id = aws_vpc.main.id

  ingress {
    protocol  = -1
    self      = true
    from_port = 0
    to_port   = 0
  }

  # VPC Peering経由でベンチからアクセスを受付けられるように
  ingress {
    protocol  = -1
    from_port = 0
    to_port   = 0
    # VPC Peering設定完了後に有効化する必要あり
    security_groups = ["${data.terraform_remote_state.operation.outputs.account_id}/${data.terraform_remote_state.operation.outputs.sg_id}"]
  }

  egress {
    protocol    = -1
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "participants-default-sg"
  }
}

resource "aws_subnet" "public_az4" {
  vpc_id               = aws_vpc.main.id
  availability_zone_id = "apne1-az4"
  cidr_block           = "10.${local.team.id}.0.0/24"

  tags = {
    Name = "public-subnet-a"
  }
}

resource "aws_subnet" "public_az1" {
  vpc_id               = aws_vpc.main.id
  availability_zone_id = "apne1-az1"
  cidr_block           = "10.${local.team.id}.64.0/24"

  tags = {
    Name = "public-subnet-c"
  }
}

resource "aws_subnet" "private_az4" {
  vpc_id               = aws_vpc.main.id
  availability_zone_id = "apne1-az4"
  cidr_block           = "10.${local.team.id}.128.0/24"

  tags = {
    Name = "private-subnet-a"
  }
}

resource "aws_subnet" "private_az1" {
  vpc_id               = aws_vpc.main.id
  availability_zone_id = "apne1-az1"
  cidr_block           = "10.${local.team.id}.192.0/24"

  tags = {
    Name = "private-subnet-c"
  }
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
}

resource "aws_eip" "main" {
  domain = "vpc"

  depends_on = [aws_internet_gateway.main]
}

resource "aws_nat_gateway" "main" {
  allocation_id = aws_eip.main.id
  subnet_id     = aws_subnet.public_az1.id
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id
}

resource "aws_route" "public" {
  route_table_id         = aws_route_table.public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.main.id
}

resource "aws_route_table_association" "public_az4" {
  route_table_id = aws_route_table.public.id
  subnet_id      = aws_subnet.public_az4.id
}

resource "aws_route_table_association" "public_az1" {
  route_table_id = aws_route_table.public.id
  subnet_id      = aws_subnet.public_az1.id
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id
}

resource "aws_route" "private" {
  route_table_id         = aws_route_table.private.id
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.main.id
}

resource "aws_route_table_association" "private_az4" {
  route_table_id = aws_route_table.private.id
  subnet_id      = aws_subnet.private_az4.id
}

resource "aws_route_table_association" "private_az1" {
  route_table_id = aws_route_table.private.id
  subnet_id      = aws_subnet.private_az1.id
}

// VPC Peering
data "terraform_remote_state" "operation" {
  backend = "s3"

  config = {
    region  = "ap-northeast-1"
    bucket  = "risuwork-tofu-state-test"
    key     = "operation/tofu.tfstate"
    profile = "risuwork"
  }
}

resource "aws_vpc_peering_connection" "main" {
  vpc_id        = aws_vpc.main.id
  peer_vpc_id   = data.terraform_remote_state.operation.outputs.vpc_id
  peer_owner_id = data.terraform_remote_state.operation.outputs.account_id
}

resource "aws_route" "peering" {
  route_table_id            = aws_route_table.private.id
  destination_cidr_block    = "10.0.0.0/22"
  vpc_peering_connection_id = aws_vpc_peering_connection.main.id
}
