output "account_id" {
  value = { for account in aws_organizations_account.main : account.name => account.id }
}
