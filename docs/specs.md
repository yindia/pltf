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
providers:
  kubernetes: true
  helm: true
  kustomize: false
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
- `providers` declares required providers (`kubernetes`, `helm`, `kustomize`) without relying on module types.

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
providers:
  kubernetes: true
  helm: true
  kustomize: false
secrets:
  db_password: {}
images:
  - name: platform-tools
    context: ./images/tools
    dockerfile: Dockerfile
    include:
      - "**/*"
    exclude:
      - "**/.git/**"
      - "**/node_modules/**"
    tags:
      - ghcr.io/example/platform-tools:${env_name}
    buildArgs:
      ENV: ${env_name}
backend:
  type: s3
  bucket: example-tfstate   # optional; auto-named if omitted
  region: us-east-1
environments:
  dev:
    account: "111111111111"
    region: us-east-1
modules:
  - id: base
    type: aws_base
  - id: dns
    type: aws_dns
    inputs:
      domain: var.base_domain
```
Notes:
- `environments` map holds per-env accounts/regions (no per-env variables or secrets).
- `modules` list holds shared modules; `id`/`type` required; `inputs` optional; `links` supported.
- Backend: `backend.type` can be `s3|gcs|azurerm` (independent of provider). `backend.profile` supports cross-account S3; `container/resource_group` for azurerm.
- Modules can set `source: custom` to force resolution from your custom modules root (`--modules` or profile `modules_root`); others fall back to the embedded catalog.
- `images` defines Docker build configs; tags are optional.

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
providers:
  kubernetes: true
  helm: true
secrets:
  api_key: {}
images:
  - name: payments-api
    context: ./services/payments
    include:
      - "**/*"
    exclude:
      - "**/.git/**"
      - "**/node_modules/**"
    tags:
      - ghcr.io/acme/payments:${env_name}
    buildArgs:
      SERVICE: payments
      ENV: ${env_name}
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
- `metadata.envRef` selects envs only (no per-env variables or secrets).
- Modules can reference environment outputs via `${parent.<output>}`.
- Git refs are supported for `metadata.ref` and `metadata.stacks` using the format `https://host/org/repo.git//path/to/spec.yaml?ref=main`.
- `images` can also be defined in Service specs for app images.

### Image config
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
- `tags` are optional; when pushing images, at least one tag is required.
- Authenticate to registries outside of pltf (e.g., `docker login`) before running `pltf image build` or `pltf terraform apply`.
- `pltf terraform plan` builds images; `pltf terraform apply` builds + pushes them.
- Use `include`/`exclude` to filter the build context sent to Dagger.
- Use `platforms` to declare the target OS/ARCH combos (`linux/amd64`, `linux/arm64`, etc.); when absent, the host OS/ARCH is used.
- For Dockerfile build secrets, set `PLTF_IMG_SECRET_<NAME>` or `PLTF_IMG_SECRET_FILE_<NAME>` and use `RUN --mount=type=secret,id=<NAME>` in the Dockerfile.

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
