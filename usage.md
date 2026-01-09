# CLI Usage

pltf auto-detects whether a spec is an **Environment**, **Service**, or **Stack** based on `kind`. Most commands accept `--file/-f`, `--env/-e`, `--modules/-m`, `--out/-o`, and `--var/-v key=value`.

## Command catalog
- `pltf validate` — validate specs.
- `pltf generate` — render a workspace-style Terraform root (single env with `--env`, all envs if omitted).
- `pltf preview` — summarize provider/backend/labels/modules.
- `pltf terraform plan|apply|destroy|output|force-unlock|graph` — generate + run Terraform/OpenTofu with standard TF flags (use `--engine tofu` for OpenTofu).
- `pltf module list|get|init` — module inventory and metadata generation.

### validate
- **What:** Validate Environment, Service, or Stack specs; auto-detects kind.
- **Flags:**
  - `--file/-f` — Path to the spec (default `env.yaml`).
  - `--env/-e` — Environment key (dev/prod/etc.).
- **Example:** `pltf validate -f service.yaml -e dev`

### generate
- **What:** Render workspace-ready Terraform only; no init/apply.
- **Flags:**
  - `--file/-f` — Path to the spec.
  - `--env/-e` — Environment key.
  - `--modules/-m` — Custom modules root (local path or git ref); `source: custom` resolves here first.
  - `--out/-o` — Output dir (defaults to `.pltf/<env_name>/workspace` or `.pltf/<env_name>/<service>/workspace`).
  - `--var/-v` — CLI var override `key=value`.
- **Example:** `pltf generate -f service.yaml -e prod -m ./modules --out .pltf/example/payments/workspace`

### preview
- **What:** Show provider, backend, labels, modules without running Terraform.
- **Flags:** `--file/-f`, `--env/-e`
- **Example:** `pltf preview -f env.yaml -e prod`

### terraform plan
- **What:** Generate + run `terraform plan` with standard flags (or `tofu` via `--engine tofu`).
- **Plan flags:**
  - `--target/-t` — Target address (repeatable).
  - `--parallelism/-p` — Max parallel operations.
  - `--lock/-l` — Lock state (default true).
  - `--lock-timeout/-T` — Lock timeout (e.g., 30s).
  - `--no-color/-C` — Disable color output.
  - `--input/-i` — Prompt for input (default false).
  - `--refresh/-r` — Refresh state before plan (default true).
  - `--detailed-exitcode/-d` — Enable TF detailed exit codes.
  - `--plan-file/-P` — Write plan to a file.
- **Shared flags:** `--file/-f`, `--env/-e`, `--modules/-m`, `--out/-o`, `--var/-v`, `--engine`
- **Example:** `pltf terraform plan -f service.yaml -e dev --detailed-exitcode --plan-file=/tmp/plan.tfplan`

### terraform apply
- **What:** Generate + run `terraform apply -auto-approve` (or `tofu` via `--engine tofu`).
- **Flags:** Shared flags (`--file/-f`, `--env/-e`, `--modules/-m`, `--out/-o`, `--var/-v`, `--engine`).
- **Example:** `pltf terraform apply -f env.yaml -e prod`

### terraform destroy
- **What:** Generate (if needed) + run `terraform destroy -auto-approve` (or `tofu` via `--engine tofu`).
- **Flags:** Same as apply.
- **Example:** `pltf terraform destroy -f env.yaml -e prod`

### terraform output
- **What:** Show outputs (optionally JSON).
- **Flags:** `--json/-j` (JSON output), plus shared `--file/-f`, `--env/-e`, `--modules/-m`, `--out/-o`, `--engine`
- **Example:** `pltf terraform output -f service.yaml -e dev --json`

### terraform force-unlock
- **What:** Force unlock state.
- **Flags:** `--lock-id` (required), plus shared `--file/-f`, `--env/-e`, `--modules/-m`, `--out/-o`, `--engine`
- **Example:** `pltf terraform force-unlock -f env.yaml -e prod --lock-id=12345`

