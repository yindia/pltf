# aws_dynamodb

Creates a DynamoDB table with encryption, throughput settings, TTL, and optional point-in-time recovery.

## What it does

- Creates a DynamoDB table with server-side encryption and customizable billing mode.
- Supports provisioned throughput settings and TTL via attributes.
- Exposes table ARN/ID and KMS key details.

## Fields

Name | Description | Default | Required
--- | --- | --- | ---
attributes |  |  | True
billing_mode |  | PROVISIONED | False
hash_key |  |  | False
range_key |  |  | False
read_capacity |  | 20 | False
write_capacity |  | 20 | False

## Outputs

Name | Description
--- | ---
kms_arn | KMS key ARN used for encryption.
kms_id | KMS key ID used for encryption.
table_arn | Table ARN.
table_id | Table name/ID.

