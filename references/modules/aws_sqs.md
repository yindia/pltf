# aws_sqs

Creates an SQS queue with encryption, visibility timeout, redrive policy, and optional DLQ linkage.

## What it does

- Creates an SQS queue (standard or FIFO) with a dedicated KMS CMK.
- Configures default queue policy allowing account root, SNS, and EventBridge producers.
- Supports content-based deduplication, delivery delays, retention, and long polling.
- Outputs KMS ARN for wiring IRSA/IAM consumers.

## Fields

Name | Description | Default | Required
--- | --- | --- | ---
content_based_deduplication | Enable content-based deduplication for FIFO queues. | False | False
delay_seconds |  | 0 | False
fifo | Create a FIFO queue (adds .fifo suffix). | False | False
message_retention_seconds |  | 345600 | False
receive_wait_time_seconds |  | 0 | False

## Outputs

Name | Description
--- | ---
kms_arn | KMS key ARN for the queue.
queue_arn | SQS queue ARN.
queue_id | SQS queue URL.
queue_name | SQS queue name.

