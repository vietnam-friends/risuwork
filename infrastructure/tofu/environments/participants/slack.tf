locals {
  user_list = merge(
    { for member in local.team.members : member.email => 1 },
    { for operator in local.operators : operator.email => 1 },
  )
}

data "slack_user" "main" {
  for_each = local.user_list
  email    = each.key
}

resource "slack_conversation" "main" {
  name              = "recruit-isucon-2024-${local.team.name}"
  is_private        = true
  topic             = "運営への問い合わせは <!subteam^S07JMQ7HZJP|@recruit-isucon-2024-admin> まで"
  permanent_members = [for user in data.slack_user.main : user.id]
}
