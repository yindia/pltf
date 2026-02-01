<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_kubernetes"></a> [kubernetes](#requirement\_kubernetes) | >= 1.13.3 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | n/a |
| <a name="provider_helm"></a> [helm](#provider\_helm) | n/a |
| <a name="provider_random"></a> [random](#provider\_random) | n/a |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [aws_ecr_lifecycle_policy.repo_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecr_lifecycle_policy) | resource |
| [aws_ecr_repository.repo](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecr_repository) | resource |
| [aws_ecr_repository_policy.repo_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecr_repository_policy) | resource |
| [aws_iam_policy.k8s_service](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) | resource |
| [aws_iam_role.k8s_service](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role_policy_attachment.extra_policies_attachment](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_iam_role_policy_attachment.vanilla_role_attachment](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [helm_release.k8s-service](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release) | resource |
| [random_string.repo_name_hash](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/string) | resource |
| [aws_caller_identity.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/caller_identity) | data source |
| [aws_ecr_image.service_image](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/ecr_image) | data source |
| [aws_iam_policy_document.repo_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_iam_policy_document.trust_k8s_openid](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_additional_iam_policies"></a> [additional\_iam\_policies](#input\_additional\_iam\_policies) | n/a | `list(string)` | `[]` | no |
| <a name="input_args"></a> [args](#input\_args) | n/a | `list(string)` | n/a | yes |
| <a name="input_autoscaling_target_cpu_percentage"></a> [autoscaling\_target\_cpu\_percentage](#input\_autoscaling\_target\_cpu\_percentage) | Percentage of requested cpu after which autoscaling kicks in | `number` | `80` | no |
| <a name="input_autoscaling_target_mem_percentage"></a> [autoscaling\_target\_mem\_percentage](#input\_autoscaling\_target\_mem\_percentage) | Percentage of requested memory after which autoscaling kicks in | `number` | `80` | no |
| <a name="input_commands"></a> [commands](#input\_commands) | n/a | `list(string)` | n/a | yes |
| <a name="input_consistent_hash"></a> [consistent\_hash](#input\_consistent\_hash) | n/a | `string` | `null` | no |
| <a name="input_cron_jobs"></a> [cron\_jobs](#input\_cron\_jobs) | n/a | `list` | `[]` | no |
| <a name="input_digest"></a> [digest](#input\_digest) | Digest of image to be deployed | `string` | `null` | no |
| <a name="input_domain"></a> [domain](#input\_domain) | n/a | `string` | `""` | no |
| <a name="input_env_name"></a> [env\_name](#input\_env\_name) | Env name | `string` | n/a | yes |
| <a name="input_env_vars"></a> [env\_vars](#input\_env\_vars) | Environment variables to pass to the container | <pre>list(object({<br/>    name  = string<br/>    value = string<br/>  }))</pre> | `[]` | no |
| <a name="input_healthcheck_command"></a> [healthcheck\_command](#input\_healthcheck\_command) | n/a | `list(string)` | n/a | yes |
| <a name="input_healthcheck_path"></a> [healthcheck\_path](#input\_healthcheck\_path) | n/a | `string` | `null` | no |
| <a name="input_http_port"></a> [http\_port](#input\_http\_port) | The port that exposes an HTTP interface | `any` | `null` | no |
| <a name="input_iam_policy"></a> [iam\_policy](#input\_iam\_policy) | n/a | `any` | n/a | yes |
| <a name="input_image"></a> [image](#input\_image) | External Image to be deployed | `string` | n/a | yes |
| <a name="input_ingress_extra_annotations"></a> [ingress\_extra\_annotations](#input\_ingress\_extra\_annotations) | n/a | `map(string)` | `{}` | no |
| <a name="input_initial_liveness_delay"></a> [initial\_liveness\_delay](#input\_initial\_liveness\_delay) | n/a | `number` | `30` | no |
| <a name="input_initial_readiness_delay"></a> [initial\_readiness\_delay](#input\_initial\_readiness\_delay) | n/a | `number` | `30` | no |
| <a name="input_keep_path_prefix"></a> [keep\_path\_prefix](#input\_keep\_path\_prefix) | n/a | `bool` | `false` | no |
| <a name="input_layer_name"></a> [layer\_name](#input\_layer\_name) | Layer name | `string` | n/a | yes |
| <a name="input_link_secrets"></a> [link\_secrets](#input\_link\_secrets) | n/a | `list(map(string))` | `[]` | no |
| <a name="input_links"></a> [links](#input\_links) | n/a | `any` | `null` | no |
| <a name="input_liveness_probe_command"></a> [liveness\_probe\_command](#input\_liveness\_probe\_command) | n/a | `list(string)` | n/a | yes |
| <a name="input_liveness_probe_path"></a> [liveness\_probe\_path](#input\_liveness\_probe\_path) | Url path for liveness probe | `string` | `null` | no |
| <a name="input_max_containers"></a> [max\_containers](#input\_max\_containers) | Max value for HPA autoscaling | `string` | `3` | no |
| <a name="input_max_history"></a> [max\_history](#input\_max\_history) | n/a | `number` | n/a | yes |
| <a name="input_min_containers"></a> [min\_containers](#input\_min\_containers) | Min value for HPA autoscaling | `string` | `1` | no |
| <a name="input_module_name"></a> [module\_name](#input\_module\_name) | Module name | `string` | n/a | yes |
| <a name="input_openid_provider_arn"></a> [openid\_provider\_arn](#input\_openid\_provider\_arn) | n/a | `string` | n/a | yes |
| <a name="input_openid_provider_url"></a> [openid\_provider\_url](#input\_openid\_provider\_url) | n/a | `string` | n/a | yes |
| <a name="input_persistent_storage"></a> [persistent\_storage](#input\_persistent\_storage) | n/a | `list(map(string))` | `[]` | no |
| <a name="input_pod_annotations"></a> [pod\_annotations](#input\_pod\_annotations) | values to add to the pod annotations for the k8s-service pods | `map(string)` | `{}` | no |
| <a name="input_pod_labels"></a> [pod\_labels](#input\_pod\_labels) | n/a | `map(string)` | n/a | yes |
| <a name="input_ports"></a> [ports](#input\_ports) | Ports to be exposed | `list(any)` | n/a | yes |
| <a name="input_probe_port"></a> [probe\_port](#input\_probe\_port) | The port that is used for health probes | `any` | `null` | no |
| <a name="input_public_uri"></a> [public\_uri](#input\_public\_uri) | n/a | `list(string)` | `[]` | no |
| <a name="input_readiness_probe_command"></a> [readiness\_probe\_command](#input\_readiness\_probe\_command) | n/a | `list(string)` | n/a | yes |
| <a name="input_readiness_probe_path"></a> [readiness\_probe\_path](#input\_readiness\_probe\_path) | Url path for readiness probe | `string` | `null` | no |
| <a name="input_resource_limits"></a> [resource\_limits](#input\_resource\_limits) | n/a | `map(any)` | n/a | yes |
| <a name="input_resource_request"></a> [resource\_request](#input\_resource\_request) | n/a | `map(any)` | <pre>{<br/>  "cpu": 100,<br/>  "memory": 128<br/>}</pre> | no |
| <a name="input_secrets"></a> [secrets](#input\_secrets) | n/a | `any` | `null` | no |
| <a name="input_service_annotations"></a> [service\_annotations](#input\_service\_annotations) | Annotations to add to the service resource | `map(string)` | `{}` | no |
| <a name="input_sticky_session"></a> [sticky\_session](#input\_sticky\_session) | n/a | `bool` | `false` | no |
| <a name="input_sticky_session_max_age"></a> [sticky\_session\_max\_age](#input\_sticky\_session\_max\_age) | n/a | `number` | `86400` | no |
| <a name="input_tag"></a> [tag](#input\_tag) | Tag of image to be deployed | `string` | `null` | no |
| <a name="input_timeout"></a> [timeout](#input\_timeout) | n/a | `number` | `300` | no |
| <a name="input_tolerations"></a> [tolerations](#input\_tolerations) | n/a | `list(map(string))` | `[]` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_current_digest"></a> [current\_digest](#output\_current\_digest) | n/a |
| <a name="output_current_image"></a> [current\_image](#output\_current\_image) | n/a |
| <a name="output_current_tag"></a> [current\_tag](#output\_current\_tag) | n/a |
| <a name="output_docker_repo_url"></a> [docker\_repo\_url](#output\_docker\_repo\_url) | n/a |
<!-- END_TF_DOCS -->