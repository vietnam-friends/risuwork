resource "aws_lb" "main" {
  name               = "app-alb"
  load_balancer_type = "application"
  internal           = true
  security_groups    = [aws_default_security_group.main.id]
  subnets            = [aws_subnet.private_az1.id, aws_subnet.private_az4.id]
}

resource "aws_lb_target_group" "main" {
  name                              = "app-tg"
  vpc_id                            = aws_vpc.main.id
  target_type                       = "ip"
  port                              = 80
  protocol                          = "HTTP"
  load_balancing_cross_zone_enabled = false
  deregistration_delay              = 5

  health_check {
    path              = "/health"
    healthy_threshold = 2
    interval          = 5
    timeout           = 2
  }
}

resource "aws_lb_listener" "main" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.main.arn
  }
}

# public ALBが必要な場合追加
# resource "aws_lb" "public" {
#   name               = "app-alb-public"
#   load_balancer_type = "application"
#   security_groups    = [aws_default_security_group.main.id]
#   subnets            = [aws_subnet.public_az1.id, aws_subnet.public_az4.id]
# }
#
# resource "aws_lb_target_group" "public" {
#   name                              = "app-tg-public"
#   vpc_id                            = aws_vpc.main.id
#   target_type                       = "ip"
#   port                              = 80
#   protocol                          = "HTTP"
#   load_balancing_cross_zone_enabled = false
#
#   health_check {
#     path              = "/health"
#     healthy_threshold = 2
#     interval          = 5
#     timeout           = 2
#   }
# }
#
# resource "aws_lb_listener" "public" {
#   load_balancer_arn = aws_lb.public.arn
#   port              = "80"
#   protocol          = "HTTP"
#
#   default_action {
#     type             = "forward"
#     target_group_arn = aws_lb_target_group.public.arn
#   }
# }
