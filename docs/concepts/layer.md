# Layer (Service)

A deployable unit composed of modules, wired into an Environment.

## What is a Layer/Service?
Services let you manage app-specific resources separately from the shared foundation. They reference an Environment and define the modules they need (databases, queues, buckets, IAM, charts), with per-environment overrides.

![Layer](../images/hero.png)

## Definition (example)
Based on `example/service.yaml`:

```yaml
apiVersion: platform.io/v1
kind: Service

metadata:
  name: payments-api
  ref: ./env.yaml
  envRef:
    prod:
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
# Add helm chart modules  
```


## Notes

- `${layer_name}` comes from the service name; `${env_name}` is the selected environment key.
- `links` let modules consume other module outputs (e.g., queue ARNs, IAM roles) without manual interpolation.
- Per-environment overrides go under `envRef` to scope variables/secrets.

## When to use Services
- Isolate app stacks (DB + queues + roles) from the shared environment.
- Share one Environment across multiple Services without duplicating YAML.
- Enable per-team or per-PR stacks while keeping consistent wiring.

## Next steps
- Revisit the [Environment](environment.md) foundations.
- Dive into module APIs in [References](../references/aws.md).
