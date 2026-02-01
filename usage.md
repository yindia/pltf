# CLI Usage

`pltf` detects the spec kind (Environment, Service, or Stack) and merges stack/env/service layers before leaning on Terraform. Most spec-based commands support the shared flags `--file/-f`, `--env/-e`, `--modules/-m`, `--out/-o`, and `--var/-v key=value` unless noted (image builds are the main exception).

## Covered commands

| Command                             | What it does                                                           |
|------------------------------------|------------------------------------------------------------------------|
| `pltf validate`                     | Structural validation for specs and wiring                              |
| `pltf preview`                      | Summarizes providers, backend, variables, and module wiring             |
| `pltf generate`                     | Renders Terraform workspaces without running Terraform                  |
| `pltf terraform plan`             | Builds spec images through Dagger (no push), streams tfsec/cost/rover logs, and runs Terraform locally |
| `pltf terraform apply`            | Builds + pushes spec images via Dagger and runs `terraform apply -auto-approve` locally |
| `pltf terraform destroy`          | Runs `terraform destroy -auto-approve` locally (skips image builds) using the generated workspace |
| `pltf terraform graph`             | Terraform/spec DOT graphs (supports `--mode`)                          |
| `pltf terraform output`            | Shows outputs (JSON support)                                            |
| `pltf terraform force-unlock`      | Removes stale Terraform locks (`--lock-id` required)                   |
| `pltf module list|get|init`         | Inspect modules from embedded/custom roots                              |
| `pltf image build`                 | Build/push Docker images defined in the spec via Dagger                 |

## Shared flags

- `--file/-f`: Path to the spec (defaults to `env.yaml`).
- `--env/-e`: Environment key (falls back to `PLTF_DEFAULT_ENV` or profile defaults).
- `--modules/-m`: Custom modules root (local path or git ref). Modules with `source: custom` are resolved here before embedded ones.
- `--out/-o`: Output directory (default `.pltf/<environment_name>/workspace` or `.pltf/<environment_name>/<service_name>/workspace`).
- `--var/-v key=value`: CLI vars merge over spec variables and support strings, numbers, JSON, and lists.
- `--target`, `--parallelism`, `--lock`, `--lock-timeout`, `--no-color`, `--input`, `--refresh`, `--plan-file`, `--detailed-exitcode`, `--json` (passed to Terraform where applicable).
- (Advanced) If you run `pltf image ...` with a remote Dagger engine, set `DAGGER_HOST` and use `PLTF_DAGGER_LOG`, `PLTF_DEBUG`, or `PLTF_VERBOSE` to control verbosity.

## Terraform helpers

### `pltf terraform plan`

- Builds the Docker images declared in the spec via Dagger (no push), renders the workspace, and runs `terraform plan` locally inside `.pltf/<environment_name>/<service_name>/workspace` (service) or `.pltf/<environment_name>/workspace` (environment).
- Supports `--scan`, `--cost`, `--rover`, `--detailed-exitcode`, `--plan-file`, and the shared flags above while streaming tfsec/cost/Rover logs in real time.
- Produces `.pltf-plan.tfplan` and optional plan JSON; `apply`/`destroy` run their own pre-plan step for summaries and then execute the apply/destroy directly.
- Example: `pltf terraform plan -f service.yaml -e dev --detailed-exitcode --plan-file=/tmp/plan.tfplan`

### `pltf terraform apply`

- Builds and pushes images (host registry auth managed on the host) via Dagger and runs `terraform apply -auto-approve` locally after a pre-plan for summaries.
- Example: `pltf terraform apply -f env.yaml -e prod`

### `pltf terraform destroy`

- Skips image builds, runs a pre-plan for summaries, and executes `terraform destroy -auto-approve` locally so automation never pauses.
- Example: `pltf terraform destroy -f env.yaml -e prod`

### `pltf terraform output`

- Prints outputs and optionally renders JSON (`--json`).
- Example: `pltf terraform output -f service.yaml -e dev --json`

### `pltf terraform graph`

- Emits DOT graphs. Default runs `terraform graph`, `--mode spec` walks YAML dependencies, and `--plan-file/-P` reuses saved plans.
- Example: `pltf terraform graph -f env.yaml -e dev | dot -Tpng > graph.png`

### `pltf terraform force-unlock`

- Clears Terraform locks with `--lock-id`.
- Example: `pltf terraform force-unlock -f env.yaml -e prod --lock-id=12345`

## Validation & generation

- `pltf validate` checks spec structure, refs, and readiness for rendering.
- `pltf generate` renders Terraform-only workspaces; file inputs are copied/rewired into the output tree.
- Flags: `--modules/-m`, `--out/-o`, `--var/-v`.
- Examples:
  ```bash
  pltf validate -f env.yaml -e prod
  pltf generate -f service.yaml -e prod -o .pltf/example/payments/workspace
  pltf generate -f service.yaml -e dev -m ./custom-mods --var cluster_name=my-dev
  ```

## Module commands

- `pltf module list|get` examine modules from the embedded root or `--modules` override.
- `pltf module init` generates `module.yaml` metadata (`--force` overwrites).
- Example: `pltf module list -m ./modules -o json`

## Images

- `pltf image build` runs a Dagger session that builds every declared image (optionally pushing via `--push`).
- Builds run concurrently and reuse the `pltf-image-cache` volume for BuildKit layers.
- Images may declare a `platforms` list (`["linux/amd64","linux/arm64"]` style) to target other OS/ARCH combos; when omitted, the host OS/ARCH is used.
- Example: `pltf image build -f service.yaml -e dev --push`
