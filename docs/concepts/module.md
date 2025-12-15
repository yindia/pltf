# Module

A high-level building block to provision infrastructure.

## What is a Module?
pltf includes an embedded library of modules you can connect to build your stack. Each module is a high-level construct that provisions the resources needed to achieve its goal (e.g., EKS cluster, S3 bucket, Postgres).

Modules are described by `module.yaml` (type/provider/version/inputs/outputs) and referenced in your Environment or Service spec.

![Module](../images/hero.png) <!-- Replace with module graphic -->

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
Generate `module.yaml` for your own Terraform module via `pltf module init --path <module_dir> [--force]`. Use `source: custom` in specs and provide `--modules` (or profile `modules_root`) to load them.

### Terraform compatible
pltf uses Terraform under the hood, so you’re never locked in. Extend with your own Terraform or take the generated code with you.

## Next Steps
- Learn about [Layer/Service](service.md) (coming soon).
- Explore the module API in [References](../references/aws.md) and per-module pages.
