# Layer (Service)

An independently managed set of modules.

## What is a Layer?
You can put all modules in an Environment file, but for finer granularity you define layers (Services). A layer provisions a set of modules together as a single unit and links to an Environment.

Layers have:
- a unique name
- the environment(s) they run in (via `metadata.ref` / `envRef`)
- a list of modules (with optional links and inputs)

![Layer](../images/hero.png) <!-- Replace with layer graphic -->

## When to use layers?
- Break down a large environment into separately maintained stacks.
- Share module definitions across multiple environments without duplicating YAML.
- Isolate per-service concerns (e.g., app plus its database) while reusing environment foundations.

## Definition (YAML)
Example: a service (layer) for a Kubernetes workload with a database.

```yaml
apiVersion: platform.io/v1
kind: Service
metadata:
  name: payments-api
  ref: ../env.yaml           # link to Environment
  envRef:
    prod: {}                 # environment keys supported
modules:
  - id: app
    type: aws_k8s_service
    inputs:
      public_uri: "/payments"
      image: "ghcr.io/acme/payments:latest"
    links:
      readwrite:
        - db
  - id: db
    type: aws_postgres
    inputs:
      instance_class: db.t3.medium
```

Notes:
- Service name maps to `${layer_name}` placeholder; `${env_name}` is the environment key.
- Modules can link to each other (`links`) to consume outputs without manual wiring (e.g., `${module.db.db_host}`).
- Per-environment overrides live under `metadata.envRef`.

## Next Steps
- See [Environment](environment.md) for foundations.
- Explore module details in [References](../references/aws.md).
