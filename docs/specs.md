# Spec Guide

pltf reads YAML specs with `kind: Environment` or `kind: Service`. The CLI validates structure and wires modules based on names and templated references.

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
backend:
  type: s3
  bucket: example-tfstate   # optional; auto-named if omitted
  region: us-east-1
environments:
  dev:
    account: "111111111111"
    region: us-east-1
    variables:
      base_domain: dev.example.com
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
- `environments` map holds per-env accounts/regions/vars/secrets.
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
  envRef:
    dev:
      variables:
        cluster_name: dev-cluster
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
- `envRef` holds per-env variables/secrets merged after environment variables.
- Modules can reference environment outputs via `${parent.<output>}`.

## Variable precedence
1) Environment variables  
2) Service envRef variables (service only)  
3) CLI `--var key=value`

## Secrets vs locals
- Secrets remain as Terraform variables (`var.<name>`).
- Non-secrets become locals; `var.<name>` resolves to locals unless marked secret.

## Templated references
- `${module.<module>.<output>}` — module output in current scope
- `${var.<name>}` — logical variable; wires to locals/secrets when names match
- `${parent.<output>}` — environment output via remote state (service only)
- `${env_name}` / `${layer_name}` — intrinsic placeholders; for services, `layer_name` is the service name
