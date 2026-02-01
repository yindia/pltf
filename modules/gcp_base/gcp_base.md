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
| [google_compute_address.address](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_address) | resource |
| [google_compute_firewall.firewall](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_firewall) | resource |
| [google_compute_firewall.k8s_extra_rules](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_firewall) | resource |
| [google_compute_global_address.private_ip_alloc](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_global_address) | resource |
| [google_compute_network.vpc](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_network) | resource |
| [google_compute_route.default](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_route) | resource |
| [google_compute_router.router](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_router) | resource |
| [google_compute_router_nat.nat](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_router_nat) | resource |
| [google_compute_subnetwork.project_subnet](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_subnetwork) | resource |
| [google_container_registry.registry](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/container_registry) | resource |
| [google_kms_crypto_key.key](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/kms_crypto_key) | resource |
| [google_kms_crypto_key_iam_member.gcs](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/kms_crypto_key_iam_member) | resource |
| [google_kms_crypto_key_iam_member.gke](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/kms_crypto_key_iam_member) | resource |
| [google_kms_key_ring.keyring](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/kms_key_ring) | resource |
| [google_secret_manager_secret.kms_secret](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/secret_manager_secret) | resource |
| [google_secret_manager_secret_version.secret-version-basic](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/secret_manager_secret_version) | resource |
| [google_service_networking_connection.private_services](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/service_networking_connection) | resource |
| [random_id.key_suffix](https://registry.terraform.io/providers/hashicorp/random/latest/docs/resources/id) | resource |
| [google_client_config.current](https://registry.terraform.io/providers/hashicorp/google/latest/docs/data-sources/client_config) | data source |
| [google_project.current](https://registry.terraform.io/providers/hashicorp/google/latest/docs/data-sources/project) | data source |
| [google_storage_project_service_account.gcs_account](https://registry.terraform.io/providers/hashicorp/google/latest/docs/data-sources/storage_project_service_account) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_cluster_ipv4_cidr_block"></a> [cluster\_ipv4\_cidr\_block](#input\_cluster\_ipv4\_cidr\_block) | n/a | `string` | `"10.0.32.0/19"` | no |
| <a name="input_env_name"></a> [env\_name](#input\_env\_name) | Env name | `string` | n/a | yes |
| <a name="input_k8s_master_ipv4_cidr_block"></a> [k8s\_master\_ipv4\_cidr\_block](#input\_k8s\_master\_ipv4\_cidr\_block) | n/a | `string` | `"10.0.80.0/28"` | no |
| <a name="input_layer_name"></a> [layer\_name](#input\_layer\_name) | Layer name | `string` | n/a | yes |
| <a name="input_module_name"></a> [module\_name](#input\_module\_name) | Module name | `string` | n/a | yes |
| <a name="input_private_ipv4_cidr_block"></a> [private\_ipv4\_cidr\_block](#input\_private\_ipv4\_cidr\_block) | Cidr block for private subnet. Don't need to worry about AZs in GCP | `string` | `"10.0.0.0/19"` | no |
| <a name="input_services_ipv4_cidr_block"></a> [services\_ipv4\_cidr\_block](#input\_services\_ipv4\_cidr\_block) | n/a | `string` | `"10.0.64.0/20"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_k8s_master_ipv4_cidr_block"></a> [k8s\_master\_ipv4\_cidr\_block](#output\_k8s\_master\_ipv4\_cidr\_block) | n/a |
| <a name="output_kms_account_key_id"></a> [kms\_account\_key\_id](#output\_kms\_account\_key\_id) | n/a |
| <a name="output_private_subnet_id"></a> [private\_subnet\_id](#output\_private\_subnet\_id) | n/a |
| <a name="output_private_subnet_self_link"></a> [private\_subnet\_self\_link](#output\_private\_subnet\_self\_link) | n/a |
| <a name="output_public_nat_ips"></a> [public\_nat\_ips](#output\_public\_nat\_ips) | n/a |
| <a name="output_vpc_id"></a> [vpc\_id](#output\_vpc\_id) | n/a |
| <a name="output_vpc_self_link"></a> [vpc\_self\_link](#output\_vpc\_self\_link) | n/a |
<!-- END_TF_DOCS -->