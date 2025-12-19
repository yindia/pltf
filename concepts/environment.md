# Environment

The common frame that powers your infrastructure.

## What is an Environment?
Environment specs declare which cloud/account/region to configure. From this file, pltf can create the base resources (e.g., Kubernetes clusters, networks, IAM roles, ingress). You’ll usually have one per staging/prod/QA; you can also create per-engineer or per-PR environments for isolated sandboxes.

![Environment](../images/hero.png) <!-- Replace with your environment graphic -->

## Definition (YAML)
```yaml
apiVersion: platform.io/v1
kind: Environment
metadata:
  name: example-aws
  org: myorg
  provider: aws        # cloud provider
  labels:
    team: platform
backend:
  type: s3             # s3|gcs|azurerm
  bucket: my-tf-bucket # optional; auto-named if omitted
  region: us-east-1
environments:
  prod:
    account: "123456789012"
    region: us-east-1
    variables:
      base_domain: prod.example.com
modules:
  - id: base
    type: aws_base
  - id: eks
    type: aws_eks
    inputs:
      cluster_name: var.base_domain
```

Key points:
- `metadata` sets name/org/provider and optional labels (become global tags).
- `backend.type` can be `s3|gcs|azurerm` (independent of provider). `backend.profile` supports cross-account S3.
- `environments` map holds per-env account/region/vars/secrets; pick one via `--env` or profile `default_env`.
- `modules` are shared across services; use embedded catalog or `source: custom` to pull from your own root.

## State Storage
pltf uses your cloud’s native bucket for remote state (S3/GCS/Azurerm). One bucket per environment; state and metadata for the environment and its services live as separate objects. Backends are managed via `backend.*` and can be cross-cloud (e.g., Azure env with S3 backend).

## Next Steps
- Learn about [Modules](../modules.md).
- Explore [Service](service.md) (coming soon) to connect workloads to environments.
