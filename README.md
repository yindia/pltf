# pltf CLI 🚀

pltf (Platform Tools) turns your high-level infrastructure intent into ready-to-run Terraform workspaces with sensible defaults, instant validation, and inline security/cost scans. Instead of hand-editing Terraform everywhere, keep your cloud knowledge in reusable specs (`Environment`, `Service`, `Stack`) and let pltf generate the modules, providers, backends, and secrets wiring for you.

## Why teams choose pltf

- **Spec-first workflows** – describe your desired environments, services, and stacks once in YAML; pltf materializes the Terraform code, backend config, and provider mappings automatically.
- **Consistent Terraform runs** – every `pltf terraform …` command generates the workspace first, runs the host Terraform binary (no hidden runtime layers), and streamlines approvals (`plan` can include tfsec/cost/rover, `apply`/`destroy` always auto-approve).
- **Image-aware operations** – Docker images declared in specs are built once per plan/apply via Dagger, reusing the shared cache, and only `apply` pushes the tags so CI/CD remains in control.
- **Security and cost transparency** – builtin tfsec reporting and optional Infracost summaries show up alongside the Terraform logs so you don’t miss risky findings before you deploy.
- **Versioned, portable tooling** – the CLI stays Go-native, works alongside your existing git/terraform installs, and documents workflows for every team member.

## Specs & Concepts

- **Environment spec** (e.g., `env.yaml`) declares the cloud provider, backend, account/region entries, and the modules/variables that define the base infrastructure (VPCs, clusters, backend buckets) shared across dev, prod, and other deployments. Put shared inputs in `variables` and sensitive data in `secrets` so the base stack stays consistent without leaking secrets into Git.
- **Service spec** references an Environment via `metadata.ref` and adds workload-specific Terraform (modules, metadata, secrets, images, overrides) that run within that environment. `metadata.envRef` maps a service into any of the environment entries so a single service can target multiple environments while tuning behavior per env.
- **Stack spec** bundles reusable modules, inputs, and outputs with documented provider requirements. Environments include stacks with `metadata.stacks`, and stacks merge before generation so downstream specs can’t silently override their contracts.
- **Modules** can be the embedded modules shipped with pltf or your own custom Terraform modules. Place a `module.yaml` next to a custom module, refer to it in your spec, and pltf treats it the same as an embedded module during generation.
- **Variables & secrets** are first-class at every level. Shared values live in stacks/environments, while service-specific overrides and secrets stay in the service spec. Secrets never land in Git: pltf injects them when building the workspace.

Use services anytime you need to deploy workload-specific Terraform on top of the base environment (see the billing service in the Quickstart example). The service shares the common infra, builds its own images, and applies per-environment tweaks while reusing the generated workspace.

## Table of Contents

