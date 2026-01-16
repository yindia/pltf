# Bring Your Own Terraform

This guide shows how to use existing Terraform modules and layouts with PLTF.

## Use existing modules
1) Put your modules under a custom root, for example:
```
modules/
  my_network/
  my_cluster/
```

2) Generate `module.yaml` for each module:
```
pltf module init --path ./modules/my_network --force
pltf module init --path ./modules/my_cluster --force
```

3) Reference the modules from your specs:
```yaml
modules:
  - id: network
    type: my_network
    source: custom
  - id: cluster
    type: my_cluster
    source: custom
```

4) Point PLTF to your custom root:
```
pltf generate -f env.yaml -e dev --modules ./modules
```

Notes:
- `source: custom` forces lookup in the custom root.
- Modules without `source: custom` still resolve from the embedded catalog.

## Use an existing Terraform root
PLTF expects a module list, so an existing Terraform root should be wrapped as a module:
1) Move the root code into a module folder under your custom root.
2) Run `pltf module init` to generate `module.yaml`.
3) Reference it as `type: <module_type>` in env/service.

This keeps the existing code intact and lets PLTF handle wiring and validation.

## Per-environment workspace copies
PLTF generates a separate workspace per environment:
- Environment: `.pltf/<env_name>/workspace`
- Service: `.pltf/<env_name>/<service>/workspace`

When a module input references a file path (relative to the spec), the file is copied
into the env workspace and the path is rewritten. This lets you keep per‑env
overrides isolated while reusing the same module code.

## Suggested folder layout
```
.
├─ env.yaml
├─ service.yaml
├─ modules/
│  ├─ my_network/
│  └─ my_cluster/
└─ files/
   ├─ dev/
   └─ prod/
```

Use file paths in module inputs (e.g., `./files/dev/config.json`), and PLTF will
copy them into the generated workspace for that environment.

## Enterprise migration
See `docs/features/migration-enterprise.md` for a phased approach to migrating
large Terraform monorepos into PLTF and splitting monoliths into modules.
