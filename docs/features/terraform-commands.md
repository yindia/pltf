# Terraform Commands

Run Terraform with consistent, auto-generated configs.

## What it does
- Commands live under `pltf terraform plan|apply|destroy|output|force-unlock`.
- Auto-generates Terraform (providers, backends, modules, outputs) before running TF.
- Ensures the backend bucket/container exists (S3/GCS/Azurerm) before init/apply.
- Passes through standard TF flags (targets, parallelism, lock, no-color, plan file, detailed exit codes).

## Examples
```bash
pltf terraform plan -f service.yaml -e dev --detailed-exitcode --plan-file=/tmp/plan.tfplan
pltf terraform apply -f env.yaml -e prod
pltf terraform destroy -f env.yaml -e prod
pltf terraform output -f service.yaml -e dev --json
pltf terraform force-unlock -f env.yaml -e prod --lock-id=12345
pltf terraform plan -f env.yaml -e prod --scan        # run tfsec on generated TF
```

### Visualize plans with Rover
- Rover (https://github.com/yindia/rover) renders Terraform plans as an interactive UI.
- Generate a plan JSON and launch Rover in one go:
  ```bash
  pltf terraform plan -f env.yaml -e prod --rover
  ```
- The CLI will:
  - create a `.tfplan` and `.json` in the generated output dir,
  - spin up Rover (using the Terraform binary from `PATH`), and
  - serve the UI on `0.0.0.0:9000`. Check the log output for the exact URL.
  
## Notes
- Backends are decoupled from provider (`s3|gcs|azurerm` supported).
- Common flags: `--target/-t`, `--parallelism/-p`, `--lock/-l`, `--lock-timeout/-T`, `--no-color/-C`, `--input/-i`, `--refresh/-r`, `--plan-file/-P`, `--detailed-exitcode/-d`, `--json/-j`.
- Security lint: `--scan` runs tfsec against the generated Terraform before returning.
- Uses the same generation path as `pltf generate`; you can inspect the rendered TF in the output directory.
