# Terraform Workflows

pltf terraform commands target the generated workspace, run the host `terraform` binary, and only use Dagger for building/pushing Docker images declared in the specs. This keeps toolchains transparent while still giving you predictable, k8s-native deployments.

## Shared motivations

- **Auto generation**: Specs always render before Terraform runs. The CLI materializes `.tf`, `.tfvars`, backend config, and provider wiring for every env.
- **Terraform runs locally with caching**: `terraform init/plan/apply` execute inside `.pltf/<environment_name>/workspace` (env) or `.pltf/<environment_name>/<service_name>/workspace` (service), so Terraform’s native `.terraform` cache downloads providers there without extra wrappers.
- **Pre-plan for summaries**: `plan` always runs first (streaming tfsec, cost, Rover logs if enabled); `apply`/`destroy` then execute directly with `-auto-approve`.
- **Image builds via Dagger**: Plan/apply trigger Dagger builds of the declared images using shared caches; apply pushes the resulting tags with host registry creds, while destroy skips image builds altogether. `platforms` drive multi-arch builds and default to the host OS/ARCH.

## Runtime steps

1. **Render the workspace** (modules, `.tfvars`, backend, and provider config).
2. **Build images via Dagger** (plan builds locally, apply builds + pushes, destroy skips this step).
3. **Run `terraform init` locally**; provider plugins are cached inside the workspace and reuse host credentials (AWS, GCP, Azure, Docker).
4. **Run `terraform plan` (and apply/destroy)** inside the workspace; the plan writes `.pltf-plan.tfplan`/`.json`, and `apply`/`destroy` proceed after the pre-plan step.

## Observability hooks

- Write artifacts (`.pltf-plan.tfplan`, `.pltf-plan.json`) to the workspace so downstream CI or comment bots re-use the data.
- Enable tfsec (`--scan`), Infracost (`--cost`), and Rover (`--rover`) during plan runs; the CLI streams their output alongside Terraform logs.
- `pltf terraform graph` can reuse the generated workspace or an existing plan via `--plan-file`.

For full CLI reference see [CLI Usage](usage.md).
