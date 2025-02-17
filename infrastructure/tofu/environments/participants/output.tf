output "app_alb_domain" {
  value = aws_lb.main.dns_name
}

output "vpc_cider_block" {
  value = aws_vpc.main.cidr_block
}

output "vpc_peering_id" {
  value = aws_vpc_peering_connection.main.id
}
