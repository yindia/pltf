# Features Overview

pltf focuses on faster, safer Terraform generation with predictable wiring. Deep dives live under `docs/features/*`.

## Map of features
- [Profiles & Defaults](features/profiles.md): org/user defaults (`default_env`, telemetry).
- [Validation](features/validation.md): structural checks before render/apply.
- [Backends](features/backends.md): `s3|gcs|azurerm` state backends matched to provider defaults.
- [Custom Modules](features/custom-modules.md): bring-your-own `module.yaml` catalog.
- [Placeholders & Wiring](features/placeholders.md): `${env_name}`, `${layer_name}`, `${module.<id>.<output>}`, `${parent.<output>}`, `${var.<name>}`.
- [Secrets](features/secrets.md): keep secrets out of specs; render as TF vars, not locals.
- [Variables](features/variables.md): spec variables and CLI `--var` overrides.
- [Telemetry](features/telemetry.md): opt-in/opt-out behavior.

## Terraform generation and execution
- Render-only: `pltf generate -f <spec> --env <name> -o <dir>` (no cloud creds required).
- Generate + run Terraform: `pltf terraform plan|apply|destroy|output|graph` regenerates code every time before invoking Terraform.
- Outputs land under `.pltf/<environment_name>/workspace` (or `.pltf/<environment_name>/<service_name>/workspace`) with providers, backends, modules, and outputs files.

Examples:
```bash
	pltf generate -f example/e2e.yaml --env prod -o .pltf/example/workspace
	pltf terraform plan -f service.yaml --env prod
```

## Variables, placeholders, links
- Reuse specs with `${env_name}`, `${layer_name}`, `${parent.<output>}`, `${module.<id>.<output>}`, `${var.<name>}`.
- Links let modules consume other module outputs without hand-wiring Terraform.
- CLI overrides: `--var key=value` augment or replace spec variables.

## Secrets
- Declare secret keys under top-level `secrets`; supply values via environment/CI secrets.
- Rendered as Terraform variables to avoid embedding values in generated code.

## Custom modules
- Run `pltf module init --path <module_dir>` to scaffold `module.yaml`.
- Reference with `source: custom` or the new per-module `source` URL syntax; previously `--modules`/`modules_root` were required but custom git sources now make those optional.

## Validation
- `pltf validate` runs structural checks before generation (Environment, Service, Stack).
- `pltf terraform ...` always regenerates to reduce drift and catches wiring issues early.
