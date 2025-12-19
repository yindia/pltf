# Features Overview

Each major capability has its own page:

- [Profiles & Defaults](features/profiles.md)
- [Validation & Lint](features/validation.md)
- [Backends](features/backends.md)
- [Custom Modules](features/custom-modules.md)
- [Placeholders & Wiring](features/placeholders.md)
- [Secrets](features/secrets.md)
- [Variables](features/variables.md)
- [Telemetry](features/telemetry.md)

## Secrets
- **What:** Manage app secrets without embedding them in specs/code. Secrets are stored as Kubernetes secrets and injected as env vars.
- **Why:** Avoid leaking credentials; keep rotation simple.
- **Usage:** Define secret keys in your spec under `secrets` and supply values via environment variables or CLI `--var`. Secrets are treated as TF variables, not locals.
- **Notes:** Services restart to pick up changes unless `--no-restart` is used. Bulk updates can consume `.env`-style inputs; values should come from env/CI secret stores, not hardcoded files.

## Terraform Generator
- **What:** Render Terraform from env/service specs without applying; handy for review, migration, or running TF directly.
- **Why:** Keep portability—inspect/modify TF, hand to CI, or migrate away without lock-in.
- **Commands:** `pltf generate` for TF only; `pltf terraform plan|apply|destroy|output|force-unlock` to generate + run.
- **Example (env):**
  ```bash
  pltf generate -f env.yaml -e prod -o .pltf/env/prod
  # outputs providers.tf, backend.tf, modules/<...>, outputs.tf, versions.tf
  ```
- **Example (service):**
  ```bash
  pltf generate -f service.yaml -e prod -o .pltf/service/payments/prod
  ```
- **Notes:** Does not require cloud credentials to render. Backends are written per spec (`s3|gcs|azurerm`). Generated modules directory is self-contained for review or VCS.

## Variables
- **What:** Minimal templating to reuse specs across envs/services.
- **Types:** CLI `--var`, env-level `variables`, and placeholders.
- **Placeholders:** `${env_name}`, `${layer_name}`, `${module.<id>.<output>}`, `${parent.<output>}`, `${var.<name>}`.
- **Spec inputs:** Declare `variables` in env specs or use `--var key=value` at runtime; service specs inherit envRef variables and can override via CLI.
- **Example (env):**
  ```yaml
  variables:
    min_nodes: "2"
    max_nodes: "5"
  modules:
    - type: aws_eks
      min_nodes: "${var.min_nodes}"
      max_nodes: "${var.max_nodes}"
  ```
  ```bash
  pltf terraform apply -f env.yaml -e prod --var min_nodes=3 --var max_nodes=6
  ```
- **Parent outputs:** In services, `${parent.<output>}` references environment outputs (e.g., `${parent.domain}`).
