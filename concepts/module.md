# Module

A high-level building block to provision infrastructure.

## What is a Module?
pltf includes an embedded library of modules you can connect to build your stack. Each module is a high-level construct that provisions the resources needed to achieve its goal (e.g., EKS cluster, S3 bucket, Postgres).

Modules are described by `module.yaml` (type/provider/version/inputs/outputs) and referenced in your Environment or Service spec.

## Module metadata (module.yaml)
Every module has a `module.yaml` descriptor. The generator uses it to validate inputs, wire outputs, and decide provider requirements.

Common fields:
- `name`: human-friendly name.
- `type`: unique module type used in specs.
- `provider`: cloud provider (`aws`, `gcp`, `azure`).
- `version`: module catalog version (not Terraform version).
- `cluster`: `true` if the module exposes a Kubernetes cluster.
- `inputs`: list of input specs (name/type/required/default/description).
- `outputs`: list of output specs (name/type/description).
- `capabilities`: optional `provides`/`accepts` tags for future routing.

Input/output types should mirror Terraform (e.g., `string`, `number`, `bool`, `list(string)`, `map(string)`).

## Cluster modules (Kubernetes/Helm/Kustomize)
Modules that provide a Kubernetes cluster must declare `cluster: true` in `module.yaml` and expose these outputs:
- `k8s_endpoint`
- `k8s_ca_data`
- `k8s_cluster_name`
- `plt_cluster_type` (bool)

Only one module marked `cluster: true` can exist across the env+service stack (including referenced stacks). The CLI wires Kubernetes/Helm/Kustomize providers against this module.

## IAM module contract
To enable IAM auto-wiring and trust policy augmentation, IAM modules must declare the following:
- `iam.role` providers: inputs `iam_policy`, `kubernetes_trusts`, output `role_arn`
- `iam.user` providers: input `iam_policy`, output `user_arn`
- `iam.policy` providers: output `policy_arn`

## Custom module guidelines
When writing custom modules, follow these rules to keep auto-wiring predictable:
- **Cluster modules:** use `cluster: true` and expose `k8s_endpoint`, `k8s_ca_data`, `k8s_cluster_name`, `plt_cluster_type`.
- **IAM modules:** follow the IAM contract above (`iam.role`, `iam.user`, `iam.policy`).
- **Auto-wiring:** if one module should feed another, make the **input name** match the **output name** exactly.
- **No duplicate outputs:** output names must be unique across the entire merged stack (stacks + env + service).

### Auto-wiring details
pltf matches inputs to outputs by name across the merged module set:
- For environment stacks, wiring happens within the env modules.
- For services, wiring can reference service modules first, then env outputs via remote state.
- If multiple modules export the same output name, validation fails.

If a module input has a value explicitly set in YAML, that value wins and no auto-wiring occurs.

### Stack + env + service merge
Stacks are merged into env/service before validation and generation:
- stack modules are added first, then env/service modules.
- stack variables/labels/providers are merged and then env/service values apply.
- overrides are not allowed: if a stack defines a module ID or variable name, the env/service cannot redefine it.

### Custom module example (cluster)
```yaml
name: my_eks
type: my_eks
provider: aws
version: 1.0.0
cluster: true
inputs:
  - name: cluster_name
    type: string
    required: true
outputs:
  - name: k8s_endpoint
    type: string
  - name: k8s_ca_data
    type: string
  - name: k8s_cluster_name
    type: string
  - name: plt_cluster_type
    type: bool
```

### Custom module example (IAM role)
```yaml
name: my_iam_role
type: my_iam_role
provider: aws
version: 1.0.0
inputs:
  - name: iam_policy
    type: string
    required: true
  - name: kubernetes_trusts
    type: list(string)
    required: false
outputs:
  - name: role_arn
    type: string
```

## Definition
Modules have:
- a **type** (e.g., `aws_eks`, `aws_s3`)
- an optional **id/name** (so you can include multiple of the same type)
- optional **inputs** (configuration)
- optional **links** (to consume other module outputs)
- optional **source** (`custom` forces lookup in your custom modules root)

Modules are defined inside the `modules` section of an Environment or Service.

### Minimal configuration
We built pltf so you can provision a resource with a single line. Defaults follow best practices; customize only what you need.

```yaml
modules:
  - id: cluster
    type: aws_eks
  - id: db
    type: aws_postgres
```

### Extra configuration
Override only the fields you care about; pltf uses recommended defaults otherwise.

```yaml
modules:
  - id: devcluster
    type: aws_eks
    inputs:
      node_instance_type: t3.medium
      max_nodes: 5
      spot_instances: true
  - id: dbfrontend
    type: aws_postgres
    inputs:
      instance_class: db.t3.medium
      engine_version: "12.4"
```

### Links (module outputs as inputs)
Modules can consume outputs from others using `links` or direct references like `${module.redis.cache_host}`.

```yaml
modules:
  - id: redis
    type: aws_redis
  - id: airflow
    type: helm_chart
    inputs:
      repository: https://airflow.apache.org
      chart: airflow
      namespace: airflow
      chart_version: 1.4.0
      values:
        brokerUrl: "rediss://:${module.redis.cache_auth_token}@${module.redis.cache_host}"
```

### Custom modules
Generate `module.yaml` for your own Terraform module via `pltf module init --path <module_dir> [--force]`. Use `source: custom` in specs, point at a shared catalog via `--modules` (or profile defaults) if you wish, or provide a git URL for the new per-module workflow to fetch metadata automatically.

### Terraform compatible
pltf uses Terraform under the hood, so you’re never locked in. Extend with your own Terraform or take the generated code with you.

## Next Steps
- Learn about [Layer/Service](layer.md).
- Explore module APIs in [References](../references/aws.md) and per-module pages.
