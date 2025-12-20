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
| [aws_iam_policy.vanilla_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) | resource |
| [aws_iam_role.role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role_policy.pass_role_to_self](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy) | resource |
| [aws_iam_role_policy_attachment.extra_policies_attachment](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_iam_role_policy_attachment.vanilla_role_attachment](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_iam_policy_document.iam_trusts](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.pass_role_to_self](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_allowed_iams"></a> [allowed\_iams](#input\_allowed\_iams) | n/a | `list(string)` | `[]` | no |
| <a name="input_allowed_k8s_services"></a> [allowed\_k8s\_services](#input\_allowed\_k8s\_services) | allowed kubernetes services | `any` | `[]` | no |
| <a name="input_env_name"></a> [env\_name](#input\_env\_name) | Env name | `string` | n/a | yes |
| <a name="input_extra_iam_policies"></a> [extra\_iam\_policies](#input\_extra\_iam\_policies) | n/a | `list(string)` | `[]` | no |
| <a name="input_iam_policy"></a> [iam\_policy](#input\_iam\_policy) | iam policy | `any` | `null` | no |
| <a name="input_kubernetes_trusts"></a> [kubernetes\_trusts](#input\_kubernetes\_trusts) | n/a | <pre>list(object({<br/>    open_id_url  = string<br/>    open_id_arn  = string<br/>    service_name = string<br/>    namespace    = string<br/>  }))</pre> | `[]` | no |
| <a name="input_layer_name"></a> [layer\_name](#input\_layer\_name) | Layer name | `string` | n/a | yes |
| <a name="input_links"></a> [links](#input\_links) | Links for module | `any` | `[]` | no |
| <a name="input_module_name"></a> [module\_name](#input\_module\_name) | Module name | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_role_arn"></a> [role\_arn](#output\_role\_arn) | n/a |
<!-- END_TF_DOCS -->