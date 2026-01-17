# Spec Guide

pltf reads YAML specs with `kind: Environment`, `kind: Service`, or `kind: Stack`. The CLI validates structure, merges stacks, wires modules/outputs, and renders provider/backends/config before calling the host `terraform` binary (no embedded Terraform layers).

## Stack spec (kind: Stack)
Minimal shape:
```yaml
apiVersion: platform.io/v1
kind: Stack
metadata:
  name: eks-observability
  labels:
    team: platform
providers:
  kubernetes: true
  helm: true
variables:
  cluster_name: ""
  base_domain: ""
modules:
  - id: eks
    type: aws_eks
    inputs:
      cluster_name: var.cluster_name
  - id: sec
    type: aws_security_baseline
  - id: obs
    type: aws_observability
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
metadata:
  name: example-aws
  org: example-org
  provider: aws
  labels:
    team: platform
  stacks:
    - ./stacks/eks-observability.yaml
backend:
  type: s3
  bucket: example-tfstate
  region: us-east-1
modules:
  - id: base
    type: aws_base
  - id: dns
    type: aws_dns
    inputs:
      domain: var.base_domain
variables:
  cluster_name: example-cluster
  base_domain: example.com
secrets:
  db_password: {}
environments:
  dev:
    account: "111111111111"
    region: us-east-1
    variables:
      log_bucket: dev-logs
  prod:
    account: "222222222222"
    region: us-west-2
    secrets:
      db_password: {}
images:
  - name: platform-tools
    context: ./images/tools
    dockerfile: Dockerfile
    tags:
      - ghcr.io/example/platform-tools:${env_name}
    buildArgs:
      ENV: ${env_name}
    platforms:
      - linux/amd64
      - linux/arm64
```
Notes:
- `environments` describe cloud/account/region entries and can override `variables`/`secrets` per entry.
- `modules` merge with stacks (referenced via `metadata.stacks`), `id` and `type` are mandatory, and `inputs`/`links` work the same as Terraform module blocks.
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
    prod:
      variables:
        replica_count: 3
      secrets:
        api_key: {}
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
- `metadata.ref` points to the environment spec; `metadata.envRef` lists every env the service runs in, optionally overriding `variables`, `secrets`, or even `images` per env.
- Services reuse the generated workspace of their referenced environment, so Terraform runs share state and graph data.
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
- Non-secret inputs become locals; `var.<name>` resolves to locals unless marked secret explicitly.

## Templated references
- `${module.<module>.<output>}` — module output in the current scope.  
- `${var.<name>}` — logical variable; wires to locals/secrets when names match.  
- `${parent.<output>}` — environment output available to services via remote state.  
- `${env_name}` / `${layer_name}` — intrinsic placeholders; for services, `layer_name` equals the service name.
