# Platform Usage

This page walks through the most common CLI flows, CI/CD patterns, and how `pltf` keeps Terraform runs deterministic and cache-friendly.

## Validate / Preview / Generate

```bash
pltf validate -f env.yaml -e prod
pltf preview  -f example/e2e.yaml -e prod
pltf generate -f service.yaml -e prod -m ./modules --out .pltf/example/payments/workspace
```

- `validate` checks spec structure, references, and wiring before any generation. It picks the environment via `--env`, `PLTF_DEFAULT_ENV`, or the profile `default_env`.  
- `preview` surfaces providers, backend, labels, and modules without invoking Terraform.  
- `generate` renders workspace-ready Terraform—file inputs are copied into the output tree, and `--var/-v` overrides merge over spec vars.  
- `--modules/-m` selects a custom module root (local path or git ref); modules tagged `source: custom` resolve only from the custom root, while git/local `source` entries bypass it.  
- `--out/-o` defaults to `.pltf/<environment_name>/workspace` or `.pltf/<environment_name>/<service_name>/workspace`.

## Terraform Commands

```bash
pltf terraform plan    -f service.yaml -e dev     # supports --scan, --cost, --rover, --plan-file
pltf terraform apply   -f env.yaml     -e prod
pltf terraform destroy -f env.yaml     -e prod
pltf terraform output  -f service.yaml -e dev --json
pltf terraform force-unlock -f env.yaml -e prod --lock-id=<id>
```

- Every Terraform helper runs locally in the generated workspace, picks up credentials (`~/.aws`, `~/.docker`, etc.) from the host, and benefits from the plain `.terraform` cache Terraform maintains inside the workspace.
- Plan/apply/destroy run `terraform plan` first and keep `.pltf-plan.tfplan` plus optional tfsec/cost summaries or Rover output for audit trails.  
- Apply/destroy append `-auto-approve` so there’s no interactive approval.  
- `plan` builds images without pushing; `apply` builds and then pushes with the host registry credentials; `destroy` skips image builds.  
### Common Terraform flags

- `--target/-t`, `--parallelism/-p`, `--lock/-l`, `--lock-timeout/-T`, `--no-color/-C`, `--input/-i`, `--refresh/-r`, `--plan-file/-P`, `--detailed-exitcode/-d`, `--json/-j`.

## Module inventory

```bash
pltf module list [-m ./modules] [-o table|json|yaml]
pltf module get aws_eks [-m ./modules] [-o table|json|yaml]
pltf module init --path ./modules/aws_eks [--force]
```

- Lists modules from embedded and custom roots; `source: custom` modules resolve only from your shared catalog (via `--modules` or profile defaults), but you can now point modules directly at git URLs so `modules_root` is optional.
- `module init` emits `module.yaml` from an existing Terraform module directory (`--force` overwrites).

## Profiles & Defaults

- `~/.pltf/profile.yaml` (or `PLTF_PROFILE`) sets `default_env`, `default_out`, and telemetry controls; the old `modules_root` setting is no longer required unless you still rely on a shared catalog.
- `PLTF_DEFAULT_ENV` is also respected when choosing the environment.

## Backends

- `backend.type` must be compatible with the provider (`s3` for AWS, `gcs` for GCP, `azurerm` for Azure).  
- Optional `backend.profile`, `region`, `container`, and `resource_group` support cross-account S3 or Azure storage.

## CI/CD integration

`pltf` works nicely in GitHub Actions, GitLab CI, or any runner that can install Go and Terraform. Jobs share the same env vars/caches as local runs.

### Matrix plan across environments

- Checkout + setup Go/Terraform.  
- Run a matrix plan job once per env/region.  
- Cache `GITHUB_TOKEN`, AWS credentials, and the `.pltf` workspace artifacts for reuse in later steps or PR comments.  
- Example plan step:

```yaml
- name: Terraform plan
  run: |
    pltf terraform plan -f example/e2e.yaml --env ${{ matrix.env }}
```

### Deploy on branch/tag

- Tags push to prod, `main` deploys staging, others map to `development`.  
- Select the env in-step and call `pltf terraform plan` + auto-approved `pltf terraform apply`.  
- Example apply step (staging/prod only):

```yaml
- name: Apply
  if: github.event_name == 'push'
  run: |
    pltf terraform apply -f example/e2e.yaml --env ${{ steps.select.outputs.env }}
```

> Tips: swap AWS env vars with GCP/Azure equivalents and run `pltf terraform plan` on service specs when verifying app deployments. Use `OPENAI_API_KEY` to automatically post plan summaries/AI critiques in PR comments.
