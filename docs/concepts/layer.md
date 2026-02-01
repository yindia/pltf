# Layer (Service)

A deployable unit composed of modules, wired into an Environment.

## What is a Layer/Service?
Services let you manage app-specific resources separately from the shared foundation. They reference an Environment and define the modules they need (databases, queues, buckets, IAM, charts), with per-environment overrides.

<div class="mermaid">
flowchart TB
    svc[(service.yaml)]

    subgraph PROD[Production Env]
        prod_service[Service A]
    end

    subgraph STAGE[Staging Env]
        stage_service[Service A]
    end

    env[(env.yaml)]

    svc --> prod_service
    svc --> stage_service

    prod_service --> env
    stage_service --> env
</div>

## Definition (example)
Based on a typical service spec:

```yaml
apiVersion: platform.io/v1
kind: Service

metadata:
  name: payments-api
  ref: ./env.yaml
  envRef:
    prod: {}
secrets:
  api_key:
    key: api_key
variables:
  db_name: "testing"
modules:
  - id: postgres
    type: aws_postgres
    inputs:
      database_name: "${var.db_name}"
  - id: s3
    type: aws_s3
    inputs:
      bucket_name: "pltf-app-${layer_name}-${env_name}"
    links:
      readWrite:
        - adminpltfrole
        - userpltfrole
  - id: topic
    type: aws_sns
    inputs:
      sqs_subscribers:
        - "${module.notifcationsQueue.queue_arn}"
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
# Add helm chart modules  
```


## Notes

- `${layer_name}` comes from the service name; `${env_name}` is the selected environment key.
- `links` let modules consume other module outputs (e.g., queue ARNs, IAM roles) without manual interpolation.
- `envRef` selects envs only; variables and secrets live at the top level.

## When to use Services
- Isolate app stacks (DB + queues + roles) from the shared environment.
- Share one Environment across multiple Services without duplicating YAML.
- Enable per-team or per-PR stacks while keeping consistent wiring.

## Next steps
- Revisit the [Environment](environment.md) foundations.
- Dive into module APIs in [References](../references/aws.md).
