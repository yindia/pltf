# pltf CLI 🚀

pltf (Platform Tools) turns your high-level infrastructure intent into ready-to-run Terraform workspaces with sensible defaults, instant validation, and inline security/cost scans. Instead of hand-editing Terraform everywhere, keep your cloud knowledge in reusable specs (`Environment`, `Service`, `Stack`) and let pltf generate the modules, providers, backends, and secrets wiring for you.

## Why teams choose pltf

- **Spec-first workflows** – describe your desired environments, services, and stacks once in YAML; pltf materializes the Terraform code, backend config, and provider mappings automatically.
- **Consistent Terraform runs** – every `pltf terraform …` command generates the workspace first, runs the host Terraform binary (no hidden runtime layers), and streamlines approvals (`plan` can include tfsec/cost/rover, `apply`/`destroy` always auto-approve).
- **Image-aware operations** – Docker images declared in specs are built once per plan/apply via Dagger, reusing the shared cache, and only `apply` pushes the tags so CI/CD remains in control.
- **Security and cost transparency** – builtin tfsec reporting and optional Infracost summaries show up alongside the Terraform logs so you don’t miss risky findings before you deploy.
- **Versioned, portable tooling** – the CLI stays Go-native, works alongside your existing git/terraform installs, and documents workflows for every team member.

## Key concepts

- **Environment** (e.g., `env.yaml`) describes cloud providers, backend/state configuration, account/region entries, and the modules/variables that belong to a physical deployment (dev, prod, etc.). Use `variables` for values that vary per environment and `secrets` to reference vaults or secret files.
- **Service** references an Environment spec and wires additional modules or metadata for a specific workload. Use `metadata.envRef` to map service-specific overrides or secrets to the environments defined upstream.
- **Stack** is a reusable collection of modules with predefined inputs/outputs, labels, and provider requirements. Environments can include one or more stacks via `metadata.stacks`, which merge before generation.
- **Variables/secrets** follow the spectype structure: `variables` is a simple map of key/value pairs available to modules, while `secrets` reference secret providers, file paths, or template-based values that get injected into Terraform without checking them into Git.

## Table of Contents

- [Install](#install)
- [Quickstart](#quickstart)
- [Terraform Workflows](#terraform-workflows)
- [Specs](#specs)
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

1. **Define reusable stacks** (`stack.yaml`):

   ```yaml
   apiVersion: platform.io/v1
   kind: Stack
   metadata:
     name: k8s-cluster
     labels:
       workload: infra
   modules:
     - id: base
       type: aws_base
     - id: eks
       type: aws_eks
       inputs:
         cluster_name: var.cluster_name
   ```

2. **Declare the base environment** (`env.yaml`):

   ```yaml
   apiVersion: platform.io/v1
   kind: Environment
   metadata:
     name: example-aws
     org: pltf
     provider: aws
     stacks:
       - ./stack.yaml
   backend:
     type: s3
     bucket: platform-tfstate
     region: us-east-1
   environments:
     dev:
       account: "111111111111"
       region: us-east-1
       variables:
         base_domain: example.com
   variables:
     cluster_name: example-dev
   modules:
     - id: dns
       type: aws_dns
       inputs:
         domain: var.base_domain
   images:
     - name: platform-tools
       context: .
       dockerfile: Dockerfile
       tags:
         - ghcr.io/example/example-aws:${env_name}
       platforms:
         - linux/amd64
         - linux/arm64
   secrets:
     s3_creds:
       type: file
       path: ~/.aws/credentials
   ```

3. **Preview, validate, and run Terraform**:

   ```bash
   ./pltf preview -f env.yaml -e dev
   ./pltf validate -f env.yaml -e dev --scan
   ./pltf generate -f env.yaml -e dev
   ./pltf terraform plan -f env.yaml -e dev --scan --cost
   ./pltf terraform apply -f env.yaml -e dev
   ```

4. **Add services** (`service.yaml`) that reference the environment and define workloads without duplicating account-level definitions.

## Terraform Workflows

- Terraform commands execute locally in the generated workspace; the CLI simply materializes Terraform files, copies `.tfvars`, and calls the host `terraform` binary (no special plugin cache or wrappers required).
- `pltf terraform plan` and `apply` build the declared Docker images via Dagger, reusing the `pltf-image-cache`. `apply` pushes the selected tags while `plan` stops after building locally. `destroy` skips image builds entirely.
- Commands like `plan` can still include optional helpers such as `--scan` (tfsec), `--cost` (Infracost), and `--rover`. `apply`/`destroy` always append `-auto-approve` so CI/CD can run unattended.
- Every command reuses the generated workspace in `.pltf/<spec-name>/<env>/workspace`, so plan/apply operate on the same graph and Terraform state.

## Specs & Concepts

- **Environment spec** (e.g., `env.yaml`) declares the cloud provider, backend, base modules, variables, secrets, and optional stacks/images that are shared by every deployment of that account-level infrastructure (dev, prod, etc.). This is your base stack: VPCs, clusters, backend S3 buckets, etc.
- **Service spec** references an Environment via `metadata.ref` and contains the workload-specific Terraform modules, metadata, secrets, images, and variable overrides that run on top of that environment. Use the service to represent applications that reuse the base infra but introduce their own modules or overrides; `metadata.envRef` maps the service into one or more of the environment entries so you can deploy it to prod, staging, or both.
- **Stack spec** is a reusable bundle of modules, inputs, and outputs with documented provider requirements. Environments include stacks through `metadata.stacks`, and stacks are merged before generation so downstream specs cannot override them unexpectedly.
- **Modules** can be either the built-in embedded modules or your own custom Terraform modules. Drop a `module.yaml` next to a custom module, refer to it in your spec, and pltf will treat it just like an embedded module during generation.
- **Variables & secrets** are first-class at every level. Place shared values in the environment or stack, and keep workload-specific overrides/secrets inside the service spec. Secrets never land in Git: pltf injects them during generation using the `secrets` block.

### Service example

```yaml
apiVersion: platform.io/v1
kind: Service
metadata:
  name: billing
  ref: ../env.yaml
  envRef:
    dev: {}
    prod:
      variables:
        replica_count: 3
modules:
  - id: api
    type: aws_service
    inputs:
      env: var.env_name
secrets:
  db_password:
    type: file
    path: ~/.secrets/db-prod.txt
images:
  - name: billing-api
    context: ./services/billing
    tags:
      - ghcr.io/example/billing:${env_name}
```

This service reuses the base `env.yaml`, injects service-specific modules/secrets, builds its own image, and opts into both `dev` and `prod` (with extra prod-only variables).

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
