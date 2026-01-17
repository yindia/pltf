# pltf CLI 🚀

High-level infrastructure specs that generate standard Terraform workflows.

## Table of Contents

- [Overview](#overview)
- [Requirements](#requirements)
- [Install](#install)
- [Quickstart](#quickstart)
- [Terraform Workflows](#terraform-workflows)
- [Image & Plugin Caching](#image--plugin-caching)
- [Commands](#commands)
- [Behavior & Rules](#behavior--rules)
- [Provider Support](#provider-support)
- [Kubernetes Deployment Support](#kubernetes-deployment-support)
- [Contributing](#contributing)

## Overview

- Validate Environment, Service, and Stack YAML specs.
- Generate Terraform modules/backends/providers/secrets/outputs automatically.
- Wrap Terraform commands with auto-generation, plan/apply/destroy/output helpers, and spec graphs.
- Keep generated Terraform in your repo to avoid lock-in.

## Requirements

- Go 1.25.x
- `git` and `terraform` (>= 1.6.6) on `PATH`
- Dagger-enabled environment with host credentials mounted (`~/.aws`, `~/.docker`, etc.)

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
   providers:
     kubernetes: true
     helm: true
   modules:
     - id: base
       type: aws_base
     - id: eks
       type: aws_eks
       inputs:
         cluster_name: var.cluster_name
   ```

2. **Reference the stack from an environment** (`env.yaml`):

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
   modules:
     - id: dns
       type: aws_dns
       inputs:
         domain: var.base_domain
   ```

3. **Preview and validate**:

   ```bash
   ./pltf preview -f env.yaml -e dev
   ./pltf validate -f env.yaml -e dev
   ```

4. **Generate and apply**:

   ```bash
   ./pltf generate -f env.yaml -e dev
   ./pltf terraform plan -f env.yaml -e dev --scan
   ./pltf terraform apply -f env.yaml -e dev
   ```

5. **Optional**: Add a service (`service.yaml`) that references the environment spec and wires new modules.

## Terraform Workflows

- Every Terraform command runs inside Dagger so it shares the workspace, cached providers/plugins, and environment credentials while keeping `.pltf-plan.tfplan` artifacts for plan/apply.
- `pltf terraform plan/apply/destroy` build the Docker images from the spec; apply pushes them using the host’s registry auth while plan/destroy only build locally.
- Apply/destroy always run `--auto-approve`, `terraform plan` produces detailed summaries, and optional flags like `--scan`, `--cost`, and `--rover` extend the plan output.
- Logs include `[pltf] …` progress prefixes, and you can enable Dagger verbosity via `PLTF_DAGGER_LOG`, `PLTF_DEBUG`, or `PLTF_VERBOSE`.

## Image & Plugin Caching

- Docker builds mount a Dagger cache volume (`pltf-image-cache`) so BuildKit reuse layers between executions.
- Terraform providers/plugins are cached in `pltf-terraform-plugin-cache`, mounted at `/work/.terraform-plugin-cache` and referenced in the generated `~/.terraformrc`, so the same provider binaries are reused across projects.
- Image builds honor the spec-level `platforms` list (`["linux/amd64","linux/arm64"]` style); if unspecified, the host OS/ARCH is used for the build and push.

## Commands

| Command                       | Description                                                                                             |
|------------------------------|---------------------------------------------------------------------------------------------------------|
| `pltf generate`               | Generate Terraform from an Environment or Service spec.                                                   |
| `pltf validate`               | Validate a spec (`--scan` runs tfsec).                                                                   |
| `pltf preview`                | Preview provider/backend/vars/module wiring.                                                            |
| `pltf version`                | Show pltf, Terraform, and provider versions.                                                              |
| `pltf module init`            | Inspect a Terraform module and write `module.yaml`.                                                     |
| `pltf module list`            | List available modules (embedded or custom).                                                            |
| `pltf module get`             | Show module inputs/outputs.                                                                              |
| `pltf terraform plan`         | Generate and run `plan`. Supports `--scan`, `--cost`, `--rover`.                                          |
| `pltf terraform apply`        | Build/push images and run `apply` (auto-approved).                                                       |
| `pltf terraform destroy`      | Build images (no push) and run `destroy`.                                                                |
| `pltf terraform output`       | Display Terraform outputs (`--json` for machine-readable).                                              |
| `pltf terraform graph`        | Terraform/spec graph or terravision output (`--format`, `--outfile`).                                    |
| `pltf terraform force-unlock` | Force unlock Terraform state (requires `--lock-id`).                                                     |

## Behavior & Rules

- Stacks merge before generation; stack modules cannot be overridden by environment/service modules.
- Stack variables apply automatically; env/service `variables` only control custom logic and must not reuse stack names already defined in stacks.
- Providers are explicit via the `providers` block—there’s no implicit module-type inference.
- Auto-wiring uses output name matching from the merged module set; duplicate outputs raise errors.
- Git refs work for `metadata.ref` and `metadata.stacks` (example: `https://host/org/repo.git//path/to/spec.yaml?ref=main`).

## Provider Support

| Provider | Status        |
|----------|---------------|
| AWS      | ✅ Supported   |
| GCP      | 🔜 Coming Soon |
| Azure    | ❌ Not Supported |
| Oracle   | ❌ Not Supported |

## Kubernetes Deployment Support

| Method              | Supported |
|---------------------|-----------|
| Helm                | ✅        |
| Kustomize           | ✅        |
| Kubernetes Provider | ✅        |
| Raw Manifests       | ✅        |

## Contributing

- Open issues/PRs for bugs, features, or module/provider additions.
- Include repro steps, sample specs/module metadata, and `go test ./...` output when possible.
- Keep changes small and portable across shells/platforms.

_Note: running `go test ./...` in this workspace currently fails because `/Users/evalsocket/Library/Caches/go-build/...` is not writable and the Go toolchain is 1.25.5 while the module uses `go 1.25.3`; these are environment/tooling limitations and are unrelated to the codebase._