### terraform graph
- **What:** Emit DOT graph. Default runs `terraform graph`; `--mode spec` renders a dependency graph from the YAML only. Use `--engine tofu` to run `tofu graph`.
- **Flags:** `--mode terraform|spec` (default terraform), `--plan-file/-P` (passed to terraform graph), plus shared `--file/-f`, `--env/-e`, `--modules/-m`, `--out/-o`, `--engine`
- **Example:** `pltf terraform graph -f env.yaml -e dev | dot -Tpng > graph.png`

### module list
- **What:** List module inventory from embedded/custom roots.
- **Flags:** `--modules/-m` (modules root), `--output/-o` (`table|json|yaml`)
- **Example:** `pltf module list -m ./modules -o json`

### module get
- **What:** Show module details (inputs/outputs).
- **Flags:** Same as module list.
- **Example:** `pltf module get aws_eks -m ./modules`

### module init
- **What:** Generate `module.yaml` from an existing Terraform module dir.
- **Flags:** `--path` (module dir), `--name`, `--type`, `--description`, `--out`, `--force` (overwrite)
- **Example:** `pltf module init --path ./modules/aws_eks --force`

## Validate
Structural validation for Environment, Service, and Stack specs.
```bash
pltf validate -f env.yaml -e prod
pltf validate -f service.yaml -e dev
```

## Generate
Render Terraform without running it. File inputs that point to existing files (relative to the spec) are copied into the output directory and paths are updated.
```bash
pltf generate -f env.yaml -e dev
pltf generate -f service.yaml -e prod -o .pltf/example/payments/workspace
pltf generate -f service.yaml -e dev -m ./custom-mods --var cluster_name=my-dev
```
Flags:
- `--modules/-m` custom modules root (local path or git ref). Modules with `source: custom` are resolved only from the custom root; others fall back to embedded modules.
- `--out/-o` output dir (defaults `.pltf/<env_name>/workspace` or `.pltf/<env_name>/<service>/workspace`).
- `--var/-v` merges vars (spec variables → CLI vars).

## Terraform helpers
Terraform commands live under `pltf terraform ...` and auto-generate before running TF. Use `--engine tofu` to run OpenTofu.
These commands now use a workspace-style root and pass `-var-file=<env>.tfvars` for the selected env; secrets remain in env/CI.
```bash
pltf terraform plan    -f service.yaml -e dev    # plan (supports --target, --parallelism, --detailed-exitcode, --plan-file)
pltf terraform apply   -f env.yaml    -e prod    # apply
pltf terraform destroy -f env.yaml    -e prod    # destroy
pltf terraform output  -f service.yaml -e dev    # outputs (--json supported)
pltf terraform force-unlock -f env.yaml -e prod --lock-id=<id>
```
Common flags: `--target/-t`, `--parallelism/-p`, `--lock/-l`, `--lock-timeout/-T`, `--no-color/-C`, `--input/-i`, `--refresh/-r`, `--plan-file/-P`, `--detailed-exitcode/-d`, `--json/-j`.

## Preview
Quick summary (provider, backend, labels, modules) without TF.
```bash
pltf preview -f env.yaml -e prod
```

## Module inventory
```bash
pltf module list [-m ./modules] [-o table|json|yaml]
pltf module get aws_eks [-m ./modules] [-o table|json|yaml]
pltf module init --path ./modules/aws_eks [--force]
```

## Custom backends
In env spec `backend.type` can be `s3|gcs|azurerm` regardless of provider. Optional `region`, `container`, `resource_group`, `profile` (S3) supported.

## Custom modules
- Mark a module with `source: custom` to force lookup in your custom modules root.
- Provide a custom root via `--modules` or profile `modules_root`; embedded modules remain available for everything else.
- Generate module.yaml for your own TF module with `pltf module init --path <module_dir> [--force]`.

## Environment defaults
`PLTF_DEFAULT_ENV` or profile `default_env` let you omit `--env` when only one environment applies.

## Completions
```bash
pltf completion bash|zsh|fish|powershell
```
