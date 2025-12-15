# Secrets

Keep sensitive values out of specs and source control.

## What it does
- Secrets stay as Terraform variables and render to Kubernetes secrets for services.
- You declare secret keys in your spec; actual values are provided at runtime via environment variables or `--var`, typically sourced from your secret store/CI.
- Services receive secrets as env vars; no values are written into locals or files.

## Example (service)
```yaml
spec:
  secrets:
    db_password: {}   # value supplied via env/CI
  modules:
    - type: aws_k8s_service
      name: app
      env_vars:
        - name: DB_PASSWORD
          value: "${var.db_password}"
```
Runtime:
```bash
PLTF_VAR_db_password=supersecret pltf terraform apply -f service.yaml -e prod
```

## Notes
- Prefer env/CI secret stores; do not commit secret values to specs or repos.
- Services restart to pick up new secret values after apply; plan rotations accordingly.
