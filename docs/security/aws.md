# AWS Architecture

Architecture overview for AWS deployments of pltf.

## Description
- Single-region deployments with networking across three AZs by default (public + private subnets). Public subnets are used for public load balancers; EC2/Databases stay in private subnets (NAT for egress).
- EKS cluster spans private subnets with managed node groups. Cluster version is configurable (`aws_eks.k8s_version`) and patched by AWS. Public endpoint by default (VPN/private endpoints can be added later). Secrets are encrypted via KMS.
- Datastores: modules for Postgres (Aurora), Redis (ElastiCache), DocumentDB. Multi-AZ supported; 5-day backup retention for Postgres/DocumentDB. Credentials are generated and passed securely to services.
- S3: buckets are private by default, encrypted at rest (AES-256); can be made public via inputs.
- SQS: queues created with dedicated KMS keys for encryption at rest.
- SNS: topics created with dedicated KMS keys for encryption at rest.
- IAM: IAM role/user modules with `links` auto-generate least-privilege policies (S3, SQS, SNS, SES, etc.) and IRSA trusts for Kubernetes services.
- DNS/SSL: Route53 hosted zone and ACM certificates; validation via Route53; records created to point to the load balancer.

## Security Overview
- End-to-end TLS when using ingress + service mesh (Linkerd optional) and delegated domains.
- Databases and EC2s in private subnets; only NAT egress.
- Databases (Postgres/Redis/DocumentDB) encrypted at rest with KMS; connections use SSL.
- S3 buckets encrypted at rest (AES-256); private by default.
- SQS/SNS encrypted at rest with per-resource KMS keys.
- Networking gated by security groups (EKS-managed + module-specific SGs) with minimal port exposure.
- EKS nodes created with scoped IAM policies; cluster storage (Secrets) encrypted via KMS.
- K8s service accounts mapped to IAM roles via OIDC (IRSA); no long-lived credentials.
- No long-lived IAM credentials are created by default; ECR images remain private.
- 5-day backup retention for Postgres/DocumentDB.
- Public EKS endpoint by default for simplicity; private/VPN options can be layered later.
