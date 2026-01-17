# Terraform Generator

Render Terraform from your specs without applying. Ideal for reviews, CI, or migrating to raw Terraform.

## Overview
`pltf generate` reads Environment or Service specs, auto-detects kind, and writes a workspace-ready Terraform directory (providers, backend, modules, outputs, versions, variables). No cloud credentials are required to render.

## Generate Terraform
Environment:
```bash
pltf generate -f env.yaml -e prod -o .pltf/example/workspace
# produces:
# .pltf/example/workspace/
# ├─ modules/               # copied/embedded module code used by this stack
# ├─ providers.tf           # provider blocks + required versions
# ├─ backend.tf             # state backend (s3|gcs|azurerm)
# ├─ locals.tf              # locals wired to variables for workspaces
# ├─ modules-*.tf           # module instantiations
# ├─ outputs.tf             # outputs
# ├─ variables.tf           # input variables for workspace runs
# ├─ prod.tfvars            # tfvars for the selected environment
# └─ versions.tf            # provider/Terraform constraints
```
All environments (single workspace root + multiple tfvars):
```bash
pltf generate -f env.yaml -o .pltf/example/workspace
```
Service:
```bash
pltf generate -f service.yaml -e prod -o .pltf/example/payments/workspace
```

## Migrate to Terraform
- Run `pltf generate` (or `pltf terraform plan` to generate+init) for each env/service workspace.
- Commit the generated directory to VCS if you want to manage TF directly.
- Backends follow your spec; use `backend.type` (`s3|gcs|azurerm`) to point at your state bucket/container. Workspace mode uses `workspace_key_prefix` (S3/Azure) or `prefix` (GCS).

## Notes
- No provider calls during generation; safe to run without credentials.
- Supports custom modules (`source: custom`) alongside embedded ones; the generated `modules/` directory is self-contained.
- Use `pltf preview` first to sanity check provider/backend/modules before generation.
