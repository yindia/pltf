# Spec Guide

pltf reads YAML specs with `kind: Environment`, `kind: Service`, or `kind: Stack`. The CLI validates structure, merges stacks, wires modules/outputs, and renders provider/backends/config before calling the host `terraform` binary (no embedded Terraform layers).

## Stack spec (kind: Stack)
Minimal shape:
```yaml
apiVersion: platform.io/v1
kind: Stack
metadata:
  name: example-eks-stack
variables:
  cluster_name: "pltf-data-${env_name}"
modules:
  - id: base
    type: aws_base
  - id: eks
    type: aws_eks
    inputs:
      cluster_name: "pltf-app-${env_name}"
      kms_account_key_arn: module.base.kms_account_key_arn
      k8s_version: 1.33
      enable_metrics: false
      max_nodes: 15
```
Notes:
- Stack specs bundle reusable module templates. Reference them from environments or services with `metadata.stacks`.
- Stack variables provide defaults that merge before environment/service values; duplicates are rejected.
- Modules in stacks cannot be redefined by downstream specs with the same `id`.
- Bring your own modules by placing a `module.yaml` next to custom Terraform code. Reference them exactly as you would the built-in modules; pltf treats both identically during generation.

## Environment spec (kind: Environment)
Minimal shape:
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
Notes:
- `environments` describe cloud/account/region entries; variables and secrets are defined at the top level and applied to each env (per-env overrides are not supported).
- `modules` merge with stacks (referenced via `metadata.stacks`), `id` is mandatory, and `type` is required unless `source` is a git/local ref. `inputs`/`links` work the same as Terraform module blocks.
- Backends (S3/GCS/Azure) stay stable after generation; when you change backend config rerun `terraform init -reconfigure`.
- `images` describe Docker builds. Plan builds them using Dagger cache, apply builds + pushes tagged images using host registry credentials, and destroy skips the image step. Omitting `platforms` uses the host OS/ARCH.

## Service spec (kind: Service)
Minimal shape:
```yaml
apiVersion: platform.io/v1
kind: Service
metadata:
  name: payments-api
  ref: ./env.yaml
  envRef:
    dev: {}
    prod: {}
modules:
  - id: api
    type: helm_chart
    inputs:
      chart: ./services/payments/chart
      repo: ./services/payments
      values:
        cluster: module.eks.cluster_name
        replicas: var.replica_count
  - id: db
    type: aws_postgres
variables:
  replica_count: 2
secrets:
  db_password: {}
images:
  - name: payments-api
    context: ./services/payments
    tags:
      - ghcr.io/acme/payments:${env_name}
```
Notes:
- `metadata.ref` points to the environment spec; `metadata.envRef` lists every env the service runs in. Variables/secrets are defined at the top level.
- Services render their own workspace but read environment outputs via remote state.
- Modules in services can reference environment outputs via `${parent.<output>}` templates.

## Image config
```yaml
images:
  - name: app
    context: ./services/app
    dockerfile: Dockerfile
    tags:
      - ghcr.io/acme/app:${env_name}
    buildArgs:
      ENV: ${env_name}
    platforms:
      - linux/amd64
      - linux/arm64
```
Notes:
- `tags` are optional but required when pushing.
- Authenticate via `docker login` before running `pltf terraform apply`; plan builds only, apply also pushes.
- Supply Docker secrets through `PLTF_IMG_SECRET_<NAME>` or `PLTF_IMG_SECRET_FILE_<NAME>` and reference them via `RUN --mount=type=secret,id=<NAME>` in your Dockerfile.

## Variable precedence
1. Stack variables  
2. Environment `variables`  
3. Service `variables`  
4. CLI `--var key=value`

## Secrets vs. locals
- Secrets remain Terraform variables (`var.<name>`).
- Non-secret inputs are also available as locals for wiring convenience; `var.<name>` is always valid.

## Templated references
- `${module.<module>.<output>}` — module output in the current scope.  
- `${var.<name>}` — logical variable; wires to locals/secrets when names match.  
- `${parent.<output>}` — environment output available to services via remote state.  
- `${env_name}` / `${layer_name}` — intrinsic placeholders; for services, `layer_name` equals the service name.
