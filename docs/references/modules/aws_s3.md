# aws_s3

Creates an S3 bucket with encryption, versioning, lifecycle/replication options, and optional bucket policies.

## What it does

- Creates an encrypted S3 bucket (AES-256) with block-public-access by default.
- Supports custom bucket policy, CORS rules, and optional same-region replication.
- Optionally uploads static files with content-type detection and creates an OAI for CloudFront reads when needed.
- Can emit access logs to the provided log bucket.

## Fields

Name | Description | Default | Required
--- | --- | --- | ---
block_public |  | True | False
bucket_name |  |  | True
bucket_policy |  |  | False
cors_rule | CORS configuration for the bucket. |  | False
files |  |  | False
s3_log_bucket_name |  |  | False
same_region_replication |  | False | False

## Outputs

Name | Description
--- | ---
bucket_arn | Bucket ARN.
bucket_id | Bucket name/ID.
cloudfront_read_path | Origin access identity path (if created).

