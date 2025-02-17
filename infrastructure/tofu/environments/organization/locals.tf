locals {
  secrets   = sensitive(yamldecode(file("../../files/secrets.yml")))
  common    = yamldecode(file("files/common.yml"))
  teams     = yamldecode(file("../../files/teams.yml"))
  operators = yamldecode(file("../../files/operators.yml"))

  # 同一ユーザが複数チームには紐づかない前提
  users = merge([
    for team_name, team in local.teams : {
      for member in team.members : member.email => {
        team = team_name
        user = member
      }
    }
  ]...)
}
