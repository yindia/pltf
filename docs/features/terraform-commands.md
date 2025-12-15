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
```

## Notes
- Backends are decoupled from provider (`s3|gcs|azurerm` supported).
- Common flags: `--target/-t`, `--parallelism/-p`, `--lock/-l`, `--lock-timeout/-T`, `--no-color/-C`, `--input/-i`, `--refresh/-r`, `--plan-file/-P`, `--detailed-exitcode/-d`, `--json/-j`.
- Uses the same generation path as `pltf generate`; you can inspect the rendered TF in the output directory.
