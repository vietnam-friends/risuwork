data "aws_organizations_organization" "main" {}

resource "aws_organizations_policy" "main" {
  for_each = fileset(path.root, "files/scp/*.json")
  name     = trimprefix(trimsuffix(each.value, ".json"), "files/scp/")
  content  = file(each.value)
}

resource "aws_organizations_policy_attachment" "main" {
  for_each  = aws_organizations_policy.main
  policy_id = each.value.id
  target_id = data.aws_organizations_organization.main.roots[0].id
}
