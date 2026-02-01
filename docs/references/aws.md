# AWS Reference

AWS is fully supported for environments, services, and the embedded module catalog. This page summarizes how the AWS provider, backends, and module wiring work in pltf.

## Provider and Backends
- **Provider:** Automatically injected; version comes from the central versions file. Region is taken from the selected environment entry.
- **Backends:** AWS uses `backend.type: s3` (default if omitted). Use `backend.profile` for cross-account state and `backend.region` to override the bucket region.
- **Default tags:** Labels in your env/service specs become global tags on the AWS provider.

## Example (Environment + Service)
Environment:
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

Service:
```yaml
apiVersion: platform.io/v1
kind: Service
metadata:
  name: payments-api
  ref: ./env.yaml
  envRef:
    prod: {}
variables:
  image: ghcr.io/acme/payments:latest
modules:
  - id: app
    type: helm_chart
    inputs:
      chart: ./charts/payments
      values:
        image: var.image
  - id: app-bucket
    type: aws_s3
    inputs:
      bucket_name: "payments-${env_name}"
  - id: app-queue
    type: aws_sqs
```

## Modules and Fields
- **id:** required and unique within the stack.
- **type:** selects the module implementation; required unless `source` is a git/local path with `module.yaml`.
- **source:** optional; `custom` forces lookup in your custom modules root, while git/paths load metadata directly.
- **inputs:** key/value config for module variables.
- **links:** access bindings that let modules consume other module outputs (IAM policies/IRSA).

## Linking
Linking lets a module consume outputs of another:
```yaml
links:
  readWrite:
    - app-bucket
  consume:
    - app-queue
```
When links are present, pltf automatically renders IAM policies and (for Kubernetes) IRSA trusts. Supported AWS link targets include S3, SQS, SNS, SES, DynamoDB, RDS, and more via module metadata.

## Template placeholders
- `${env_name}` and `${layer_name}` become the resolved environment/service names.
- `${module.<module_name>.<output_name>}` references another module’s output.
- `${parent.<output_name>}` references outputs from the parent environment when authoring a service.
- `${var.<name>}` references variables defined in the spec or via `--var`.

## Useful commands
- `pltf module list -o table` — see available AWS modules.
- `pltf module get aws_eks` — inspect inputs/outputs.
- `pltf generate -f env.yaml -e prod` — render Terraform for AWS.
- `pltf terraform plan/apply ...` — generate + execute Terraform (plan/apply/destroy/output/force-unlock).

See the module-specific pages under “Modules (AWS)” for detailed inputs, outputs, and examples.
