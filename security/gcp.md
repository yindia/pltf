# GCP Architecture

Security highlights for GCP environments managed by pltf.

## Identity and access
- Use GCP service accounts for workloads and modules.
- For GKE workloads, prefer Workload Identity with `gcp_k8s_service` or `gcp_service_account`.
- Restrict IAM roles to the minimum required and use `links` for GCS access grants.

## Network controls
- GKE uses VPC/subnet wiring from `gcp_base`.
- Keep control plane access limited to trusted networks.

## State and secrets
- Use the `gcs` backend for Terraform state with bucket-level IAM.
- Store credentials in CI secret stores, not in specs.

## Logging
- Enable GKE and VPC flow logs as required for audit trails.
