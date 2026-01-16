# Terraform Workflows

Every `pltf terraform ...` command uses the same Dagger session abstraction so Terraform runs and image builds execute inside a consistent container environment.

## Shared motivations

- **Auto generation**: Specs render before each command. The CLI writes the env-specific `.tfvars` file, ensures the backend bucket, and rehydrates the workspace with modules from embedded/custom roots.
- **Plan first for all actions**: `plan`, `apply`, and `destroy` all run `terraform plan` first and collect the `.pltf-plan.tfplan` alongside optional tfsec/rover/infracost summaries.
- **Apply/destroy are auto-approved**: Both commands append `-auto-approve`, and no interactive approval prompt is shown inside Dagger.
- **Image build before Terraform**: The same Dagger session builds (and for `apply`, pushes) the Docker images declared in the spec before calling Terraform. Image builds respect the optional spec-level `platforms` list (e.g., `["linux/amd64","linux/arm64"]`) and default to the host OS/ARCH when omitted.

## Runtime steps

1. **Start Dagger session** (logs sent to stderr unless `PLTF_DAGGER_*` opts are set).
2. **Mount cached assets**: workdir, AWS/GCP credentials, `.docker`, and the shared `pltf-terraform-plugin-cache` volume are mounted into `/work`.
3. **Run `terraform init`**: Backends are configured (S3 etc) and plugins download into the cached volume.
4. **Select or create workspace** matching the env name.
5. **Build + push images**: plan/apply/destroy build, apply pushes via host credentials, plan/destroy only build locally.
6. **Run `terraform plan`**: The planner uses `TF_PLUGIN_CACHE_DIR` and writes its plan; if the plan succeeds, `apply`/`destroy` follow immediately.

## Offline-friendly hooks

- Write artifacts (`.pltf-plan.tfplan`, plan JSON) to the workspace so CI can ship them elsewhere.
- Enable tfsec scans (`--scan`), cost breakdowns (`--cost`), and Rover output (`--rover`) from the same run.
- `pltf terraform graph` can reuse generated Terraform or take an existing plan with `--plan-file`.

For full CLI reference see [CLI Usage](usage.md).
