# Azure Reference

# Azure Reference

Azure is supported for environments, services, and the embedded module catalog. This page summarizes provider configuration, backends, and module wiring for Azure.

## Provider and Backends
- **Provider:** `provider: azure` (or `azurerm`) in the spec. Region comes from the selected environment entry.
- **Subscription:** `environments.<env>.account` is treated as the Azure subscription ID.
- **Backends:** `backend.type: azurerm` (default if omitted). Provide `backend.bucket` (storage account name) and optionally `backend.container`/`backend.resource_group`.

## Example (Environment + Service)
Environment:
```yaml
apiVersion: platform.io/v1
kind: Environment
metadata:
  name: example-azure
  org: pltf
  provider: azure
  labels:
    team: platform
    cost_center: shared
environments:
  dev:
    account: "00000000-0000-0000-0000-000000000000"
    region: eastus
variables:
  location: eastus
modules:
  - id: base
    type: azure_base
    inputs:
      location: "${var.location}"
  - id: aks
    type: azure_aks
    inputs:
      cluster_name: "pltf-${env_name}"
```

Service:
```yaml
apiVersion: platform.io/v1
kind: Service
metadata:
  name: payments
  ref: ./env.yaml
  envRef:
    dev: {}
modules:
  - id: api
    type: helm_chart
    inputs:
      chart: ./services/payments/chart
      values:
        replicas: 2
  - id: db
    type: azure_postgres
```

## Modules and Fields
- **id:** required and unique within the stack.
- **type:** selects the module implementation; required unless `source` is a git/local path with `module.yaml`.
- **source:** optional; git/paths load metadata directly.
- **inputs:** key/value config for module variables.
- **links:** access bindings used to connect modules where supported.

## Template placeholders
- `${env_name}` and `${layer_name}` become the resolved environment/service names.
- `${module.<module_name>.<output_name>}` references another module’s output.
- `${parent.<output_name>}` references outputs from the parent environment when authoring a service.
- `${var.<name>}` references variables defined in the spec or via `--var`.

## Useful commands
- `pltf module list -o table` — see available Azure modules.
- `pltf module get azure_aks` — inspect inputs/outputs.
- `pltf generate -f env.yaml -e dev` — render Terraform for Azure.
- `pltf terraform plan/apply ...` — generate + execute Terraform.

See the module-specific pages under “Modules (Azure)” for detailed inputs, outputs, and examples.
