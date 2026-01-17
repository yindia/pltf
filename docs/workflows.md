# Terraform Workflows

`pltf terraform ...` commands share the same generated workspace: specs render first, images build via Dagger, and Terraform runs locally with cached plugins and forced plans (no Dagger layers for Terraform commands themselves).

## Shared motivations

- **Auto generation**: Specs render before each command. The CLI writes the env-specific `.tfvars` file, ensures the backend bucket, and rehydrates the workspace with modules from embedded/custom roots.
- **Terraform runs locally with caching**: `terraform init/plan/apply` run inside the generated workspace so the plain `.terraform` cache inside `.pltf/<env>/workspace` handles provider downloads.
- **Plan first and auto-approved apply/destroy**: `plan`, `apply`, and `destroy` all run `terraform plan` first; apply/destroy continue automatically with `-auto-approve` so automation never stalls.
- **Image build before Terraform**: Dagger builds (and for `apply`, pushes) the Docker images declared in the spec before invoking Terraform so builds benefit from BuildKit caches. Image builds respect the optional spec-level `platforms` list (`["linux/amd64","linux/arm64"]`) and default to the host OS/ARCH when omitted (`destroy` skips image builds entirely).

## Runtime steps

1. **Render the workspace** (shared code ensures `.tfvars`, modules, and backend config are ready).
2. **Build + push images** via Dagger: plan/apply build once (apply pushes with host registry credentials); destroy skips the image build step.
3. **Run `terraform init` locally**: Terraform config keeps its downloads inside the workspace and reuses credentials mounted from your host (AWS, Docker, etc.).
4. **Select or create workspace** matching the env name.
5. **Run `terraform plan` locally**; the planner writes `.pltf-plan.tfplan` and `.pltf-plan.json`, and if the plan succeeds, `apply`/`destroy` continue automatically.

## Offline-friendly hooks

- Write artifacts (`.pltf-plan.tfplan`, plan JSON) to the workspace so CI can ship them elsewhere.
- Enable tfsec scans (`--scan`), cost breakdowns (`--cost`), and Rover output (`--rover`) from the same run.
- `pltf terraform graph` can reuse generated Terraform or take an existing plan with `--plan-file`.

For full CLI reference see [CLI Usage](usage.md).
