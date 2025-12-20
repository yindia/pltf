<!-- BEGIN_TF_DOCS -->
## Requirements

No requirements.

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | n/a |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_acm_certificate.certificate](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/acm_certificate) | resource |
| [aws_acm_certificate.imported_certificate](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/acm_certificate) | resource |
| [aws_acm_certificate_validation.example_com](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/acm_certificate_validation) | resource |
| [aws_route53_record.validation_record](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route53_record) | resource |
| [aws_route53_zone.public](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route53_zone) | resource |
| [aws_ssm_parameter.certificate_body](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ssm_parameter) | data source |
| [aws_ssm_parameter.certificate_chain](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ssm_parameter) | data source |
| [aws_ssm_parameter.private_key](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ssm_parameter) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_cert_chain_included"></a> [cert\_chain\_included](#input\_cert\_chain\_included) | n/a | `bool` | `false` | no |
| <a name="input_delegated"></a> [delegated](#input\_delegated) | n/a | `bool` | `false` | no |
| <a name="input_domain"></a> [domain](#input\_domain) | n/a | `string` | n/a | yes |
| <a name="input_env_name"></a> [env\_name](#input\_env\_name) | Env name | `string` | n/a | yes |
| <a name="input_external_cert_arn"></a> [external\_cert\_arn](#input\_external\_cert\_arn) | n/a | `string` | `""` | no |
| <a name="input_force_update"></a> [force\_update](#input\_force\_update) | tflint-ignore: terraform\_unused\_declarations | `bool` | `false` | no |
| <a name="input_layer_name"></a> [layer\_name](#input\_layer\_name) | Layer name | `string` | n/a | yes |
| <a name="input_module_name"></a> [module\_name](#input\_module\_name) | Module name | `string` | n/a | yes |
| <a name="input_upload_cert"></a> [upload\_cert](#input\_upload\_cert) | n/a | `bool` | `false` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_cert_arn"></a> [cert\_arn](#output\_cert\_arn) | n/a |
| <a name="output_domain"></a> [domain](#output\_domain) | n/a |
| <a name="output_name_servers"></a> [name\_servers](#output\_name\_servers) | n/a |
| <a name="output_zone_id"></a> [zone\_id](#output\_zone\_id) | n/a |
<!-- END_TF_DOCS -->