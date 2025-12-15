# aws_iam_role

Creates IAM roles with inline/managed policies and OIDC trust for Kubernetes service accounts (IRSA).

## What it does

- Creates an IAM role with inline policy and optional managed policies.
- Supports IRSA/OIDC trust for Kubernetes service accounts and trust for other IAM principals.
- Auto-generates least-privilege policies from `links` when used with supported modules.

## Fields

Name | Description | Default | Required
--- | --- | --- | ---
allowed_iams |  | [] | False
allowed_k8s_services |  | [] | False
extra_iam_policies |  | [] | False
iam_policy |  |  | True
kubernetes_trusts |  | [] | False
links |  | [] | False

## Outputs

Name | Description
--- | ---
role_arn | IAM role ARN.

