# Platform Usage

Use this page as a practical guide to the most common flows in pltf.

## Validate + Lint
```bash
pltf validate -f env.yaml -e prod
pltf validate -f service.yaml -e dev
```
- Runs structural validation and lint suggestions (labels, unused vars).
- Picks environment from `--env`, `PLTF_DEFAULT_ENV`, or profile `default_env`.

## Preview
```bash
pltf preview -f env.yaml -e prod
```
- Shows provider, backend type, labels, and modules without running Terraform.

## Generate (Terraform only)
```bash
pltf generate -f env.yaml -e dev
pltf generate -f service.yaml -e prod -m ./modules --out .pltf/service/prod
pltf generate -f service.yaml -e dev --var cluster_name=my-dev
```
- `--modules/-m` custom root; modules with `source: custom` resolve here first.
- `--out/-o` defaults to `.pltf/<env_name>/env/<env>` or `.pltf/<env_name>/<service>/<env>`.
- `--var/-v` merges over env vars → service envRef vars → CLI vars.
- File inputs pointing to existing files in the spec directory are copied into the output and paths are updated.

## Terraform commands
```bash
pltf terraform plan    -f service.yaml -e dev    # supports --target, --parallelism, --detailed-exitcode, --plan-file
pltf terraform apply   -f env.yaml    -e prod
pltf terraform destroy -f env.yaml    -e prod
pltf terraform output  -f service.yaml -e dev --json
pltf terraform force-unlock -f env.yaml -e prod --lock-id=<id>
```
- Automatically generates Terraform, ensures backend (S3/GCS/Azurerm).
- Common flags: `--target/-t`, `--parallelism/-p`, `--lock/-l`, `--lock-timeout/-T`, `--no-color/-C`, `--input/-i`, `--refresh/-r`, `--plan-file/-P`, `--detailed-exitcode/-d`, `--json/-j`.

## Module inventory
```bash
pltf module list [-m ./modules] [-o table|json|yaml]
pltf module get aws_eks [-m ./modules] [-o table|json|yaml]
pltf module init --path ./modules/aws_eks [--force]
```
- Use `source: custom` in specs to force lookup from your custom root (`--modules` or profile `modules_root`); embedded modules remain available.

## Profiles & Defaults
- `~/.pltf/profile.yaml` (or `PLTF_PROFILE`) can set `modules_root`, `default_env`, `default_out`, `telemetry`.
- `PLTF_DEFAULT_ENV` is also respected for picking the environment.

## Backends
- `backend.type` can be `s3|gcs|azurerm` (independent of provider).
- `backend.profile` supports cross-account S3; optional `region`, `container`, `resource_group`.
