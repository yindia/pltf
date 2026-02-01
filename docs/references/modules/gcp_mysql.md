<!-- BEGIN_TF_DOCS -->
## Requirements

No requirements.

## Providers

| Name | Version |
|------|---------|
| <a name="provider_google"></a> [google](#provider\_google) | n/a |
| <a name="provider_random"></a> [random](#provider\_random) | n/a |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [google_sql_database.main](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/sql_database) | resource |
| [google_sql_database_instance.instance](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/sql_database_instance) | resource |
| [google_sql_user.root](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/sql_user) | resource |
| [random_id.key_suffix](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/id) | resource |
| [random_password.root_auth](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/password) | resource |
| [google_compute_network.vpc](https://registry.terraform.io/providers/hashicorp/google/latest/docs/data-sources/compute_network) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_engine_version"></a> [engine\_version](#input\_engine\_version) | n/a | `string` | `"5_7"` | no |
| <a name="input_env_name"></a> [env\_name](#input\_env\_name) | Env name | `string` | n/a | yes |
| <a name="input_instance_tier"></a> [instance\_tier](#input\_instance\_tier) | n/a | `string` | `"db-f1-micro"` | no |
| <a name="input_layer_name"></a> [layer\_name](#input\_layer\_name) | Layer name | `string` | n/a | yes |
| <a name="input_module_name"></a> [module\_name](#input\_module\_name) | Module name | `string` | n/a | yes |
| <a name="input_safety"></a> [safety](#input\_safety) | n/a | `bool` | `false` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_db_host"></a> [db\_host](#output\_db\_host) | n/a |
| <a name="output_db_name"></a> [db\_name](#output\_db\_name) | n/a |
| <a name="output_db_password"></a> [db\_password](#output\_db\_password) | n/a |
| <a name="output_db_user"></a> [db\_user](#output\_db\_user) | This list of outputs needs to be the same as the ones for gcp-postgres and should match the ones at gcp\_k8s\_service.py:handle\_rds\_link() |
<!-- END_TF_DOCS -->