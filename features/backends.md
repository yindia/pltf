# Backends

Choose where Terraform state lives, aligned with the provider.

## What it does
- Supports `backend.type` = `s3|gcs|azurerm` with provider compatibility (AWS → S3, GCP → GCS, Azure → Azurerm).
- Allows `backend.profile` for cross-account S3, `region` override, and container/resource_group for azurerm.
- Ensures the backend bucket/container exists before running Terraform.

## Example
```yaml
backend:
  type: s3
  bucket: platform-tfstate
  region: us-east-1
  profile: ops-account
```

## Notes
- Backends are rendered into `backend.tf`/`terraform.tfvars` alongside providers.
- If `backend.type` is omitted, pltf picks the default backend for the provider.
