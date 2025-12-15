# aws_k8s_base

Bootstraps core Kubernetes add-ons on EKS: ingress, autoscaler, metrics server, external-dns, and optional admins.

## What it does

- Deploys ingress-nginx (optionally HA) and configures extra TCP ports if set.
- Deploys external-dns, cert-manager, metrics-server, and cluster-autoscaler.
- Optionally deploys Linkerd service mesh and grants admin access via admin_arns.
- Wires ingress/records to the provided domain and hosted zone.

## Fields

Name | Description | Default | Required
--- | --- | --- | ---
admin_arns |  | [] | False
cert_arn |  |  | False
cert_manager_values |  | {} | False
certificate_body |  |  | False
certificate_chain |  |  | False
domain |  |  | True
eks_cluster_name |  |  | True
enable_auto_dns |  |  | True
expose_self_signed_ssl |  | False | False
ingress_nginx_values |  | {} | False
k8s_cluster_name |  |  | True
k8s_version |  |  | True
linkerd_enabled |  | True | False
linkerd_high_availability |  | False | False
linkerd_values |  | {} | False
nginx_config |  | {} | False
nginx_enabled |  |  | True
nginx_extra_tcp_ports |  | {} | False
nginx_extra_tcp_ports_tls |  | [] | False
nginx_high_availability |  | False | False
openid_provider_arn |  |  | True
openid_provider_url |  |  | True
private_key |  |  | False
s3_log_bucket_name |  |  | True
zone_id |  |  | True

## Outputs

Name | Description
--- | ---
load_balancer_arn | Ingress load balancer ARN.
load_balancer_raw_dns | Ingress load balancer DNS name.

