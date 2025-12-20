# Features Overview

pltf focuses on faster, safer Terraform generation with predictable wiring. Deep dives live under `docs/features/*`.

## Map of features
- [Profiles & Defaults](features/profiles.md): org/user defaults (`modules_root`, `default_env`, telemetry).
- [Validation & Lint](features/validation.md): structural checks before render/apply.
- [Backends](features/backends.md): `s3|gcs|azurerm` state backends, independent of target cloud.
- [Custom Modules](features/custom-modules.md): bring-your-own `module.yaml` catalog.
- [Placeholders & Wiring](features/placeholders.md): `${env_name}`, `${layer_name}`, `${module.<id>.<output>}`, `${parent.<output>}`, `${var.<name>}`.
- [Secrets](features/secrets.md): keep secrets out of specs; render as TF vars, not locals.
- [Variables](features/variables.md): env/service vars and CLI `--var` overrides.
- [Telemetry](features/telemetry.md): opt-in/opt-out behavior.

## Terraform generation and execution
- Render-only: `pltf generate -f <spec> --env <name> -o <dir>` (no cloud creds required).
- Generate + run Terraform: `pltf terraform plan|apply|destroy|output|graph` regenerates code every time before invoking Terraform.
- Outputs land under `.pltf/<env>/<layer>/...` with providers, backends, modules, and outputs files.

Examples:
```bash
pltf generate -f example/env.yaml --env prod -o .pltf/example/env/prod
pltf terraform plan -f example/service.yaml --env prod
```

## Variables, placeholders, links
- Reuse specs with `${env_name}`, `${layer_name}`, `${parent.<output>}`, `${module.<id>.<output>}`, `${var.<name>}`.
- Links let modules consume other module outputs without hand-wiring Terraform.
- CLI overrides: `--var key=value` augment or replace spec variables.

## Secrets
- Declare secret keys under `envRef.<name>.secrets`; supply values via environment/CI secrets.
- Rendered as Terraform variables to avoid embedding values in generated code.

## Custom modules
- Run `pltf module init --path <module_dir>` to scaffold `module.yaml`.
- Reference with `source: custom` and point `--modules` or profile `modules_root` to your catalog.

## Validation and lint
- `pltf validate` runs structural checks before generation.
- Terraform `plan/apply` via `pltf terraform ...` always regenerates to reduce drift and catches wiring issues early.