- [Install](#install)
- [Quickstart](#quickstart)
- [Terraform Workflows](#terraform-workflows)
- [Specs & Concepts](#specs--concepts)
- [Image & Terraform Caching](#image--terraform-caching)
- [Commands](#commands)
- [Behavior & Rules](#behavior--rules)
- [Provider Support](#provider-support)
- [Contributing](#contributing)

## Install

```bash
go build -o pltf ./main.go
go install ./...
```

## Quickstart

1. **Define reusable stacks** (`stack-cluster.yaml`):

   ```yaml
   apiVersion: platform.io/v1
   kind: Stack
   metadata:
     name: k8s-cluster
     labels:
       tier: infra
   providers:
     aws: true
   modules:
     - id: network
       type: aws_network
       inputs:
         cidr_block: 10.0.0.0/16
     - id: eks
       type: aws_eks
       inputs:
         cluster_name: var.cluster_name
         vpc_id: module.network.vpc_id
         subnet_ids: module.network.private_subnet_ids
     - id: observability
       type: aws_logging
       inputs:
         log_bucket: var.log_bucket
   outputs:
     - name: cluster_name
       value: module.eks.cluster_name
   ```

2. **Declare the base environment** (`env.yaml`) that references the stack, backend, variables, secrets, and multi-arch images:

   ```yaml
   apiVersion: platform.io/v1
   kind: Environment
   metadata:
     name: enterprise-aws
     org: acme
     provider: aws
     stacks:
       - ./stack-cluster.yaml
   backend:
     type: s3
     bucket: acme-tfstate
     region: us-west-2
   environments:
     dev:
       account: "111111111111"
       region: us-west-2
       variables:
         log_bucket: dev-logs
     prod:
       account: "222222222222"
       region: us-east-1
       variables:
         log_bucket: prod-logs
   variables:
     cluster_name: enterprise-cluster
     base_domain: example.com
   modules:
     - id: dns
       type: aws_dns
       inputs:
         domain: var.base_domain
   images:
     - name: platform-tools
       context: .
       platforms:
         - linux/amd64
         - linux/arm64
       tags:
         - ghcr.io/acme/platform-tools:${env_name}
   secrets:
     aws:
       type: file
       path: ~/.aws/credentials
   ```

   Point to your own Terraform modules by dropping a `module.yaml` next to them and referencing them alongside the embedded modules in your spec.

3. **Add a service** (`service.yaml`) that plugs into the environment and runs workload-specific modules/secrets across multiple envs:

   ```yaml
   apiVersion: platform.io/v1
   kind: Service
   metadata:
     name: billing
     ref: ./env.yaml
     envRef:
       dev: {}
       prod:
         variables:
           replica_count: 3
   modules:
     - id: api
       type: helm_chart
       inputs:
         chart: ./services/billing/chart
         repo: ./services/billing
         values:
           cluster: module.eks.cluster_name
           replicas: var.replica_count
   secrets:
     db_password:
       type: file
       path: ~/.vault/db-prod.txt
   images:
     - name: billing-api
       context: ./services/billing
       tags:
         - ghcr.io/acme/billing:${env_name}
   ```

4. **Preview, validate, and run Terraform**:

   ```bash
   ./pltf preview -f env.yaml -e dev
   ./pltf validate -f env.yaml -e prod --scan
   ./pltf generate -f env.yaml -e prod
   ./pltf terraform plan -f env.yaml -e prod --scan --cost
   ./pltf terraform apply -f env.yaml -e prod
   ```


## Terraform Workflows

- Terraform commands execute locally in the generated workspace; the CLI simply materializes Terraform files, copies `.tfvars`, and calls the host `terraform` binary (no special plugin cache or wrappers required).
- `pltf terraform plan` and `apply` build the declared Docker images via Dagger, reusing the `pltf-image-cache`. `apply` pushes the selected tags while `plan` stops after building locally. `destroy` skips image builds entirely.
- Commands like `plan` can still include optional helpers such as `--scan` (tfsec), `--cost` (Infracost), and `--rover`. `apply`/`destroy` always append `-auto-approve` so CI/CD can run unattended.
- Every command reuses the generated workspace in `.pltf/<spec-name>/<env>/workspace`, so plan/apply operate on the same graph and Terraform state.

## Image & Terraform Caching

- Image builds run through Dagger’s `Directory.DockerBuild`, mounting `pltf-image-cache` so BuildKit layers persist across commands. `platforms` lists drive multi-arch builds and default to the host OS/ARCH when unspecified.
- Terraform runs use the host binary and keep generated artifacts inside the workspace, so init/plan/apply/destroy all work against feature-parity Terraform state without extra caching layers.

## Commands

| Command                       | Description                                                                                           |
|-------------------------------|-------------------------------------------------------------------------------------------------------|
| `pltf generate`               | Materialize Terraform from an Environment or Service spec.                                           |
| `pltf validate`               | Lint a spec (`--scan` runs tfsec and prints findings).                                                |
| `pltf preview`                | Summarize provider/backend/stack/module wiring without executing Terraform.                         |
| `pltf version`                | Report pltf, Terraform, and key provider versions.                                                    |
| `pltf module init`            | Inspect a Terraform module and emit `module.yaml`.                                                   |
| `pltf module list`            | List embedded or custom modules available to specs.                                                  |
| `pltf module get`             | Show module inputs/outputs.                                                                          |
| `pltf terraform plan`         | Generate the workspace, build images, and run `terraform plan`. Supports `--scan`, `--cost`, `--rover`.|
| `pltf terraform apply`        | Build/push images and run `terraform apply` (auto-approved).                                          |
| `pltf terraform destroy`      | Skip image builds and execute `terraform destroy` (auto-approved).                                   |
| `pltf terraform output`       | Display Terraform outputs (`--json` for machine-readable output).                                    |
| `pltf terraform graph`        | Export dependency graphs (`--format`, `--outfile`).                                                   |
| `pltf terraform force-unlock` | Unlock a workspace state file (requires `--lock-id`).                                                 |

## Behavior & Rules

- Stacks merge before generation; stack modules and outputs cannot be overridden by environments/services.
- Environment/service variables only cover overrides after stacks merge; duplicate variable names throw errors.
- Providers are explicit; use the `providers` block to document requirements instead of inferring from module type.
- Auto-wiring matches outputs among the merged module set; conflicting outputs result in clear errors.
- Git refs in `metadata.ref` and `metadata.stacks` support remote specs (e.g., `https://host/org/repo.git//path/to/spec.yaml?ref=main`).

## Provider Support

| Provider | Status        |
|----------|---------------|
| AWS      | ✅ Supported   |
| GCP      | 🔜 Coming Soon |
| Azure    | ❌ Not Supported |
| Oracle   | ❌ Not Supported |

## Contributing

- Open issues/PRs for bugs, features, or module/provider additions.
- Provide reproducible steps, sample specs/module metadata, and `go test ./...` output when applicable.
- Keep diffs focused and cross-platform friendly.

_Note: running `go test ./...` here currently fails because `/Users/evalsocket/Library/Caches/go-build/...` is not writable and the Go toolchain version differs from `go 1.25.3`; these issues are environmental and unrelated to the repository._
