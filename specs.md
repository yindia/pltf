# Spec Guide

pltf reads YAML specs with `kind: Environment`, `kind: Service`, or `kind: Stack`. The CLI validates structure and wires modules based on names and templated references.

## Stack spec (kind: Stack)
Minimal shape:
```yaml
apiVersion: platform.io/v1
kind: Stack
metadata:
  name: eks-observability
  labels:
    team: platform
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
- Stack specs define reusable module templates.
- Reference stacks from Environment or Service with `metadata.stacks`.
- Modules defined in the spec cannot override stack modules with the same `id`.
- Stack variables provide default inputs and are auto-applied at runtime.
- Environment and Service specs can define top-level `variables` for custom logic only; they must not use stack variable names.

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
  bucket: example-tfstate   # optional; auto-named if omitted
  region: us-east-1
environments:
  dev:
    account: "111111111111"
    region: us-east-1
    secrets:
      db_password: {}
modules:
  - id: base
    type: aws_base
  - id: dns
    type: aws_dns
    inputs:
      domain: var.base_domain
```
Notes:
- `environments` map holds per-env accounts/regions/secrets (no per-env variables).
- `modules` list holds shared modules; `id`/`type` required; `inputs` optional; `links` supported.
- Backend: `backend.type` can be `s3|gcs|azurerm` (independent of provider). `backend.profile` supports cross-account S3; `container/resource_group` for azurerm.
- Modules can set `source: custom` to force resolution from your custom modules root (`--modules` or profile `modules_root`); others fall back to the embedded catalog.

## Service spec (kind: Service)
Minimal shape:
```yaml
apiVersion: platform.io/v1
kind: Service
metadata:
  name: payments-api
  ref: ./env.yaml       # path to Environment spec
  stacks:
    - ./stacks/eks-observability.yaml
  envRef:
    dev: {}
modules:
  - id: app
    type: aws_k8s_service
    inputs:
      cluster_name: var.cluster_name
      public_uri: "/payments"
      image: "ghcr.io/acme/payments:latest"
    links:
      readwrite:
        - db
  - id: db
    type: aws_postgres
```
Notes:
- `metadata.ref` points to the Environment file (relative paths allowed).
- `metadata.envRef` holds per-env secrets (no per-env variables).
- Modules can reference environment outputs via `${parent.<output>}`.
- Git refs are supported for `metadata.ref` and `metadata.stacks` using the format `https://host/org/repo.git//path/to/spec.yaml?ref=main`.

## Variable precedence
1) Stack variables  
2) Environment `variables`  
3) Service `variables`  
4) CLI `--var key=value`

## Secrets vs locals
- Secrets remain as Terraform variables (`var.<name>`).
- Non-secrets become locals; `var.<name>` resolves to locals unless marked secret.

## Templated references
- `${module.<module>.<output>}` — module output in current scope
- `${var.<name>}` — logical variable; wires to locals/secrets when names match
- `${parent.<output>}` — environment output via remote state (service only)
- `${env_name}` / `${layer_name}` — intrinsic placeholders; for services, `layer_name` is the service name
