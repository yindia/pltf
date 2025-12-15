# aws_iam_policy

Defines IAM policies (inline or managed) to attach to roles/users created by other modules.

## What it does

- Creates a standalone IAM policy from a JSON document for reuse.

## Fields

Name | Description | Default | Required
--- | --- | --- | ---
file | Json file path containing the Policy |  | False

## Outputs

Name | Description
--- | ---
policy_arn | IAM policy ARN.
policy_id | IAM policy ID.
policy_name | IAM policy name.

