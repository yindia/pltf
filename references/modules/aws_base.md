<!-- BEGIN_TF_DOCS -->
# aws_base

Provision networking (VPC), subnets across AZs, flow logs, NAT, and a default KMS key + log bucket for the environment.

## What it does

- Creates a new VPC (or imports an existing one) with public/private subnets across three AZs.
- Adds internet/NAT gateways and route tables for public/private egress.
- Enables VPC flow logs to the log bucket and provisions a default KMS key.
- Creates a log bucket for access/flow logs used by other modules.

## Bring your own VPC

To use an existing VPC, set `vpc_id`, `public_subnet_ids`, and `private_subnet_ids`. Public subnets must route to an internet gateway and assign public IPs. Private subnets must route 0.0.0.0/0 to a NAT gateway with a public IP. Misconfigured routes may yield Terraform errors like "No routes matching supplied arguments found in Route Table". IPv6 imports are not validated; dual-stack may work but is not verified.

#### Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider_aws) | 6.27.0 |
| <a name="provider_random"></a> [random](#provider_random) | 3.7.2 |

#### Modules

No modules.

#### Resources

| Name | Type |
|------|------|
| [aws_cloudwatch_log_group.vpc_flow_log](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_log_group) | resource |
| [aws_db_subnet_group.main](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/db_subnet_group) | resource |
| [aws_default_security_group.default](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/default_security_group) | resource |
| [aws_docdb_subnet_group.main](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/docdb_subnet_group) | resource |
| [aws_ebs_encryption_by_default.default](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ebs_encryption_by_default) | resource |
| [aws_eip.nat_eips](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eip) | resource |
| [aws_elasticache_subnet_group.main](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/elasticache_subnet_group) | resource |
| [aws_flow_log.vpc](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/flow_log) | resource |
| [aws_iam_role.vpc_flow_log](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role_policy.vpc_flow_log](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy) | resource |
| [aws_iam_service_linked_role.autoscaling](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_service_linked_role) | resource |
| [aws_internet_gateway.igw](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/internet_gateway) | resource |
| [aws_kms_alias.alias](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_alias) | resource |
| [aws_kms_key.key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/kms_key) | resource |
| [aws_nat_gateway.nat_gateways](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/nat_gateway) | resource |
| [aws_route.nat_routes](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route) | resource |
| [aws_route_table.private_route_tables](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route_table) | resource |
| [aws_route_table.public_route_table](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route_table) | resource |
| [aws_route_table_association.private_associations](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route_table_association) | resource |
| [aws_route_table_association.public_association](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route_table_association) | resource |
| [aws_s3_bucket.log_bucket](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket) | resource |
| [aws_s3_bucket_acl.log_bucket](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_acl) | resource |
| [aws_s3_bucket_lifecycle_configuration.log_bucket](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_lifecycle_configuration) | resource |
| [aws_s3_bucket_ownership_controls.log_bucket](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_ownership_controls) | resource |
| [aws_s3_bucket_policy.log_bucket_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_policy) | resource |
| [aws_s3_bucket_public_access_block.log_bucket](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_public_access_block) | resource |
| [aws_s3_bucket_server_side_encryption_configuration.log_bucket](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_server_side_encryption_configuration) | resource |
| [aws_s3_bucket_versioning.log_bucket](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_versioning) | resource |
| [aws_security_group.db](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group) | resource |
| [aws_security_group.documentdb](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group) | resource |
| [aws_security_group.elasticache](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group) | resource |
| [aws_subnet.private_subnets](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/subnet) | resource |
| [aws_subnet.public_subnets](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/subnet) | resource |
| [aws_vpc.vpc](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/vpc) | resource |
| [aws_vpc_endpoint.s3](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/vpc_endpoint) | resource |
| [aws_vpc_endpoint_route_table_association.s3](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/vpc_endpoint_route_table_association) | resource |
| [random_id.bucket_suffix](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/id) | resource |
| [random_id.vpc_flow_log_suffix](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/id) | resource |

#### Inputs

| Name | Description | Type |
|------|-------------|------|
| <a name="input_env_name"></a> [env_name](#input_env_name) | Env name | `string` |
| <a name="input_layer_name"></a> [layer_name](#input_layer_name) | Layer name | `string` |
| <a name="input_module_name"></a> [module_name](#input_module_name) | Module name | `string` |
| <a name="input_private_ipv4_cidr_blocks"></a> [private_ipv4_cidr_blocks](#input_private_ipv4_cidr_blocks) | Cidr blocks for private subnets. One for each desired AZ | `list(string)` |
| <a name="input_private_subnet_ids"></a> [private_subnet_ids](#input_private_subnet_ids) | List of pre-existing private subnets to use instead of creating new subnets for pltf. Required when var.vpc_id is set. | `list(string)` |
| <a name="input_public_ipv4_cidr_blocks"></a> [public_ipv4_cidr_blocks](#input_public_ipv4_cidr_blocks) | Cidr blocks for public subnets. One for each desired AZ | `list(string)` |
| <a name="input_public_subnet_ids"></a> [public_subnet_ids](#input_public_subnet_ids) | List of pre-existing public subnets to use instead of creating new subnets for pltf. Required when var.vpc_id is set. | `list(string)` |
| <a name="input_total_ipv4_cidr_block"></a> [total_ipv4_cidr_block](#input_total_ipv4_cidr_block) | Cidr block to reserve for whole vpc | `string` |
| <a name="input_vpc_id"></a> [vpc_id](#input_vpc_id) | The ID of an pre-existing VPC to use instead of creating a new VPC for pltf | `string` |
| <a name="input_vpc_log_retention"></a> [vpc_log_retention](#input_vpc_log_retention) | n/a | `number` |

#### Outputs

| Name | Description |
|------|-------------|
| <a name="output_db_aws_security_group"></a> [db_aws_security_group](#output_db_aws_security_group) | n/a |
| <a name="output_documentdb_aws_security_group"></a> [documentdb_aws_security_group](#output_documentdb_aws_security_group) | n/a |
| <a name="output_elasticache_aws_security_group"></a> [elasticache_aws_security_group](#output_elasticache_aws_security_group) | n/a |
| <a name="output_kms_account_key_arn"></a> [kms_account_key_arn](#output_kms_account_key_arn) | n/a |
| <a name="output_kms_account_key_id"></a> [kms_account_key_id](#output_kms_account_key_id) | n/a |
| <a name="output_kms_key_alias"></a> [kms_key_alias](#output_kms_key_alias) | n/a |
| <a name="output_private_subnet_ids"></a> [private_subnet_ids](#output_private_subnet_ids) | n/a |
| <a name="output_public_nat_ips"></a> [public_nat_ips](#output_public_nat_ips) | n/a |
| <a name="output_public_subnets_ids"></a> [public_subnets_ids](#output_public_subnets_ids) | n/a |
| <a name="output_s3_log_bucket_name"></a> [s3_log_bucket_name](#output_s3_log_bucket_name) | n/a |
| <a name="output_vpc_id"></a> [vpc_id](#output_vpc_id) | n/a |
<!-- END_TF_DOCS -->