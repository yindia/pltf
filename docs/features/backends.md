# Backends

Choose where Terraform state lives, independent of the target cloud.

## What it does
- Supports `backend.type` = `s3|gcs|azurerm` for any provider.
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
- You can point all clouds to a single backend (e.g., S3) if desired.
