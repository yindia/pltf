# Stack

Stacks are reusable templates that define a module bundle you can reference from Environment
or Service specs. They let you share common infrastructure patterns (e.g., EKS + nodegroups +
observability) without duplicating module lists.

## When to use
- You want a standard cluster baseline across many environments.
- You want repeatable add-ons (observability, security, ingress) with minimal config.
- You want a single place to update shared module wiring.

## Minimal Stack spec
```yaml
apiVersion: platform.io/v1
kind: Stack
metadata:
  name: k8s-cluster
variables:
  cluster_name: "pltf-app-${env_name}"
  enable_metrics: "true"
modules:
  - id: base
    type: aws_base
  - id: eks
    type: aws_eks
    inputs:
      cluster_name: var.cluster_name
      enable_metrics: var.enable_metrics
```

## Referencing stacks
In Environment or Service specs, set `metadata.stacks`:
```yaml
metadata:
  stacks:
    - ./stacks/k8s-cluster.yaml
```

Git refs are supported:
```yaml
metadata:
  stacks:
    - https://gitlab.com/org/repo.git//stacks/k8s-cluster.yaml?ref=main
```

## Merge behavior
- Stack modules are added first, then env/service modules are appended.
- Stack module IDs cannot be overridden by env/service modules.
- Stack variables are auto-applied at runtime.
- Env/service `variables` can be used for custom logic only; names must not collide with stack variables.

## Auto-wiring
After merge, module inputs are auto-wired by matching output names across **all**
modules (stack + env/service). Duplicate output names across the merged set are errors.
