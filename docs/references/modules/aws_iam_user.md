# aws_iam_user

Creates IAM users with optional access keys and managed/inline policy attachments.

## What it does

- Creates an IAM user with inline/managed policies.
- Optionally auto-generates policies from `links` and returns access keys if created.

## Fields

Name | Description | Default | Required
--- | --- | --- | ---
extra_iam_policies |  | [] | False
iam_policy |  |  | True
links |  | [] | False

## Outputs

Name | Description
--- | ---
user_arn | IAM user ARN.

