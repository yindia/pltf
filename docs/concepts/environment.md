# Environment

The shared foundation for your stacks: cloud, account/project, region, and base modules (VPC, DNS, EKS/GKE/AKS, IAM).

![Environment](../images/hero.png)

## Definition (example)
Based on `example/env.yaml`:

```yaml
apiVersion: platform.io/v1
kind: Environment

metadata:
  name: example-aws
  org: pltf
  provider: aws
  labels:
    team: platform
    cost_center: shared
environments:
  prod:
    account: "556169302489"
    region: ap-northeast-1
    variables:
      base_domain: prod.pltf.internal
      cluster_name: pltf-data
modules:
  - id: base
    type: aws_base
  - id: dns
    type: aws_dns
    inputs:
      domain: ${{var.base_domain}}
      delegated: false
  - id: eks
    type: aws_eks
    inputs:
      cluster_name: "pltf-app-${layer_name}-${env_name}"
      k8s_version: 1.33
      enable_metrics: false
      max_nodes: 15
  - id: nodegroup1
    type: aws_nodegroup
    inputs:
      max_nodes: 15
      node_disk_size: 20
```

## Key points

- **Metadata**: name/org/provider; labels become tags.
- **environments**: per-env account/region/vars/secrets; select with `--env prod`.
- **modules**: shared building blocks. Use the embedded catalog or `source: custom` with your module root.
- **Backends**: choose `s3|gcs|azurerm` independently of provider; use profiles for cross-account S3 (set in profiles or flags).

## Outputs and linking
Environment module outputs are addressable by Services via links or `${module.<id>.<output>}`. This keeps Services thin while reusing the foundation.

## Next steps
- See [Layer/Service](layer.md) to attach workloads.
- Browse module APIs in [References](../references/aws.md) and the per-module pages.
