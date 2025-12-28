pltf CLI 🚀
============


The next generation of Infrastructure-as-Code: high-level constructs, less low-level cloud yak-shaving.

What It Does
------------
- 🚧 Active development; not production-hardened. Expect breaking changes and review generated code before applying.
- ✅ Validates environment and service YAML specs.
- 🛠️ Generates Terraform (providers, backend, locals, secrets, modules, outputs).
- ▶️ Wraps Terraform commands (init/plan/apply/destroy/output) with auto-generation.
- 🕸️ Emits DOT graphs (Terraform graph or spec-only dependency graph).
- 🔓 Zero tool lock-in: generated Terraform stays in your repo and is fully yours.
- 📈 Scale anytime with many engineers without enforced opinions or workflows.
- 🔐 Security scanning via `--scan` (tfsec) and optional cost estimation via `--cost` (infracost).
- 🌎 Cloud-agnostic backends: use `s3|gcs|azurerm` regardless of provider.
- 🧩 Module catalog with custom module support (`source: custom` + `--modules`).

Getting Started
---------------
1) Build the CLI:
```bash
go build -o pltf ./main.go
```

2) Create an environment spec (`env.yaml`):
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
  bucket: platform-tfstate
  region: us-east-1
environments:
  dev:
    account: "111111111111"
    region: us-east-1
    variables:
      base_domain: dev.example.internal
modules:
  - id: base
    type: aws_base
  - id: dns
    type: aws_dns
    inputs:
      domain: var.base_domain
```

3) Preview or validate:
```bash
./pltf preview -f env.yaml -e dev
./pltf validate -f env.yaml -e dev
```

4) Generate Terraform:
```bash
./pltf generate -f env.yaml -e dev
```

5) Plan/apply with Terraform:
```bash
./pltf terraform plan -f env.yaml -e dev --scan
./pltf terraform apply -f env.yaml -e dev
```

6) Add a service spec (`service.yaml`) that references the environment:
```yaml
apiVersion: platform.io/v1
kind: Service
metadata:
  name: payments-api
  ref: ./env.yaml
  envRef:
    dev: {}
modules:
  - id: app
    type: aws_k8s_service
    inputs:
      public_uri: "/payments"
      image: "ghcr.io/acme/payments:latest"
```

7) Plan/apply a service:
```bash
./pltf terraform plan -f service.yaml -e dev
```

Install / Build
---------------
- Prereqs: Go 1.25.x, `git`, `terraform` on PATH.
- Build: `go build -o pltf ./main.go`
- Install: `go install ./...`

Command Reference
-----------------

| Command                       | Description                                                                                                                               |
| ----------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `pltf generate`               | Generates Terraform code from a spec file. Auto-detects between Environment and Service specs.                                            |
| `pltf validate`               | Validates the structure and syntax of a spec file. Can optionally run a security scan with `--scan`.                                      |
| `pltf preview`                | Shows a summary of what will be generated (provider, backend, modules, etc.) without running Terraform.                                   |
| `pltf version`                | Displays the version of `pltf`, Terraform, and key providers.                                                                             |
| `pltf module init`            | Scans a Terraform module and generates a `module.yaml` metadata file.                                                                     |
| `pltf module list`            | Lists all available modules found in the configured module sources.                                                                       |
| `pltf module get`             | Shows detailed information about a specific module, including its inputs and outputs.                                                     |
| `pltf terraform plan`         | Runs `terraform plan` on the generated code. Supports flags like `--scan` (tfsec), `--cost` (infracost), and `--rover` for visualization. |
| `pltf terraform apply`        | Runs `terraform apply` on the generated code.                                                                                             |
| `pltf terraform destroy`      | Runs `terraform destroy` on the generated code.                                                                                           |
| `pltf terraform output`       | Runs `terraform output` to display outputs from the state.                                                                                |
| `pltf terraform graph`        | Generates a dependency graph. Use `--mode spec` to create a graph from the spec file without running Terraform.                             |
| `pltf terraform force-unlock` | Forcibly unlocks a Terraform state file.                                                                                                  |

Behavior & Conventions
----------------------
- File inputs: If a module input points to an existing file (relative to the spec), it is copied into the generated output dir and the path is updated.
- Outputs: All module outputs are written to `outputs.tf`; duplicate names are module-prefixed; outputs tagged with capability `secret` are marked `sensitive = true`.
- Defaults: Respects `PLTF_DEFAULT_ENV` (or profile) for env selection; embedded modules are used unless `--modules` overrides.
- Paths: Filesystem operations use platform-safe handling; generated HCL uses forward slashes for Terraform compatibility.
- Coverage: Bundled modules are AWS-first today; contributions welcome for more services and providers.

Key Features
------------
- Spec-driven infra: environment and service YAML with templated references (`module.*`, `parent.*`, `var.*`).
- Autowiring: module inputs can be satisfied automatically via outputs in the same stack.
- Remote state wiring for services via parent outputs.
- IAM policy and IRSA trust augmentation for supported AWS modules.
- Optional plan summarization, Rover visualization, and Infracost breakdowns.

Provider Support
----------------

| Provider | Status          |
|----------|-----------------|
| AWS      | ✅ Supported     |
| GCP      | 🔜 Coming Soon   |
| Azure    | ❌ Not Supported |
| Oracle   | ❌ Not Supported |

Kubernetes Native Runtime
-------------------------

The code generated by `pltf` is designed to be native to Kubernetes, which serves as the default runtime environment. 

### Kubernetes Deployment Support

| Method     | Supported |
|------------|-----------|
| Helm       | ❌ Yes      |
| Kustomize  | ❌ Yes      |
| Kubernetes | ❌ Yes      |

Contributing
------------
- Open issues/PRs for bugs, features, or module/provider additions.
- Include repro steps, sample specs/module metadata, and tests (`go test ./...`) when possible.
- Keep changes small and platform-portable (paths, shells).
- Future work ideas:
  - Expand module catalog and provider support.
  - Harden generation/validation for production use.
  - Improve graphs and diagnostics.
