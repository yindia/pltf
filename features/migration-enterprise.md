# Enterprise Migration Guide

This guide explains how to migrate a large Terraform monorepo into PLTF and
incrementally split a monolith into smaller components.

## Goals
- Start using the PLTF CLI without rewriting everything.
- Keep existing Terraform code intact at first.
- Move toward smaller, reusable modules over time.

## Phase 1: Wrap the existing root
1) Identify the current Terraform root(s) you run in CI/CD.
2) Move each root into a module folder under a custom modules root.
3) Generate `module.yaml` for each:
```
pltf module init --path ./modules/monolith --force
```
4) Create a minimal Environment spec:
```yaml
apiVersion: platform.io/v1
kind: Environment
metadata:
  name: enterprise
  org: corp
  provider: aws
environments:
  prod:
    account: "111111111111"
    region: us-east-1
modules:
  - id: monolith
    type: monolith
    source: custom
```
5) Generate and plan via PLTF:
```
pltf terraform plan -f env.yaml -e prod --modules ./modules
```

This keeps your existing Terraform intact while switching the CLI and validation layer.

## Phase 2: Introduce stacks and shared baselines
Extract shared infrastructure (networking, cluster, security) into a Stack:
```yaml
kind: Stack
metadata:
  name: baseline
modules:
  - id: base
    type: aws_base
  - id: eks
    type: aws_eks
```
Then reference it from Environment or Service specs:
```yaml
metadata:
  stacks:
    - baseline
```

## Phase 3: Split the monolith
Pick natural boundaries:
- networking
- cluster/platform
- data stores
- application services

Process:
1) Copy the relevant Terraform code into a new module folder.
2) Define clear inputs/outputs in the module.
3) Run `pltf module init` to create metadata.
4) Add the module to the spec and connect outputs by name.

Small, safe increments:
- Move one subsystem at a time.
- Use stack variables to keep defaults stable.
- Let PLTF auto-wire module outputs by name.

## Phase 4: Organize by layer
Use Environment specs for shared infra and Service specs for app‑specific modules:
- Environment: base networking, cluster, shared data stores.
- Service: app service modules, app databases, app-specific DNS/ingress.

This reduces blast radius and improves deploy cadence.

## Directory layout example
```
repo/
├─ env.yaml
├─ service.yaml
├─ modules/
│  ├─ monolith/
│  ├─ network/
│  ├─ cluster/
│  └─ app/
└─ stacks/
   └─ baseline.yaml
```

## Tips
- Keep module outputs unique across the merged set (stack + env/service).
- Prefer stable, backward‑compatible inputs while splitting.
- Use `pltf preview` to verify module ordering and stacks.
- Use `pltf module list|get` to inspect metadata changes over time.
