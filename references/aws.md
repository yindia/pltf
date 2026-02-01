# AWS Reference

AWS is fully supported for environments, services, and the embedded module catalog. This page summarizes how the AWS provider, backends, and module wiring work in pltf.

## Provider and Backends
- **Provider:** Automatically injected; version comes from the central versions file. Region is taken from the selected environment entry.
- **Backends:** AWS uses `backend.type: s3` (default if omitted). Use `backend.profile` for cross-account state and `backend.region` to override the bucket region.
- **Default tags:** Labels in your env/service specs become global tags on the AWS provider.

## Example (Environment + Service)
Environment:
```yaml
apiVersion: platform.io/v1
kind: Environment
metadata:
  name: example-aws
  org: pltf
  provider: aws
  labels:
    team: platform
backend:
  type: s3
  bucket: pltf-tfstate
  region: us-east-1
environments:
  prod:
    account: "556169302489"
    region: us-east-1
variables:
  base_domain: prod.pltf.internal
modules:
  - id: base
    type: aws_base
  - id: eks
    type: aws_eks
    inputs:
      cluster_name: "pltf-app-${layer_name}-${env_name}"
```

Service:
```yaml
apiVersion: platform.io/v1
kind: Service
metadata:
  name: payments-api
  ref: ./env.yaml
  envRef:
    prod: {}
variables:
  image: ghcr.io/acme/payments:latest
modules:
  - id: app
    type: helm_chart
    inputs:
      chart: ./charts/payments
      values:
        image: var.image
  - id: app-bucket
    type: aws_s3
    inputs:
      bucket_name: "payments-${env_name}"
  - id: app-queue
    type: aws_sqs
```

## Modules and Fields
- **id:** required and unique within the stack.
- **type:** selects the module implementation; required unless `source` is a git/local path with `module.yaml`.
- **source:** optional; `custom` forces lookup in your custom modules root, while git/paths load metadata directly.
- **inputs:** key/value config for module variables.
- **links:** access bindings that let modules consume other module outputs (IAM policies/IRSA).

## Linking
Linking lets a module consume outputs of another:
```yaml
links:
  readWrite:
    - app-bucket
  consume:
    - app-queue
```
When links are present, pltf automatically renders IAM policies and (for Kubernetes) IRSA trusts. Supported AWS link targets include S3, SQS, SNS, SES, DynamoDB, RDS, and more via module metadata.

## Template placeholders
- `${env_name}` and `${layer_name}` become the resolved environment/service names.
- `${module.<module_name>.<output_name>}` references another module’s output.
- `${parent.<output_name>}` references outputs from the parent environment when authoring a service.
- `${var.<name>}` references variables defined in the spec or via `--var`.

## Useful commands
- `pltf module list -o table` — see available AWS modules.
- `pltf module get aws_eks` — inspect inputs/outputs.
- `pltf generate -f env.yaml -e prod` — render Terraform for AWS.
- `pltf terraform plan/apply ...` — generate + execute Terraform (plan/apply/destroy/output/force-unlock).

See the module-specific pages under “Modules (AWS)” for detailed inputs, outputs, and examples.
