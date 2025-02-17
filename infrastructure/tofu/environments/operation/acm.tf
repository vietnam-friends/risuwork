resource "aws_acm_certificate" "portal" {
  domain_name       = local.portal_domain
  validation_method = "DNS"

  provider = aws.us_east
}
