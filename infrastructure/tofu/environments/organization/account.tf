resource "aws_organizations_account" "main" {
  for_each          = local.teams
  name              = each.key
  email             = format("%s+%s@%s", local.common.primary_email.local, each.key, local.common.primary_email.domain)
  close_on_deletion = true
}
