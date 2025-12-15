pltf CLI 🚀
===========

What It Does
------------
- 🚧 Active development; not production-hardened. Expect breaking changes and review generated code before applying.
- ✅ Validates environment and service YAML specs.
- 🛠️ Generates Terraform (providers, backend, locals, secrets, modules, outputs).
- ▶️ Wraps Terraform commands (init/plan/apply/destroy/output) with auto-generation.
- 🕸️ Emits DOT graphs (Terraform graph or spec-only dependency graph).

Install / Build
---------------
- Prereqs: Go 1.25.x, `git`, `terraform` on PATH.
- Build: `go build -o pltf ./main.go`
- Install: `go install ./...`

Quickstart
----------
- Validate: `pltf env validate -f env.yaml` or `pltf service validate -f service.yaml`
- Generate TF only:
  - Env: `pltf generate -f env.yaml -e dev`
  - Service: `pltf generate -f service.yaml -e dev`
  - Outputs land in `.pltf/<env_name>/env/<env>` (env) or `.pltf/<env_name>/<service>/env/<env>` (service).
- Terraform wrapper:
  - `pltf terraform plan -f env.yaml -e dev`
  - `pltf terraform apply -f service.yaml -e dev --auto-approve`
  - `pltf terraform output -f env.yaml -e dev --json`
- Graphs:
  - Terraform graph: `pltf terraform graph -f env.yaml -e dev | dot -Tpng > graph.png`
  - Spec graph (no TF): `pltf terraform graph -f service.yaml -e dev --mode spec > spec.dot`
- Other commands:
  - `pltf env validate` / `pltf service validate`
  - `pltf terraform destroy|output|force-unlock` (see `pltf terraform --help`)
  - `pltf module list|get|init` to inspect or scaffold module metadata

Repo Examples
-------------
- Env: `example/env.yaml` → `pltf generate -f example/env.yaml -e dev`
- Service: `example/service.yaml` → `pltf generate -f example/service.yaml -e dev`

Behavior & Conventions
----------------------
- File inputs: If a module input points to an existing file (relative to the spec), it is copied into the generated output dir and the path is updated.
- Outputs: All module outputs are written to `outputs.tf`; duplicate names are module-prefixed; outputs tagged with capability `secret` are marked `sensitive = true`.
- Defaults: Respects `PLTF_DEFAULT_ENV` (or profile) for env selection; embedded modules are used unless `--modules` overrides.
- Paths: Filesystem operations use platform-safe handling; generated HCL uses forward slashes for Terraform compatibility.
- Coverage: Bundled modules are AWS-first today; contributions welcome for more services and providers.

Contributing
------------
- Open issues/PRs for bugs, features, or module/provider additions.
- Include repro steps, sample specs/module metadata, and tests (`go test ./...`) when possible.
- Keep changes small and platform-portable (paths, shells).
- Future work ideas:
  - Expand module catalog and provider support.
  - Harden generation/validation for production use.
  - Improve graphs and diagnostics.
