# aws_sns

Creates an SNS topic with encryption, delivery policies, and optional subscriptions.

## What it does

- Creates an SNS topic (standard or FIFO) with a dedicated KMS CMK.
- Applies a default topic policy for account root and subscribes provided SQS endpoints.
- Supports content-based deduplication for FIFO topics and custom delivery policy.

## Fields

Name | Description | Default | Required
--- | --- | --- | ---
content_based_deduplication | Enable content-based deduplication for FIFO topics. | False | False
fifo | Create a FIFO topic (adds .fifo suffix). | False | False
sqs_subscribers | List of SQS queue ARNs to subscribe. | [] | False

## Outputs

Name | Description
--- | ---
kms_arn | KMS key ARN for the topic.
topic_arn | SNS topic ARN.
