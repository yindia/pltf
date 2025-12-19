# aws_eks

Creates an EKS control plane in private subnets, configurable Kubernetes version, logging, and OIDC provider for IRSA.

## What it does

- Provisions the EKS control plane in private subnets with security groups.
- Creates an OIDC provider for IRSA and enables control-plane logging.
- Stands up a default managed nodegroup with scaling/min/max and optional spot.

## Fields

Name | Description | Default | Required
--- | --- | --- | ---
ami_type |  | AL2023_x86_64_STANDARD | False
cluster_name |  |  | True
control_plane_security_groups | List of security groups to give control plane access to | [] | False
eks_log_retention |  | 7 | False
enable_metrics |  |  | True
k8s_version |  | 1.21 | False
kms_account_key_arn |  |  | True
max_nodes |  | 5 | False
min_nodes |  | 3 | False
node_disk_size |  | 20 | False
node_instance_type |  | t3.medium | False
node_launch_template |  | {} | False
private_subnet_ids |  |  | True
spot_instances |  | False | False
vpc_id |  |  | True

## Outputs

Name | Description
--- | ---
k8s_ca_data | Cluster CA data (base64).
k8s_cluster_name | EKS cluster name.
k8s_endpoint | EKS API endpoint.
k8s_node_group_security_id | Security group ID for nodes.
k8s_openid_provider_arn | OIDC provider ARN for IRSA.
k8s_openid_provider_url | OIDC provider URL.
k8s_version | Kubernetes version.

