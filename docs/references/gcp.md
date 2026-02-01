# GCP Reference

GCP is supported for environments, services, and the embedded module catalog. This page summarizes provider configuration, backends, and module wiring for GCP.

## Provider and Backends
- **Provider:** `provider: gcp` (or `google`) in the spec. Region comes from the selected environment entry.
- **Project:** `environments.<env>.account` is treated as the GCP project ID.
- **Backends:** `backend.type: gcs` (default if omitted). Set `backend.bucket` to use an existing state bucket.

## Example (Environment + Service)
Environment:
```yaml
apiVersion: platform.io/v1
kind: Environment
metadata:
  name: example-gcp
  org: pltf
  provider: gcp
  labels:
    team: platform
    cost_center: shared
environments:
  dev:
    account: "pltf-dev-project"
    region: us-central1
modules:
  - id: base
    type: gcp_base
  - id: gke
    type: gcp_gke
    inputs:
      cluster_name: "pltf-${env_name}"
      node_zone_names:
        - us-central1-a
        - us-central1-b
```

Service:
```yaml
apiVersion: platform.io/v1
kind: Service
metadata:
  name: data-jobs
  ref: ./env.yaml
  envRef:
    dev: {}
modules:
  - id: data-bucket
    type: gcp_gcs
    inputs:
      bucket_name: "pltf-data-${env_name}"
    links:
      readwrite: data-sa
  - id: data-sa
    type: gcp_service_account
```

## Modules and Fields
- **id:** required and unique within the stack.
- **type:** selects the module implementation; required unless `source` is a git/local path with `module.yaml`.
- **source:** optional; git/paths load metadata directly.
- **inputs:** key/value config for module variables.
- **links:** access bindings used to connect modules (for example, GCS buckets to service accounts).

## Linking
Linking lets a GCS module grant access to a service account:
```yaml
links:
  readwrite:
    - data-sa
```
For GCP, pltf maps these links to `read_buckets`/`write_buckets` inputs on `gcp_service_account` (or `gcp_k8s_service`).

## Template placeholders
- `${env_name}` and `${layer_name}` become the resolved environment/service names.
- `${module.<module_name>.<output_name>}` references another module’s output.
- `${parent.<output_name>}` references outputs from the parent environment when authoring a service.
- `${var.<name>}` references variables defined in the spec or via `--var`.

## Useful commands
- `pltf module list -o table` — see available GCP modules.
- `pltf module get gcp_gke` — inspect inputs/outputs.
- `pltf generate -f env.yaml -e dev` — render Terraform for GCP.
- `pltf terraform plan/apply ...` — generate + execute Terraform.

See the module-specific pages under “Modules (GCP)” for detailed inputs, outputs, and examples.
