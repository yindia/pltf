<!-- BEGIN_TF_DOCS -->
## Requirements

No requirements.

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | n/a |
| <a name="provider_random"></a> [random](#provider\_random) | n/a |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_iam_policy.policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) | resource |
| [random_string.policy_suffix](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/string) | resource |
| [aws_caller_identity.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity) | data source |
| [aws_region.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/region) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_env_name"></a> [env\_name](#input\_env\_name) | Env name | `string` | n/a | yes |
| <a name="input_file"></a> [file](#input\_file) | Json file path containing the Policy | `string` | `""` | no |
| <a name="input_layer_name"></a> [layer\_name](#input\_layer\_name) | Layer name | `string` | n/a | yes |
| <a name="input_module_name"></a> [module\_name](#input\_module\_name) | Module name | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_policy_arn"></a> [policy\_arn](#output\_policy\_arn) | n/a |
| <a name="output_policy_id"></a> [policy\_id](#output\_policy\_id) | n/a |
| <a name="output_policy_name"></a> [policy\_name](#output\_policy\_name) | n/a |
<!-- END_TF_DOCS -->