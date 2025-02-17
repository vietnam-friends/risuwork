# 事後に一部リポジトリを全参加者に公開するために利用

resource "github_team" "all" {
  name    = "all"
  privacy = "closed"
}

resource "github_membership" "all" {
  for_each = local.users
  username = each.value.user.github_username
  role     = "member"
}

resource "github_team_membership" "all" {
  for_each = local.users
  team_id  = github_team.all.id
  username = each.value.user.github_username
}
