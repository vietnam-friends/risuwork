data "slack_user" "main" {
  for_each = { for operator in local.operators : operator.email => operator }
  email    = each.key
}

resource "slack_usergroup" "main" {
  name   = "recruit-isucon-2024運営"
  handle = "recruit-isucon-2024-admin"
  users  = [for user in data.slack_user.main : user.id]
}
