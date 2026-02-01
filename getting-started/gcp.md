# Getting Started: GCP

Follow this walkthrough to go from a simple GCP Environment spec to a working GKE cluster and service wiring.

## 1) Prerequisites

- Terraform 1.6.6 locally (or any version satisfying `>=1.6.6` in the workspace).
- GCP credentials (`gcloud auth application-default login` or `GOOGLE_APPLICATION_CREDENTIALS`).
- `pltf` installed (see [Installation](../installation.md)).

## 2) Render the Environment (VPC + GKE)

Create `env.yaml`:

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

That config boots:

- A VPC and subnet via `gcp_base`.
- A GKE control plane plus default node pool via `gcp_gke`.

Now run:

```bash
pltf validate -f env.yaml --env dev
pltf terraform plan  -f env.yaml --env dev
pltf terraform apply -f env.yaml --env dev
```

## 3) Add a Service (GCS + Service Account)

Create `service.yaml`:

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

This grants the service account access to the bucket using `links`, wired through `read_buckets`/`write_buckets`.

Run:

```bash
pltf validate        -f service.yaml --env dev
pltf terraform plan  -f service.yaml --env dev
pltf terraform apply -f service.yaml --env dev
```

## 4) Cleanup

```bash
pltf terraform destroy -f service.yaml --env dev
pltf terraform destroy -f env.yaml --env dev
```
