# aws_base

Provision networking (VPC), subnets across AZs, flow logs, NAT, and a default KMS key + log bucket for the environment.

## What it does

- Creates a new VPC (or imports an existing one) with public/private subnets across three AZs.
- Adds internet/NAT gateways and route tables for public/private egress.
- Enables VPC flow logs to the log bucket and provisions a default KMS key.
- Creates a log bucket for access/flow logs used by other modules.

## Fields

Name | Description | Default | Required
--- | --- | --- | ---
private_ipv4_cidr_blocks | Cidr blocks for private subnets. One for each desired AZ | ['10.0.128.0/21', '10.0.136.0/21', '10.0.144.0/21'] | False
private_subnet_ids | List of pre-existing private subnets to use instead of creating new subnets for pltf. Required when var.vpc_id is set. |  | False
public_ipv4_cidr_blocks | Cidr blocks for public subnets. One for each desired AZ | ['10.0.0.0/21', '10.0.8.0/21', '10.0.16.0/21'] | False
public_subnet_ids | List of pre-existing public subnets to use instead of creating new subnets for pltf. Required when var.vpc_id is set. |  | False
total_ipv4_cidr_block | Cidr block to reserve for whole vpc | 10.0.0.0/16 | False
vpc_id | The ID of an pre-existing VPC to use instead of creating a new VPC for pltf |  | False
vpc_log_retention |  | 90 | False

## Bring your own VPC

To use an existing VPC, set `vpc_id`, `public_subnet_ids`, and `private_subnet_ids`. Public subnets must route to an internet gateway and assign public IPs. Private subnets must route 0.0.0.0/0 to a NAT gateway with a public IP. Misconfigured routes may yield Terraform errors like "No routes matching supplied arguments found in Route Table". IPv6 imports are not validated; dual-stack may work but is not verified.

## Outputs

Name | Description
--- | ---
kms_account_key_arn | ARN of the default KMS key for environment resources
kms_account_key_id | ID of the default KMS key
private_subnet_ids | Private subnet IDs provisioned/imported
public_nat_ips | Elastic IPs of NAT gateways
public_subnets_ids | Public subnet IDs provisioned/imported
s3_log_bucket_name | Name of the environment log bucket
vpc_id | VPC ID provisioned/imported
