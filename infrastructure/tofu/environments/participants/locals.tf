locals {
  team      = yamldecode(file("../../files/teams.yml"))[terraform.workspace]
  operators = yamldecode(file("../../files/operators.yml"))
  secrets   = sensitive(yamldecode(file("../../files/secrets.yml")))
}
