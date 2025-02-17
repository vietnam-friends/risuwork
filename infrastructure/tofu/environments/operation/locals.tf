locals {
  operators = yamldecode(file("../../files/operators.yml"))
  teams     = yamldecode(file("../../files/teams.yml"))
  secrets   = sensitive(yamldecode(file("../../files/secrets.yml")))

  repository_url = "https://github.com/risuwork/risuwork"
  portal_domain  = "portal.r-isucon.blue"
}
