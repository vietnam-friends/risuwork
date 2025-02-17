resource "aws_instance" "ecs_node" {
  ami                    = data.aws_ssm_parameter.ecs_ami.value
  instance_type          = "m5.large"
  subnet_id              = aws_subnet.private_az1.id
  vpc_security_group_ids = [aws_default_security_group.main.id]
  iam_instance_profile   = aws_iam_instance_profile.ecs_node.id

  user_data = base64encode(file("files/user_data.sh"))

  tags = {
    Name = "ecs-node"
  }

  lifecycle {
    // AMIの更新で再作成されるのを回避する
    ignore_changes = [ami]
  }
}

# 作業用インスタンス
resource "aws_instance" "ope" {
  ami                    = data.aws_ssm_parameter.ecs_ami.value
  instance_type          = "t3.small"
  subnet_id              = aws_subnet.private_az1.id
  vpc_security_group_ids = [aws_default_security_group.main.id]
  iam_instance_profile   = aws_iam_instance_profile.ecs_node.id

  tags = {
    Name = "operation"
  }

  lifecycle {
    // AMIの更新で再作成されるのを回避する
    ignore_changes = [ami]
  }
}
