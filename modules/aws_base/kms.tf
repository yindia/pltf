# Make sure you have these:
# data "aws_caller_identity" "current" {}
# data "aws_region" "current" {}

########################################
# 1. Create the Auto Scaling SLR
########################################
resource "aws_iam_service_linked_role" "autoscaling" {
  aws_service_name = "autoscaling.amazonaws.com"
  description      = "Service-linked role for EC2 Auto Scaling used by KMS key policy"
}

########################################
# 2. KMS key policy (unchanged)
########################################
data "aws_iam_policy_document" "kms_policy" {
  statement {
    sid    = "Enable IAM User/Role Permissions"
    effect = "Allow"
    principals {
      identifiers = ["arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"]
      type        = "AWS"
    }
    actions   = ["kms:*"]
    resources = ["*"]
  }

  statement {
    sid    = "Allow service-linked role use of the CMK"
    effect = "Allow"
    principals {
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"
      ]
      type = "AWS"
    }
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:DescribeKey"
    ]
    resources = ["*"]
  }

  statement {
    sid    = "Allow sns access"
    effect = "Allow"
    principals {
      identifiers = ["sns.amazonaws.com"]
      type        = "Service"
    }
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey*"
    ]
    resources = ["*"]
  }

  statement {
    sid    = "Allow attachment of persistent resources"
    effect = "Allow"
    principals {
      identifiers = [
        "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"
      ]
      type = "AWS"
    }
    actions = [
      "kms:CreateGrant",
      "kms:ListGrants",
      "kms:RevokeGrant"
    ]
    resources = ["*"]
    condition {
      test     = "Bool"
      values   = [true]
      variable = "kms:GrantIsForAWSResource"
    }
  }

  statement {
    sid    = "Allow events access"
    effect = "Allow"
    principals {
      identifiers = ["events.amazonaws.com"]
      type        = "Service"
    }
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey*"
    ]
    resources = ["*"]
  }

  statement {
    sid    = "Allow s3 access"
    effect = "Allow"
    principals {
      identifiers = ["s3.amazonaws.com"]
      type        = "Service"
    }
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey*"
    ]
    resources = ["*"]
  }

  statement {
    sid    = "Allow cloudwatch log access"
    effect = "Allow"
    principals {
      identifiers = ["logs.${data.aws_region.current.name}.amazonaws.com"]
      type        = "Service"
    }
    actions = [
      "kms:Encrypt*",
      "kms:Decrypt*",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:Describe*"
    ]
    resources = ["*"]
  }
}

########################################
# 3. KMS key with explicit dependency
########################################
resource "aws_kms_key" "key" {
  description         = "Base key for account"
  policy              = data.aws_iam_policy_document.kms_policy.json
  enable_key_rotation = true

  # Ensure the SLR exists before KMS key is created
  depends_on = [aws_iam_service_linked_role.autoscaling]

  tags = {
    Name      = "pltf-${var.layer_name}"
    terraform = "true"
  }
}

resource "aws_kms_alias" "alias" {
  name          = "alias/pltf-${var.layer_name}"
  target_key_id = aws_kms_key.key.key_id
}
