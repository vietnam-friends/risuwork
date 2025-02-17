data "aws_ssoadmin_instances" "main" {}

data "aws_caller_identity" "main" {}

locals {
  identity_store_id = tolist(data.aws_ssoadmin_instances.main.identity_store_ids)[0]
  instance_arn      = tolist(data.aws_ssoadmin_instances.main.arns)[0]
}

// 運営担当
resource "aws_identitystore_user" "operators" {
  for_each          = { for operator in local.operators : operator.email => operator }
  identity_store_id = local.identity_store_id
  display_name      = "${each.value.first_name} ${each.value.last_name}"
  user_name         = each.value.email

  name {
    family_name = each.value.last_name
    given_name  = each.value.first_name
  }

  emails {
    value = each.value.email
  }
}

resource "aws_identitystore_group" "operators" {
  identity_store_id = local.identity_store_id
  display_name      = "admin"
}

resource "aws_identitystore_group_membership" "operators" {
  for_each          = { for operator in local.operators : operator.email => operator }
  identity_store_id = local.identity_store_id
  group_id          = aws_identitystore_group.operators.group_id
  member_id         = aws_identitystore_user.operators[each.key].user_id
}

resource "aws_ssoadmin_permission_set" "operators" {
  instance_arn     = local.instance_arn
  name             = "AdminUser"
  session_duration = "PT12H"
}

resource "aws_ssoadmin_managed_policy_attachment" "operators" {
  instance_arn       = local.instance_arn
  managed_policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
  permission_set_arn = aws_ssoadmin_permission_set.operators.arn
}

resource "aws_ssoadmin_account_assignment" "operators" {
  for_each           = aws_organizations_account.main
  instance_arn       = local.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.operators.arn
  principal_id       = aws_identitystore_group.operators.group_id
  principal_type     = "GROUP"
  target_id          = each.value.id
  target_type        = "AWS_ACCOUNT"
}

resource "aws_ssoadmin_account_assignment" "root" {
  instance_arn       = local.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.operators.arn
  principal_id       = aws_identitystore_group.operators.group_id
  principal_type     = "GROUP"
  target_id          = data.aws_caller_identity.main.account_id
  target_type        = "AWS_ACCOUNT"
}

// 参加者
resource "aws_identitystore_user" "participants" {
  for_each          = local.users
  identity_store_id = local.identity_store_id
  display_name      = "${each.value.user.first_name} ${each.value.user.last_name}"
  user_name         = each.value.user.email

  name {
    family_name = each.value.user.last_name
    given_name  = each.value.user.first_name
  }

  emails {
    value = each.value.user.email
  }
}

resource "aws_identitystore_group" "participants" {
  for_each          = local.teams
  identity_store_id = local.identity_store_id
  display_name      = each.key
}

resource "aws_identitystore_group_membership" "participants" {
  for_each          = local.users
  identity_store_id = local.identity_store_id
  group_id          = aws_identitystore_group.participants[each.value.team].group_id
  member_id         = aws_identitystore_user.participants[each.key].user_id
}

resource "aws_ssoadmin_permission_set" "participants" {
  instance_arn     = local.instance_arn
  name             = "ParticipantsUser"
  session_duration = "PT12H"
}

// TODO: consider policy
resource "aws_ssoadmin_managed_policy_attachment" "participants" {
  instance_arn       = local.instance_arn
  managed_policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess"
  permission_set_arn = aws_ssoadmin_permission_set.participants.arn
}

data "aws_iam_policy_document" "participants_deny" {
  statement {
    sid = "DenyPolicy"

    actions = [
      "ec2:CreateVpc",
      "ec2:CreateNatGateway",
      "ec2:RunInstances",
      "ecs:CreateCluster",
      "elasticloadbalancing:CreateLoadBalancer",
      "rds:CreateDBCluster",
      "rds:CreateDBInstance",
      "elasticache:CreateCacheCluster",
      "elasticache:CreateServerlessCache"
    ]
    effect = "Deny"

    resources = [
      "*",
    ]
  }
}

resource "aws_ssoadmin_permission_set_inline_policy" "participants" {
  instance_arn       = local.instance_arn
  inline_policy      = data.aws_iam_policy_document.participants_deny.json
  permission_set_arn = aws_ssoadmin_permission_set.participants.arn
}

resource "aws_ssoadmin_account_assignment" "participants" {
  for_each           = aws_organizations_account.main
  instance_arn       = local.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.participants.arn
  principal_id       = aws_identitystore_group.participants[each.value.name].group_id
  principal_type     = "GROUP"
  target_id          = each.value.id
  target_type        = "AWS_ACCOUNT"
}
