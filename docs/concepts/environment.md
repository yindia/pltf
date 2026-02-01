# Environment

The shared foundation for your stacks: cloud, account/project, region, and base modules (VPC, DNS, EKS/GKE/AKS, IAM).

<div class="mermaid">
flowchart TB
    svc[(service.yaml)]

    subgraph PROD[Production Env]
        prod_service[Service A]
    end

    subgraph STAGE[Staging Env]
        stage_service[Service A]
    end

    env[(env.yaml)]

    svc --> prod_service
    svc --> stage_service

    prod_service --> env
    stage_service --> env
</div>

## Definition (example)
Based on `example/e2e.yaml`:

```yaml
apiVersion: platform.io/v1
kind: Environment
gitProvider: github

metadata:
  name: example-aws
  org: pltf
  provider: aws
  labels:
    team: platform
    cost_center: shared
  stacks:
    - example-eks-stack
# images:
#   - name: platform-tools
#     context: .
#     dockerfile: Dockerfile
#     platforms:
#       - linux/amd64
#       - linux/arm64
#     tags:
#       - ghcr.io/example/${layer_name}:${env_name}
#     buildArgs:
#       ENV: ${env_name}
environments:
  dev:
    account: "556169302489"
    region: ap-northeast-1
  stage:
    account: "556169302489"
    region: ap-northeast-1
  prod:
    account: "556169302489"
    region: ap-northeast-1
variables:
  replica_counts: '{"dev":1,"prod":3}'
  environment_settings: '{"region":"us-west-2","zones":["us-west-2a","us-west-2b"]}'
modules:
  - id: nodegroup1
    source: ../modules/aws_nodegroup
    inputs:
      max_nodes: 15
      node_disk_size: 20
  - id: postgres
    source: https://github.com/yindia/pltf.git//modules/aws_postgres?ref=main
    inputs:
      database_name: "${layer_name}-${env_name}"
  - id: s3
    type: aws_s3
    inputs:
      bucket_name: "pltf-app-${env_name}"
    links:
      readWrite:
        - adminpltfrole
        - userpltfrole
  - id: topic
    type: aws_sns
    inputs:
      sqs_subscribers:
        - "${module.notifcationsQueue.queue_arn}"
    links:
      read: adminpltfrole
  - id: notifcationsQueue
    type: aws_sqs
    inputs:
      fifo: false
    links:
      readWrite: adminpltfrole
  - id: schedulesQueue
    type: aws_sqs
    inputs:
      fifo: false
    links:
      readWrite: adminpltfrole
  - id: adminpltfrole
    type: aws_iam_role
    inputs:
      extra_iam_policies:
        - "arn:aws:iam::aws:policy/CloudWatchEventsFullAccess"
      allowed_k8s_services:
        - namespace: "*"
          service_name: "*"
  - id: userpltfrole
    type: aws_iam_role
    inputs:
      extra_iam_policies:
        - "arn:aws:iam::aws:policy/CloudWatchEventsFullAccess"
      allowed_k8s_services:
        - namespace: "*"
          service_name: "*"
```

## Key points

- **Metadata**: name/org/provider; labels become tags.
- **environments**: per-env account/region; variables/secrets are top-level and applied to each env; select with `--env prod`.
- **modules**: shared building blocks. Use the embedded catalog or `source: custom` with your module root.
- **Backends**: choose a provider-compatible backend (`s3|gcs|azurerm`); use profiles for cross-account S3 (set in profiles or flags).

## Outputs and linking
Environment module outputs are addressable by Services via links or `${module.<id>.<output>}`. This keeps Services thin while reusing the foundation.

## Next steps
- See [Layer/Service](layer.md) to attach workloads.
- Browse module APIs in [References](../references/aws.md) and the per-module pages.
