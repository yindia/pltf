# Getting Started: AWS

Follow this walkthrough to go from the checked-in sample (`example/e2e.yaml`) to a working Kubernetes-native AWS stack. The guide shows how services reference envs, how Docker images build via Dagger, and how Terraform runs natively on the host.

## 1) Prerequisites

- Terraform 1.6.6 locally (or any version satisfying `>=1.6.6` in the workspace).  
- AWS credentials (from `aws configure`, env vars, or another provider).  
- `pltf` installed (see [Installation](../installation.md)).  
- Dagger installed (only required when building/pushing images via `pltf image ...` or when specs declare Docker images).

## 2) Render the Environment (VPC + EKS + DNS)

The sample `example/e2e.yaml` already wires AWS base, DNS, and EKS modules into the `prod` environment. Copy it to `env.yaml` and adjust variables as needed.

```yaml
apiVersion: platform.io/v1
kind: Environment

metadata:
  name: example-aws
  org: pltf
  provider: aws
variables:
  base_domain: prod.pltf.internal
  cluster_name: pltf-data
environments:
  prod:
    account: "556169302489"
    region: ap-northeast-1
modules:
  - id: base
    type: aws_base
  - id: dns
    type: aws_dns
    inputs:
      domain: ${{var.base_domain}}
      delegated: false
  - id: eks
    type: aws_eks
    inputs:
      cluster_name: "pltf-app-${layer_name}-${env_name}"
      k8s_version: 1.33
      enable_metrics: false
      max_nodes: 15
  - id: nodegroup1
    type: aws_nodegroup
    inputs:
      max_nodes: 15
      node_disk_size: 20
```

That config boots:

- A VPC/subnets/security groups via `aws_base`.  
- A DNS zone and records via `aws_dns`.  
- An EKS control plane plus one nodegroup (`aws_eks`, `aws_nodegroup`).  

Now run:

```bash
pltf validate -f example/e2e.yaml --env prod
pltf terraform plan  -f example/e2e.yaml --env prod
pltf terraform apply -f example/e2e.yaml --env prod
```

 Terraform runs operate inside `.pltf/<environment_name>/workspace` so the standard `.terraform` cache keeps provider downloads per workspace. `pltf terraform plan` builds Docker images (no push) before planning; `apply` builds + pushes them using your host registry credentials.

## 3) Add a Service (Postgres + S3 + SNS/SQS + IAM)

Create `service.yaml` to reference your environment and bind service modules, variables, and secrets.

```yaml
apiVersion: platform.io/v1
kind: Service

metadata:
  name: payments-api
  ref: ./env.yaml
  envRef:
    prod: {}
variables:
  db_name: "testing"
secrets:
  api_key:
    key: api_key
modules:
  - id: postgres
    type: aws_postgres
    inputs:
      database_name: "${{var.db_name}}"
  - id: s3
    type: aws_s3
    inputs:
      bucket_name: "pltf-app-${layer_name}-${env_name}"
    links:
      readWrite: adminpltfrole
      readWrite: userpltfrole
  - id: topic
    type: aws_sns
    inputs:
      sqs_subscribers:
        - "${{module.notifcationsQueue.queue_arn}}"
    links:
      read: adminpltfrole
  - id: notifcationsQueue
    type: aws_sqs
    inputs:
      fifo: false
    links:
      readWrite: adminpltfrole
  - id: schedulesQueue
    type: aws_sqs
    inputs:
      fifo: false
    links:
      readWrite: adminpltfrole
  - id: adminpltfrole
    type: aws_iam_role
    inputs:
      extra_iam_policies:
        - "arn:aws:iam::aws:policy/CloudWatchEventsFullAccess"
      allowed_k8s_services:
        - namespace: "*"
          service_name: "*"
  - id: userpltfrole
    type: aws_iam_role
    inputs:
      extra_iam_policies:
        - "arn:aws:iam::aws:policy/CloudWatchEventsFullAccess"
      allowed_k8s_services:
        - namespace: "*"
          service_name: "*"
```

This adds:

- Postgres plus service-scoped DB name.  
- An S3 bucket named after `layer_name`/`env_name`.  
- An SNS topic, two SQS queues, and IAM roles wired via `links`.  
- Service-scoped variables and secrets.  

Run:

```bash
pltf validate        -f service.yaml --env prod
pltf terraform plan  -f service.yaml --env prod
pltf terraform apply -f service.yaml --env prod
```

Inspect outputs/graphs:

```bash
pltf terraform output -f service.yaml --env prod
pltf terraform graph  -f service.yaml --env prod | dot -Tpng > graph.png
```

## 4) Cleanup

```bash
pltf terraform destroy -f service.yaml --env prod
pltf terraform destroy -f example/e2e.yaml --env prod
```

## 5) Extend the stack

- Add Helm charts (Flyte, Argo) that rely on the IAM roles and buckets you created.  
- Drop in more modules (Redis, SES, DocumentDB) and wire them with `links`.  
- Use profile/remote backend configuration to match your AWS org structure.  
