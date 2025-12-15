data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

locals {
  templatevars = {
    account_id = data.aws_caller_identity.current.account_id
    arn        = data.aws_caller_identity.current.arn
    region     = data.aws_region.current.name
  }
}

resource "random_string" "policy_suffix" {
  length  = 4
  special = false
  upper   = false
}

resource "aws_iam_policy" "policy" {
  name   = "${var.env_name}-${var.layer_name}-${var.module_name}"
  policy = templatefile(var.file, local.templatevars)
  tags = {
    "pltf-created" : true
  }
}
