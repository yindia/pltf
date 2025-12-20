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
| [aws_autoscaling_group_tag.node_group](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/autoscaling_group_tag) | resource |
| [aws_eks_node_group.node_group](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_node_group) | resource |
| [aws_iam_role.node_group](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role) | resource |
| [aws_iam_role_policy_attachment.node_group_AmazonEC2ContainerRegistryReadOnly](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_iam_role_policy_attachment.node_group_AmazonEKSWorkerNodePolicy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [aws_iam_role_policy_attachment.node_group_AmazonEKS_CNI_Policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) | resource |
| [random_id.key_suffix](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/id) | resource |
| [aws_eks_cluster.main](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/eks_cluster) | data source |
| [aws_iam_policy_document.node_group](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) | data source |
| [aws_region.current](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/region) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_ami_type"></a> [ami\_type](#input\_ami\_type) | n/a | `string` | `"AL2023_x86_64_STANDARD"` | no |
| <a name="input_autoscaling_tags"></a> [autoscaling\_tags](#input\_autoscaling\_tags) | n/a | `map(string)` | `{}` | no |
| <a name="input_env_name"></a> [env\_name](#input\_env\_name) | Env name | `string` | n/a | yes |
| <a name="input_k8s_cluster_name"></a> [k8s\_cluster\_name](#input\_k8s\_cluster\_name) | n/a | `string` | n/a | yes |
| <a name="input_labels"></a> [labels](#input\_labels) | n/a | `map(string)` | `{}` | no |
| <a name="input_layer_name"></a> [layer\_name](#input\_layer\_name) | Layer name | `string` | n/a | yes |
| <a name="input_max_nodes"></a> [max\_nodes](#input\_max\_nodes) | n/a | `number` | `15` | no |
| <a name="input_min_nodes"></a> [min\_nodes](#input\_min\_nodes) | n/a | `number` | `3` | no |
| <a name="input_module_name"></a> [module\_name](#input\_module\_name) | Module name | `string` | n/a | yes |
| <a name="input_node_disk_size"></a> [node\_disk\_size](#input\_node\_disk\_size) | n/a | `number` | `20` | no |
| <a name="input_node_instance_type"></a> [node\_instance\_type](#input\_node\_instance\_type) | n/a | `string` | `"t3.medium"` | no |
| <a name="input_spot_instances"></a> [spot\_instances](#input\_spot\_instances) | n/a | `bool` | `false` | no |
| <a name="input_taints"></a> [taints](#input\_taints) | n/a | `list(map(string))` | `[]` | no |
| <a name="input_use_gpu"></a> [use\_gpu](#input\_use\_gpu) | tflint-ignore: terraform\_unused\_declarations | `bool` | `false` | no |

## Outputs

No outputs.
<!-- END_TF_DOCS -->