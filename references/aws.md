# AWS Reference

The next generation of Infrastructure-as-Code: work with high-level constructs instead of getting lost in low-level cloud configuration. Status: active development; review generated code before applying.

AWS is fully supported for environments, services, and modules. This page summarizes how the AWS provider, backends, and module wiring work in pltf.

![AWS](../images/hero.png) <!-- Replace with an AWS-specific image if desired -->

## Provider and Backends
- **Provider:** Automatically injected; version comes from the central versions file. Region is taken from your env spec.
- **Backends:** You can store state in `s3`, `gcs`, or `azurerm` even when targeting AWS. For cross-account S3, set `backend.profile`. Optional `backend.region` overrides the bucket region.
- **Default tags:** Labels in your env/service specs become global tags on the AWS provider.

## Example (Environment + Service)
```yaml
apiVersion: platform.io/v1
kind: Environment
metadata:
  name: example-aws
  org: pltf
  provider: aws
  labels:
    team: platform
    cost_center: shared
environments:
  prod:
    account: "556169302489"
    region: us-east-1
    backend:
      type: s3
      profile: cross-account
    modules:
      - type: aws_base
      - type: aws_eks
      - type: aws_k8s_base
```

```yaml
apiVersion: platform.io/v1
kind: Service
metadata:
  name: payments-api
  org: pltf
  provider: aws
  envRef:
    name: prod
    path: ../env.yaml
spec:
  variables:
    image: ghcr.io/acme/payments:latest
  modules:
    - name: app
      type: aws_k8s_service
      port:
        http: 8080
      links:
        - app-bucket: [write]
        - app-queue: [consume]
    - name: app-bucket
      type: aws_s3
      bucket_name: "payments-${env_name}"
    - name: app-queue
      type: aws_sqs
```

## Modules and Fields
- **Fields:** Each module instance accepts inputs declared in its `module.yaml`. Only set what you need; defaults apply otherwise.
- **Names:** `name` is optional; defaults to the module `type`. Names are used for Terraform resource names and template placeholders.
- **Types:** `type` selects the module implementation. Embedded AWS modules are documented under “Modules (AWS)” in the nav.
- **Sources:** Add `source: custom` to pull a module from your custom modules root; otherwise the embedded catalog is used.

## Linking
Linking lets a module consume outputs of another:
```yaml
links:
  - app-bucket: [read, write]
  - app-queue: [consume]
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
