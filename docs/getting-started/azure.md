# Getting Started: Azure

Follow this walkthrough to go from a simple Azure Environment spec to a working AKS cluster and service wiring.

## 1) Prerequisites

- Terraform 1.6.6 locally (or any version satisfying `>=1.6.6` in the workspace).
- Azure credentials (`az login`, or a service principal).
- `pltf` installed (see [Installation](../installation.md)).

## 2) Render the Environment (VNet + AKS)

Create `env.yaml`:

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

That config boots:

- A VNet/subnet, key vault, logging, and ACR via `azure_base`.
- An AKS cluster via `azure_aks` (RBAC + OMS agent enabled).

Run:

```bash
pltf validate -f env.yaml --env dev
pltf terraform plan  -f env.yaml --env dev
pltf terraform apply -f env.yaml --env dev
```

## 3) Add a Service (Helm + Postgres)

Create `service.yaml`:

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
