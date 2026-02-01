# Variables

Reuse specs across environments and services with minimal templating.

## Overview
Variables can be declared in your specs and overridden at runtime. They resolve into Terraform variables so you can keep a single spec and tune it per environment.

## Declare variables
Define inputs in your spec and reference them with `${var.<name>}`:
```yaml
variables:
  min_nodes: "2"
  max_nodes: "5"
modules:
  - id: eks
    type: aws_eks
    inputs:
      min_nodes: "${var.min_nodes}"
      max_nodes: "${var.max_nodes}"
```

## Override at runtime
Use repeatable `--var` flags or environment variables:
```bash
pltf terraform apply -f env.yaml -e prod --var min_nodes=3 --var max_nodes=6
# or
PLTF_VAR_min_nodes=3 PLTF_VAR_max_nodes=6 pltf generate -f env.yaml -e prod
```

## Scope rules
- Stack variables provide shared defaults and are auto‑applied.
- Environment/Service `variables` are defined at the top level and must not reuse stack variable names.

## Parent outputs
Services can use environment outputs via `${parent.<output>}`:
```yaml
public_uri: "${parent.domain}/hello"
```

## Placeholder catalog
- `${env_name}`, `${layer_name}` (intrinsics)
- `${module.<module_name>.<output_name>}`
- `${parent.<output_name>}` (service only)
- `${var.<name>}` (declared variables or CLI/env overrides)

## Notes
- Required variables without defaults must be provided via `--var` or env.
- Precedence: stack variables → spec variables → CLI `--var`.
- Secrets remain Terraform variables; non-secret values may also be mirrored into locals for wiring.
