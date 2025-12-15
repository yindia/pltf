# aws_ses

Configures SES domain/identities with DNS verification records and optional inbound/notification settings.

## What it does

- Verifies a domain in SES and configures MAIL FROM.
- Creates IAM policy for sending and exposes DKIM tokens and identity ARN.

## Fields

Name | Description | Default | Required
--- | --- | --- | ---
domain |  |  | True
mail_from_prefix |  | mail | False
zone_id |  |  | True

## Outputs

Name | Description
--- | ---
identity_arn | SES identity ARN.
sender_policy_arn | IAM policy ARN permitting SES send.

